// Package user 定义 offline 侧用户模型。
package user

import "context"

// User 表示待刷新的推荐用户。
type User struct {
	UserID string
}

// Repository 定义用户读取接口。
type Repository interface {
	List(ctx context.Context) ([]User, error)
}
