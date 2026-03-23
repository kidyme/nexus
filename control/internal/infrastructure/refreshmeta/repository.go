// Package refreshmeta 提供推荐刷新元数据仓储实现。
package refreshmeta

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/kidyme/nexus/common/redisx"
	refreshmetadomain "github.com/kidyme/nexus/control/internal/domain/refreshmeta"
)

const (
	lastModifyUserTimeKey = "last_modify_user_time"
	lastModifyItemTimeKey = "last_modify_item_time"
)

// Repository 是基于 Redis 的刷新元数据仓储实现。
type Repository struct {
	client *redisx.Client
	now    func() time.Time
}

// NewRepository 创建刷新元数据仓储。
func NewRepository(client *redisx.Client) *Repository {
	return &Repository{
		client: client,
		now:    time.Now,
	}
}

// TouchUsers 批量更新用户修改时间。
func (r *Repository) TouchUsers(ctx context.Context, userIDs []string) error {
	return r.touch(ctx, lastModifyUserTimeKey, userIDs)
}

// TouchItems 批量更新物品修改时间。
func (r *Repository) TouchItems(ctx context.Context, itemIDs []string) error {
	return r.touch(ctx, lastModifyItemTimeKey, itemIDs)
}

func (r *Repository) touch(ctx context.Context, prefix string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	timestamp := strconv.FormatInt(r.now().UTC().UnixMilli(), 10)
	values := make(map[string]string, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		values[prefix+"/"+id] = timestamp
	}
	return r.client.SetMany(ctx, values)
}

var _ refreshmetadomain.Repository = (*Repository)(nil)
