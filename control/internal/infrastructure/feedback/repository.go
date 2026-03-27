// Package feedback 提供反馈仓储实现。
package feedback

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	feedbackdomain "github.com/kidyme/nexus/control/internal/domain/feedback"
)

// Repository 是基于 MySQL 的反馈仓储实现。
type Repository struct {
	db *sql.DB
}

// NewRepository 创建反馈仓储。
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建反馈。
func (r *Repository) Create(ctx context.Context, feedback feedbackdomain.Feedback) error {
	return r.CreateBatch(ctx, []feedbackdomain.Feedback{feedback})
}

// CreateBatch 批量创建反馈。
func (r *Repository) CreateBatch(ctx context.Context, feedbacks []feedbackdomain.Feedback) error {
	if len(feedbacks) == 0 {
		return nil
	}

	query := `INSERT INTO feedback (feedback_type, user_id, item_id, value, timestamp) VALUES `
	args := make([]any, 0, len(feedbacks)*5)

	for i, feedback := range feedbacks {
		if i > 0 {
			query += `, `
		}
		query += `(?, ?, ?, ?, ?)`
		args = append(args, feedback.FeedbackType, feedback.UserID, feedback.ItemID, feedback.Value, feedback.Timestamp)
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// Update 更新反馈。
func (r *Repository) Update(ctx context.Context, feedback feedbackdomain.Feedback) error {
	return r.UpdateBatch(ctx, []feedbackdomain.Feedback{feedback})
}

// UpdateBatch 批量更新反馈。
func (r *Repository) UpdateBatch(ctx context.Context, feedbacks []feedbackdomain.Feedback) error {
	if len(feedbacks) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, feedback := range feedbacks {
		_, err = tx.ExecContext(ctx,
			`UPDATE feedback SET value = ?, timestamp = ? WHERE feedback_type = ? AND user_id = ? AND item_id = ?`,
			feedback.Value,
			feedback.Timestamp,
			feedback.FeedbackType,
			feedback.UserID,
			feedback.ItemID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete 删除反馈。
func (r *Repository) Delete(ctx context.Context, feedbackType, userID, itemID string) error {
	return r.DeleteBatch(ctx, []feedbackdomain.Key{{
		FeedbackType: feedbackType,
		UserID:       userID,
		ItemID:       itemID,
	}})
}

// DeleteBatch 批量删除反馈。
func (r *Repository) DeleteBatch(ctx context.Context, keys []feedbackdomain.Key) error {
	if len(keys) == 0 {
		return nil
	}

	keys = uniqueFeedbackKeys(keys)
	query := `DELETE FROM feedback WHERE (`
	args := make([]any, 0, len(keys)*3)

	for i, key := range keys {
		if i > 0 {
			query += ` OR (`
		}
		query += `feedback_type = ? AND user_id = ? AND item_id = ?)`
		args = append(args, key.FeedbackType, key.UserID, key.ItemID)
	}
	query += `)`

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected < int64(len(keys)) {
		return feedbackdomain.ErrFeedbackNotFound
	}

	return nil
}

// Find 查询单个反馈。
func (r *Repository) Find(ctx context.Context, feedbackType, userID, itemID string) (*feedbackdomain.Feedback, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT feedback_type, user_id, item_id, value, timestamp, created_at, updated_at
		 FROM feedback WHERE feedback_type = ? AND user_id = ? AND item_id = ?`,
		feedbackType, userID, itemID,
	)

	var feedback feedbackdomain.Feedback
	if err := row.Scan(
		&feedback.FeedbackType,
		&feedback.UserID,
		&feedback.ItemID,
		&feedback.Value,
		&feedback.Timestamp,
		&feedback.CreatedAt,
		&feedback.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, feedbackdomain.ErrFeedbackNotFound
		}
		return nil, err
	}
	return &feedback, nil
}

// List 列出反馈。
func (r *Repository) List(ctx context.Context, filter feedbackdomain.Filter) ([]feedbackdomain.Feedback, error) {
	query := `SELECT feedback_type, user_id, item_id, value, timestamp, created_at, updated_at FROM feedback WHERE 1 = 1`
	args := make([]any, 0, 3)

	if strings.TrimSpace(filter.FeedbackType) != "" {
		query += ` AND feedback_type = ?`
		args = append(args, filter.FeedbackType)
	}
	if strings.TrimSpace(filter.UserID) != "" {
		query += ` AND user_id = ?`
		args = append(args, filter.UserID)
	}
	if strings.TrimSpace(filter.ItemID) != "" {
		query += ` AND item_id = ?`
		args = append(args, filter.ItemID)
	}
	query += ` ORDER BY timestamp DESC, feedback_type, user_id, item_id`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []feedbackdomain.Feedback
	for rows.Next() {
		var feedback feedbackdomain.Feedback
		if err := rows.Scan(
			&feedback.FeedbackType,
			&feedback.UserID,
			&feedback.ItemID,
			&feedback.Value,
			&feedback.Timestamp,
			&feedback.CreatedAt,
			&feedback.UpdatedAt,
		); err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, feedback)
	}
	return feedbacks, rows.Err()
}

// ListPage 分页列出反馈。
func (r *Repository) ListPage(ctx context.Context, filter feedbackdomain.Filter, page, size int) ([]feedbackdomain.Feedback, int64, error) {
	whereClause, args := buildFilterQuery(filter)

	var total int64
	countQuery := `SELECT COUNT(*) FROM feedback WHERE 1 = 1` + whereClause
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	query := `SELECT feedback_type, user_id, item_id, value, timestamp, created_at, updated_at FROM feedback WHERE 1 = 1` +
		whereClause +
		` ORDER BY timestamp DESC, feedback_type, user_id, item_id LIMIT ? OFFSET ?`
	pageArgs := append(append([]any{}, args...), size, offset)
	rows, err := r.db.QueryContext(ctx, query, pageArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var feedbacks []feedbackdomain.Feedback
	for rows.Next() {
		var feedback feedbackdomain.Feedback
		if err := rows.Scan(
			&feedback.FeedbackType,
			&feedback.UserID,
			&feedback.ItemID,
			&feedback.Value,
			&feedback.Timestamp,
			&feedback.CreatedAt,
			&feedback.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		feedbacks = append(feedbacks, feedback)
	}
	return feedbacks, total, rows.Err()
}

func uniqueFeedbackKeys(keys []feedbackdomain.Key) []feedbackdomain.Key {
	seen := make(map[string]struct{}, len(keys))
	result := make([]feedbackdomain.Key, 0, len(keys))
	for _, key := range keys {
		composite := key.FeedbackType + "\x00" + key.UserID + "\x00" + key.ItemID
		if _, ok := seen[composite]; ok {
			continue
		}
		seen[composite] = struct{}{}
		result = append(result, key)
	}
	return result
}

func buildFilterQuery(filter feedbackdomain.Filter) (string, []any) {
	query := ""
	args := make([]any, 0, 3)

	if strings.TrimSpace(filter.FeedbackType) != "" {
		query += ` AND feedback_type = ?`
		args = append(args, filter.FeedbackType)
	}
	if strings.TrimSpace(filter.UserID) != "" {
		query += ` AND user_id = ?`
		args = append(args, filter.UserID)
	}
	if strings.TrimSpace(filter.ItemID) != "" {
		query += ` AND item_id = ?`
		args = append(args, filter.ItemID)
	}

	return query, args
}
