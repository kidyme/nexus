// Package refreshmeta 定义推荐刷新元数据访问抽象。
package refreshmeta

import "context"

// Repository 定义推荐刷新元数据写入接口。
type Repository interface {
	TouchUsers(ctx context.Context, userIDs []string) error
	TouchItems(ctx context.Context, itemIDs []string) error
}
