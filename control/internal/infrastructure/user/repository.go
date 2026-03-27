// Package user 提供用户仓储实现。
package user

import (
	"context"
	"database/sql"
	"errors"

	userdomain "github.com/kidyme/nexus/control/internal/domain/user"
)

// Repository 是基于 MySQL 的用户仓储实现。
type Repository struct {
	db *sql.DB
}

// NewRepository 创建用户仓储。
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建用户。
func (r *Repository) Create(ctx context.Context, user userdomain.User) error {
	return r.CreateBatch(ctx, []userdomain.User{user})
}

// CreateBatch 批量创建用户。
func (r *Repository) CreateBatch(ctx context.Context, users []userdomain.User) error {
	if len(users) == 0 {
		return nil
	}

	query := `INSERT INTO users (user_id, labels, comment) VALUES `
	args := make([]any, 0, len(users)*3)

	for i, user := range users {
		if i > 0 {
			query += `, `
		}
		query += `(?, ?, ?)`
		args = append(args, user.UserID, nullableJSON(user.Labels), user.Comment)
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// Update 更新用户。
func (r *Repository) Update(ctx context.Context, user userdomain.User) error {
	return r.UpdateBatch(ctx, []userdomain.User{user})
}

// UpdateBatch 批量更新用户。
func (r *Repository) UpdateBatch(ctx context.Context, users []userdomain.User) error {
	if len(users) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, user := range users {
		result, err := tx.ExecContext(ctx,
			`UPDATE users SET labels = ?, comment = ? WHERE user_id = ?`,
			nullableJSON(user.Labels),
			user.Comment,
			user.UserID,
		)
		if err != nil {
			return err
		}
		if err := expectRows(result, userdomain.ErrUserNotFound); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete 删除用户。
func (r *Repository) Delete(ctx context.Context, userID string) error {
	return r.DeleteBatch(ctx, []string{userID})
}

// DeleteBatch 批量删除用户。
func (r *Repository) DeleteBatch(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	userIDs = uniqueStrings(userIDs)
	query := `DELETE FROM users WHERE user_id IN (`
	args := make([]any, 0, len(userIDs))

	for i, userID := range userIDs {
		if i > 0 {
			query += `, `
		}
		query += `?`
		args = append(args, userID)
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
	if affected < int64(len(userIDs)) {
		return userdomain.ErrUserNotFound
	}

	return nil
}

// Find 查询单个用户。
func (r *Repository) Find(ctx context.Context, userID string) (*userdomain.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT user_id, labels, comment, created_at, updated_at FROM users WHERE user_id = ?`,
		userID,
	)

	var user userdomain.User
	var labels []byte
	if err := row.Scan(&user.UserID, &labels, &user.Comment, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, userdomain.ErrUserNotFound
		}
		return nil, err
	}
	user.Labels = labels
	return &user, nil
}

// List 列出全部用户。
func (r *Repository) List(ctx context.Context) ([]userdomain.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT user_id, labels, comment, created_at, updated_at FROM users ORDER BY user_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []userdomain.User
	for rows.Next() {
		var user userdomain.User
		var labels []byte
		if err := rows.Scan(&user.UserID, &labels, &user.Comment, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		user.Labels = labels
		users = append(users, user)
	}
	return users, rows.Err()
}

// ListPage 分页列出用户。
func (r *Repository) ListPage(ctx context.Context, page, size int) ([]userdomain.User, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := r.db.QueryContext(ctx,
		`SELECT user_id, labels, comment, created_at, updated_at FROM users ORDER BY user_id LIMIT ? OFFSET ?`,
		size,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []userdomain.User
	for rows.Next() {
		var user userdomain.User
		var labels []byte
		if err := rows.Scan(&user.UserID, &labels, &user.Comment, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, err
		}
		user.Labels = labels
		users = append(users, user)
	}
	return users, total, rows.Err()
}

func nullableJSON(data []byte) any {
	if len(data) == 0 {
		return nil
	}
	return data
}

func expectRows(result sql.Result, notFound error) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return notFound
	}
	return nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
