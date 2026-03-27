// Package recommendation 提供离线推荐应用服务。
package recommendation

import (
	"context"
	"fmt"
	"time"

	offlineconfig "github.com/kidyme/nexus/offline/config"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
	userdomain "github.com/kidyme/nexus/offline/internal/domain/user"
)

// Service 编排用户刷新与推荐写缓存。
type Service struct {
	users     userdomain.Repository
	cache     recdomain.CacheRepository
	recallers []recdomain.Recaller
	config    offlineconfig.RecommendConfig
}

// NewService 创建离线推荐应用服务。
func NewService(
	users userdomain.Repository,
	cache recdomain.CacheRepository,
	recallers []recdomain.Recaller,
	config offlineconfig.RecommendConfig,
) *Service {
	return &Service{
		users:     users,
		cache:     cache,
		recallers: recallers,
		config:    config,
	}
}

// RefreshAll 扫描所有用户并按需刷新缓存。
func (s *Service) RefreshAll(ctx context.Context) error {
	for _, recaller := range s.recallers {
		if err := recaller.Prepare(ctx); err != nil {
			return fmt.Errorf("prepare recaller %s: %w", recaller.Name(), err)
		}
	}

	users, err := s.users.List(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	cacheExpire := offlineconfig.MustParseDuration(s.config.CacheExpire)
	activeUserTTL := offlineconfig.MustParseDuration(s.config.ActiveUserTTL)
	digest := s.config.Hash()

	for _, user := range users {
		modifyTime, err := s.cache.GetLastModifyUserTime(ctx, user.UserID)
		if err != nil {
			return fmt.Errorf("load user modify time %s: %w", user.UserID, err)
		}
		if !isActiveUser(now, modifyTime, activeUserTTL) {
			continue
		}

		meta, err := s.cache.GetMeta(ctx, user.UserID)
		if err != nil {
			return fmt.Errorf("load recommend meta %s: %w", user.UserID, err)
		}
		if !needRefresh(now, meta, modifyTime, digest, cacheExpire) {
			continue
		}

		items, err := s.recall(ctx, user, now)
		if err != nil {
			return fmt.Errorf("recall user %s: %w", user.UserID, err)
		}
		if err := s.cache.Save(ctx, user.UserID, items, recdomain.Meta{
			UpdateTime: now,
			Digest:     digest,
		}); err != nil {
			return fmt.Errorf("save recommend cache %s: %w", user.UserID, err)
		}
	}
	return nil
}

func (s *Service) recall(ctx context.Context, user userdomain.User, now time.Time) ([]recdomain.Candidate, error) {
	if len(s.recallers) == 0 {
		return nil, nil
	}
	limit := s.config.CacheSize
	if limit <= 0 {
		limit = 100
	}
	configuredRecallers := s.config.RecallersByNames(recallerNames(s.recallers))

	seen := make(map[string]struct{}, limit)
	result := make([]recdomain.Candidate, 0, limit)
	for i, recaller := range s.recallers {
		remaining := limit - len(result)
		if remaining <= 0 {
			return result, nil
		}

		quotas := offlineconfig.AllocateRecallerQuotas(remaining, configuredRecallers[i:])
		quota := quotas[recaller.Name()]
		if quota > remaining {
			quota = remaining
		}
		if quota <= 0 {
			continue
		}

		items, err := recaller.Recall(ctx, user, quota)
		if err != nil {
			return nil, err
		}
		added := 0
		for _, item := range items {
			if added >= quota {
				break
			}
			if item.ItemID == "" {
				continue
			}
			if _, ok := seen[item.ItemID]; ok {
				continue
			}
			seen[item.ItemID] = struct{}{}
			item.Created = now
			result = append(result, item)
			added++
			if len(result) >= limit {
				return result, nil
			}
		}
	}
	return result, nil
}

func needRefresh(now time.Time, meta recdomain.Meta, modifyTime time.Time, digest string, cacheExpire time.Duration) bool {
	if meta.UpdateTime.IsZero() {
		return true
	}
	if meta.Digest == "" || meta.Digest != digest {
		return true
	}
	if cacheExpire > 0 && meta.UpdateTime.Before(now.Add(-cacheExpire)) {
		return true
	}
	if !modifyTime.IsZero() && modifyTime.After(meta.UpdateTime) {
		return true
	}
	return false
}

func isActiveUser(now time.Time, modifyTime time.Time, ttl time.Duration) bool {
	if ttl <= 0 || modifyTime.IsZero() {
		return true
	}
	return modifyTime.After(now.Add(-ttl))
}

func recallerNames(recallers []recdomain.Recaller) []string {
	names := make([]string, 0, len(recallers))
	for _, recaller := range recallers {
		names = append(names, recaller.Name())
	}
	return names
}
