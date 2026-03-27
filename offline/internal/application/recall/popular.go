package recall

import (
	"context"

	offlineconfig "github.com/kidyme/nexus/offline/config"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
	userdomain "github.com/kidyme/nexus/offline/internal/domain/user"
)

// PopularRecaller 实现热门召回。
type PopularRecaller struct {
	source        recdomain.SourceRepository
	feedbackTypes []string
	cacheSize     int
	items         []recdomain.Candidate
}

// NewPopularRecaller 创建热门召回器。
func NewPopularRecaller(source recdomain.SourceRepository, cfg offlineconfig.RecommendConfig) *PopularRecaller {
	return &PopularRecaller{
		source:        source,
		feedbackTypes: append([]string(nil), cfg.PositiveFeedbackTypes...),
		cacheSize:     cfg.CacheSize,
	}
}

// Name 返回召回器名称。
func (r *PopularRecaller) Name() string {
	return RecallerPopular
}

// Prepare 预热本轮热门候选。
func (r *PopularRecaller) Prepare(ctx context.Context) error {
	limit := r.cacheSize
	if limit <= 0 {
		limit = 100
	}
	items, err := r.source.ListPopularItems(ctx, r.feedbackTypes, limit)
	if err != nil {
		return err
	}
	r.items = items
	return nil
}

// Recall 返回热门候选。
func (r *PopularRecaller) Recall(_ context.Context, _ userdomain.User, limit int) ([]recdomain.Candidate, error) {
	return cloneCandidates(r.items, limit), nil
}
