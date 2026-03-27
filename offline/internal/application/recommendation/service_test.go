package recommendation

import (
	"context"
	"testing"
	"time"

	offlineconfig "github.com/kidyme/nexus/offline/config"
	recallapp "github.com/kidyme/nexus/offline/internal/application/recall"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
	userdomain "github.com/kidyme/nexus/offline/internal/domain/user"
)

type fakeUserRepository struct {
	listFn func(context.Context) ([]userdomain.User, error)
}

func (f *fakeUserRepository) List(ctx context.Context) ([]userdomain.User, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx)
}

type fakeCacheRepository struct {
	modifyTime time.Time
	meta       recdomain.Meta
	saveFn     func(context.Context, string, []recdomain.Candidate, recdomain.Meta) error
}

func (f *fakeCacheRepository) GetLastModifyUserTime(context.Context, string) (time.Time, error) {
	return f.modifyTime, nil
}

func (f *fakeCacheRepository) GetMeta(context.Context, string) (recdomain.Meta, error) {
	return f.meta, nil
}

func (f *fakeCacheRepository) Save(ctx context.Context, userID string, items []recdomain.Candidate, meta recdomain.Meta) error {
	if f.saveFn == nil {
		return nil
	}
	return f.saveFn(ctx, userID, items, meta)
}

type fakeRecaller struct {
	name   string
	items  []recdomain.Candidate
	limits *[]int
}

func (f *fakeRecaller) Name() string                  { return f.name }
func (f *fakeRecaller) Prepare(context.Context) error { return nil }
func (f *fakeRecaller) Recall(_ context.Context, _ userdomain.User, limit int) ([]recdomain.Candidate, error) {
	if f.limits != nil {
		*f.limits = append(*f.limits, limit)
	}
	return append([]recdomain.Candidate(nil), f.items...), nil
}

type fakeSourceRepository struct {
	popularFn func(context.Context, []string, int) ([]recdomain.Candidate, error)
	latestFn  func(context.Context, int) ([]recdomain.Candidate, error)
}

func (f *fakeSourceRepository) ListPopularItems(ctx context.Context, feedbackTypes []string, limit int) ([]recdomain.Candidate, error) {
	if f.popularFn == nil {
		return nil, nil
	}
	return f.popularFn(ctx, feedbackTypes, limit)
}

func (f *fakeSourceRepository) ListLatestItems(ctx context.Context, limit int) ([]recdomain.Candidate, error) {
	if f.latestFn == nil {
		return nil, nil
	}
	return f.latestFn(ctx, limit)
}

func TestRefreshAllSavesWhenCacheMissing(t *testing.T) {
	service := NewService(
		&fakeUserRepository{
			listFn: func(context.Context) ([]userdomain.User, error) {
				return []userdomain.User{{UserID: "u-1"}}, nil
			},
		},
		&fakeCacheRepository{
			modifyTime: time.Now().UTC(),
			saveFn: func(_ context.Context, userID string, items []recdomain.Candidate, meta recdomain.Meta) error {
				if userID != "u-1" {
					t.Fatalf("expected user u-1, got %s", userID)
				}
				if len(items) != 1 || items[0].ItemID != "i-1" {
					t.Fatalf("unexpected items %#v", items)
				}
				if meta.Digest == "" || meta.UpdateTime.IsZero() {
					t.Fatalf("unexpected meta %#v", meta)
				}
				return nil
			},
		},
		[]recdomain.Recaller{
			&fakeRecaller{name: "popular", items: []recdomain.Candidate{{ItemID: "i-1", Score: 1, Source: "popular"}}},
		},
		offlineconfig.RecommendConfig{
			CacheSize:     10,
			CacheExpire:   "24h",
			ActiveUserTTL: "720h",
			Recallers: []offlineconfig.RecallerConfig{
				{Name: "popular", Enabled: true, Quota: 1},
			},
		},
	)

	if err := service.RefreshAll(context.Background()); err != nil {
		t.Fatalf("refresh all: %v", err)
	}
}

func TestNeedRefresh(t *testing.T) {
	now := time.Now().UTC()
	meta := recdomain.Meta{
		UpdateTime: now.Add(-time.Hour),
		Digest:     "same",
	}
	if needRefresh(now, meta, now.Add(-2*time.Hour), "same", 24*time.Hour) {
		t.Fatal("expected cache to be fresh")
	}
	if !needRefresh(now, meta, now.Add(time.Minute), "same", 24*time.Hour) {
		t.Fatal("expected newer user modify time to invalidate cache")
	}
	if !needRefresh(now, meta, now.Add(-2*time.Hour), "changed", 24*time.Hour) {
		t.Fatal("expected digest mismatch to invalidate cache")
	}
}

func TestRecallReallocatesRemainingQuota(t *testing.T) {
	var limits []int
	service := NewService(
		nil,
		nil,
		[]recdomain.Recaller{
			&fakeRecaller{
				name:   "popular",
				limits: &limits,
				items: []recdomain.Candidate{
					{ItemID: "i-1", Source: "popular"},
					{ItemID: "i-2", Source: "popular"},
				},
			},
			&fakeRecaller{
				name:   "latest",
				limits: &limits,
				items: []recdomain.Candidate{
					{ItemID: "i-3", Source: "latest"},
					{ItemID: "i-4", Source: "latest"},
				},
			},
			&fakeRecaller{
				name:   "fallback",
				limits: &limits,
				items: []recdomain.Candidate{
					{ItemID: "i-5", Source: "fallback"},
					{ItemID: "i-6", Source: "fallback"},
				},
			},
		},
		offlineconfig.RecommendConfig{
			CacheSize: 5,
			Recallers: []offlineconfig.RecallerConfig{
				{Name: "popular", Enabled: true, Quota: 2},
				{Name: "latest", Enabled: true, Quota: 1},
				{Name: "fallback", Enabled: true, Quota: 1},
			},
		},
	)

	items, err := service.recall(context.Background(), userdomain.User{UserID: "u-1"}, time.Now().UTC())
	if err != nil {
		t.Fatalf("recall with quota: %v", err)
	}
	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}
	if got := []string{items[0].ItemID, items[1].ItemID, items[2].ItemID, items[3].ItemID, items[4].ItemID}; got[0] != "i-1" || got[1] != "i-2" || got[2] != "i-3" || got[3] != "i-4" || got[4] != "i-5" {
		t.Fatalf("unexpected order %#v", got)
	}
	if len(limits) != 3 || limits[0] != 3 || limits[1] != 2 || limits[2] != 1 {
		t.Fatalf("unexpected recaller limits %#v", limits)
	}
}

func TestLatestRecallerRecall(t *testing.T) {
	recaller := recallapp.NewLatestRecaller(&fakeSourceRepository{
		latestFn: func(_ context.Context, limit int) ([]recdomain.Candidate, error) {
			if limit != 5 {
				t.Fatalf("expected limit 5, got %d", limit)
			}
			return []recdomain.Candidate{{ItemID: "i-2", Score: 3, Source: "latest"}}, nil
		},
	}, offlineconfig.RecommendConfig{
		CacheSize: 5,
	})

	if err := recaller.Prepare(context.Background()); err != nil {
		t.Fatalf("prepare latest: %v", err)
	}
	items, err := recaller.Recall(context.Background(), userdomain.User{UserID: "u-1"}, 5)
	if err != nil {
		t.Fatalf("recall latest: %v", err)
	}
	if len(items) != 1 || items[0].ItemID != "i-2" {
		t.Fatalf("unexpected items %#v", items)
	}
}
