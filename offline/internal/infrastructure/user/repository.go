// Package user 提供 offline 侧用户仓储实现。
package user

import (
	"context"
	"database/sql"

	userdomain "github.com/kidyme/nexus/offline/internal/domain/user"
)

// Repository 是基于 MySQL 的用户仓储实现。
type Repository struct {
	db *sql.DB
}

// NewRepository 创建用户仓储。
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List 读取全部用户。
func (r *Repository) List(ctx context.Context) ([]userdomain.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT user_id FROM users ORDER BY user_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []userdomain.User
	for rows.Next() {
		var user userdomain.User
		if err := rows.Scan(&user.UserID); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}
