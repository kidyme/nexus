// Package recommendation 提供召回源数据访问实现。
package recommendation

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
	"github.com/kidyme/nexus/offline/internal/recallkey"
)

// SourceRepository 是基于 MySQL 的召回源数据仓储实现。
type SourceRepository struct {
	db *sql.DB
}

// NewSourceRepository 创建召回源数据仓储。
func NewSourceRepository(db *sql.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

// ListPopularItems 返回热门物品。
func (r *SourceRepository) ListPopularItems(ctx context.Context, feedbackTypes []string, limit int) ([]recdomain.Candidate, error) {
	query := `SELECT item_id, COUNT(*) AS score FROM feedback`
	args := make([]any, 0, len(feedbackTypes)+1)
	if len(feedbackTypes) > 0 {
		query += ` WHERE feedback_type IN (` + placeholders(len(feedbackTypes)) + `)`
		for _, feedbackType := range feedbackTypes {
			args = append(args, feedbackType)
		}
	}
	query += ` GROUP BY item_id ORDER BY score DESC, item_id LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]recdomain.Candidate, 0, limit)
	for rows.Next() {
		var candidate recdomain.Candidate
		if err := rows.Scan(&candidate.ItemID, &candidate.Score); err != nil {
			return nil, err
		}
		candidate.Source = recallkey.RecallerPopular
		result = append(result, candidate)
	}
	return result, rows.Err()
}

// ListLatestItems 返回最新物品。
func (r *SourceRepository) ListLatestItems(ctx context.Context, limit int) ([]recdomain.Candidate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT item_id, timestamp FROM items WHERE is_hidden = FALSE ORDER BY timestamp DESC, item_id LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]recdomain.Candidate, 0, limit)
	for rows.Next() {
		var candidate recdomain.Candidate
		var itemTime sql.NullTime
		if err := rows.Scan(&candidate.ItemID, &itemTime); err != nil {
			return nil, err
		}
		if itemTime.Valid {
			candidate.Score = float64(itemTime.Time.UTC().Unix())
		}
		candidate.Source = recallkey.RecallerLatest
		result = append(result, candidate)
	}
	return result, rows.Err()
}

// ListItemDocuments 返回可见物品的内容快照。
func (r *SourceRepository) ListItemDocuments(ctx context.Context) ([]recdomain.ItemDocument, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT item_id, categories, labels, timestamp FROM items WHERE is_hidden = FALSE ORDER BY item_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]recdomain.ItemDocument, 0)
	for rows.Next() {
		var item recdomain.ItemDocument
		var categoriesRaw sql.NullString
		var labelsRaw sql.NullString
		var itemTime sql.NullTime
		if err := rows.Scan(&item.ItemID, &categoriesRaw, &labelsRaw, &itemTime); err != nil {
			return nil, err
		}
		if categoriesRaw.Valid && strings.TrimSpace(categoriesRaw.String) != "" {
			if err := json.Unmarshal([]byte(categoriesRaw.String), &item.Categories); err != nil {
				return nil, err
			}
		}
		if labelsRaw.Valid && strings.TrimSpace(labelsRaw.String) != "" {
			item.Labels = json.RawMessage(labelsRaw.String)
		}
		if itemTime.Valid {
			item.Timestamp = itemTime.Time.UTC()
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// ListPositiveUserItemFeedback 返回正反馈 user-item 关联。
func (r *SourceRepository) ListPositiveUserItemFeedback(ctx context.Context, feedbackTypes []string) ([]recdomain.UserItemFeedback, error) {
	query := `
SELECT DISTINCT f.user_id, f.item_id
FROM feedback f
INNER JOIN items i ON i.item_id = f.item_id
WHERE i.is_hidden = FALSE`
	args := make([]any, 0, len(feedbackTypes))
	if len(feedbackTypes) > 0 {
		query += ` AND f.feedback_type IN (` + placeholders(len(feedbackTypes)) + `)`
		for _, feedbackType := range feedbackTypes {
			args = append(args, feedbackType)
		}
	}
	query += ` ORDER BY f.user_id, f.item_id`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]recdomain.UserItemFeedback, 0)
	for rows.Next() {
		var edge recdomain.UserItemFeedback
		if err := rows.Scan(&edge.UserID, &edge.ItemID); err != nil {
			return nil, err
		}
		result = append(result, edge)
	}
	return result, rows.Err()
}

// GetItemToItemDigest 返回 item-to-item 数据摘要，用于判断是否需要重建邻居。
func (r *SourceRepository) GetItemToItemDigest(ctx context.Context, feedbackTypes []string) (string, error) {
	itemCount, itemUpdatedAt, err := r.aggregateItemsDigest(ctx)
	if err != nil {
		return "", err
	}
	feedbackCount, feedbackUpdatedAt, err := r.aggregateFeedbackDigest(ctx, feedbackTypes)
	if err != nil {
		return "", err
	}
	sum := sha1.Sum([]byte(fmt.Sprintf("items:%d:%s|feedback:%d:%s",
		itemCount,
		itemUpdatedAt,
		feedbackCount,
		feedbackUpdatedAt,
	)))
	return hex.EncodeToString(sum[:]), nil
}

func (r *SourceRepository) aggregateItemsDigest(ctx context.Context) (int64, string, error) {
	var count int64
	var updatedAt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*), MAX(updated_at) FROM items WHERE is_hidden = FALSE`,
	).Scan(&count, &updatedAt)
	if err != nil {
		return 0, "", err
	}
	return count, formatDigestTime(updatedAt), nil
}

func (r *SourceRepository) aggregateFeedbackDigest(ctx context.Context, feedbackTypes []string) (int64, string, error) {
	query := `
SELECT COUNT(*), MAX(f.updated_at)
FROM feedback f
INNER JOIN items i ON i.item_id = f.item_id
WHERE i.is_hidden = FALSE`
	args := make([]any, 0, len(feedbackTypes))
	if len(feedbackTypes) > 0 {
		query += ` AND f.feedback_type IN (` + placeholders(len(feedbackTypes)) + `)`
		for _, feedbackType := range feedbackTypes {
			args = append(args, feedbackType)
		}
	}
	var count int64
	var updatedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count, &updatedAt); err != nil {
		return 0, "", err
	}
	return count, formatDigestTime(updatedAt), nil
}

func formatDigestTime(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339Nano)
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = "?"
	}
	return strings.Join(parts, ", ")
}
