// Package user 定义用户业务模型。
package user

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrUserIDRequired = errors.New("user domain: user id is required")
var ErrUserNotFound = errors.New("user domain: user not found")

// User 表示推荐系统中的用户。
type User struct {
	UserID    string          `json:"user_id"`
	Labels    json.RawMessage `json:"labels"`
	Comment   string          `json:"comment"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Validate 校验用户数据是否合法。
func (u User) Validate() error {
	if strings.TrimSpace(u.UserID) == "" {
		return ErrUserIDRequired
	}
	return nil
}

// Repository 定义用户数据访问接口。
type Repository interface {
	Create(ctx context.Context, user User) error
	CreateBatch(ctx context.Context, users []User) error
	Update(ctx context.Context, user User) error
	UpdateBatch(ctx context.Context, users []User) error
	Delete(ctx context.Context, userID string) error
	DeleteBatch(ctx context.Context, userIDs []string) error
	Find(ctx context.Context, userID string) (*User, error)
	List(ctx context.Context) ([]User, error)
	ListPage(ctx context.Context, page, size int) ([]User, int64, error)
}
