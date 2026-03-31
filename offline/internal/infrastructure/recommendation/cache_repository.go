package recommendation

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/kidyme/nexus/common/redisx"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
)

const (
	lastModifyUserTimeKey  = "last_modify_user_time/"
	recommendCacheKey      = "recommend_cache/"
	recommendUpdateTimeKey = "recommend_update_time/"
	recommendDigestKey     = "recommend_digest/"
	itemToItemNeighborsKey = "item_to_item_neighbors/"
	itemToItemDigestKey    = "item_to_item_digest/"
)

// CacheRepository 是基于 Redis 的推荐缓存仓储实现。
type CacheRepository struct {
	client *redisx.Client
}

// NewCacheRepository 创建推荐缓存仓储。
func NewCacheRepository(client *redisx.Client) *CacheRepository {
	return &CacheRepository{client: client}
}

// GetLastModifyUserTime 读取用户最后修改时间。
func (r *CacheRepository) GetLastModifyUserTime(ctx context.Context, userID string) (time.Time, error) {
	raw, err := r.client.Get(ctx, lastModifyUserTimeKey+userID)
	if err != nil || raw == "" {
		return time.Time{}, err
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(value).UTC(), nil
}

// GetMeta 读取推荐元数据。
func (r *CacheRepository) GetMeta(ctx context.Context, userID string) (recdomain.Meta, error) {
	var meta recdomain.Meta

	rawDigest, err := r.client.Get(ctx, recommendDigestKey+userID)
	if err != nil {
		return meta, err
	}
	rawTime, err := r.client.Get(ctx, recommendUpdateTimeKey+userID)
	if err != nil {
		return meta, err
	}

	meta.Digest = rawDigest
	if rawTime == "" {
		return meta, nil
	}
	value, err := strconv.ParseInt(rawTime, 10, 64)
	if err != nil {
		return meta, err
	}
	meta.UpdateTime = time.UnixMilli(value).UTC()
	return meta, nil
}

// Save 写入推荐结果与元数据。
func (r *CacheRepository) Save(ctx context.Context, userID string, items []recdomain.Candidate, meta recdomain.Meta) error {
	payload, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.client.SetMany(ctx, map[string]string{
		recommendCacheKey + userID:      string(payload),
		recommendUpdateTimeKey + userID: strconv.FormatInt(meta.UpdateTime.UTC().UnixMilli(), 10),
		recommendDigestKey + userID:     meta.Digest,
	})
}

// GetItemToItemNeighbors 读取 item-to-item 预计算邻居。
func (r *CacheRepository) GetItemToItemNeighbors(ctx context.Context, key string) (map[string][]recdomain.Candidate, string, error) {
	digest, err := r.client.Get(ctx, itemToItemDigestKey+key)
	if err != nil {
		return nil, "", err
	}
	payload, err := r.client.Get(ctx, itemToItemNeighborsKey+key)
	if err != nil {
		return nil, "", err
	}
	if payload == "" {
		return nil, digest, nil
	}
	var neighbors map[string][]recdomain.Candidate
	if err := json.Unmarshal([]byte(payload), &neighbors); err != nil {
		return nil, "", err
	}
	return neighbors, digest, nil
}

// SaveItemToItemNeighbors 写入 item-to-item 预计算邻居。
func (r *CacheRepository) SaveItemToItemNeighbors(ctx context.Context, key string, neighbors map[string][]recdomain.Candidate, digest string) error {
	payload, err := json.Marshal(neighbors)
	if err != nil {
		return err
	}
	return r.client.SetMany(ctx, map[string]string{
		itemToItemNeighborsKey + key: string(payload),
		itemToItemDigestKey + key:    digest,
	})
}
