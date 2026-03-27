package recall

import (
	"context"

	offlineconfig "github.com/kidyme/nexus/offline/config"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
	userdomain "github.com/kidyme/nexus/offline/internal/domain/user"
)

// LatestRecaller 实现最新物品召回。
type LatestRecaller struct {
	source    recdomain.SourceRepository
	cacheSize int
	items     []recdomain.Candidate
}

// NewLatestRecaller 创建最新召回器。
func NewLatestRecaller(source recdomain.SourceRepository, cfg offlineconfig.RecommendConfig) *LatestRecaller {
	return &LatestRecaller{
		source:    source,
		cacheSize: cfg.CacheSize,
	}
}

// Name 返回召回器名称。
func (r *LatestRecaller) Name() string {
	return RecallerLatest
}

// Prepare 预热最新候选。
func (r *LatestRecaller) Prepare(ctx context.Context) error {
	limit := r.cacheSize
	if limit <= 0 {
		limit = 100
	}
	items, err := r.source.ListLatestItems(ctx, limit)
	if err != nil {
		return err
	}
	r.items = items
	return nil
}

// Recall 返回最新候选。
func (r *LatestRecaller) Recall(_ context.Context, _ userdomain.User, limit int) ([]recdomain.Candidate, error) {
	return cloneCandidates(r.items, limit), nil
}
