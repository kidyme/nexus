package item

import (
	"context"
	"encoding/json"
	"testing"

	offlineconfig "github.com/kidyme/nexus/offline/config"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
	userdomain "github.com/kidyme/nexus/offline/internal/domain/user"
	"github.com/kidyme/nexus/offline/internal/recallkey"
)

type fakeSourceRepository struct {
	items  []recdomain.ItemDocument
	edges  []recdomain.UserItemFeedback
	digest string
}

type fakePreparedModelCacheRepository struct {
	neighbors map[string][]recdomain.Candidate
	digest    string
}

func (f *fakeSourceRepository) ListPopularItems(context.Context, []string, int) ([]recdomain.Candidate, error) {
	return nil, nil
}

func (f *fakeSourceRepository) ListLatestItems(context.Context, int) ([]recdomain.Candidate, error) {
	return nil, nil
}

func (f *fakeSourceRepository) ListItemDocuments(context.Context) ([]recdomain.ItemDocument, error) {
	return append([]recdomain.ItemDocument(nil), f.items...), nil
}

func (f *fakeSourceRepository) ListPositiveUserItemFeedback(context.Context, []string) ([]recdomain.UserItemFeedback, error) {
	return append([]recdomain.UserItemFeedback(nil), f.edges...), nil
}

func (f *fakeSourceRepository) GetItemToItemDigest(context.Context, []string) (string, error) {
	return f.digest, nil
}

func (f *fakePreparedModelCacheRepository) GetItemToItemNeighbors(context.Context, string) (map[string][]recdomain.Candidate, string, error) {
	if f.neighbors == nil {
		return nil, f.digest, nil
	}
	result := make(map[string][]recdomain.Candidate, len(f.neighbors))
	for key, value := range f.neighbors {
		result[key] = append([]recdomain.Candidate(nil), value...)
	}
	return result, f.digest, nil
}

func (f *fakePreparedModelCacheRepository) SaveItemToItemNeighbors(_ context.Context, _ string, neighbors map[string][]recdomain.Candidate, digest string) error {
	f.digest = digest
	f.neighbors = make(map[string][]recdomain.Candidate, len(neighbors))
	for key, value := range neighbors {
		f.neighbors[key] = append([]recdomain.Candidate(nil), value...)
	}
	return nil
}

func TestItemToItemUsersRecall(t *testing.T) {
	recaller, err := NewRecaller(recallkey.CFItemToItemRecaller("users"), &fakeSourceRepository{
		items: []recdomain.ItemDocument{
			{ItemID: "a"}, {ItemID: "b"}, {ItemID: "c"},
		},
		edges: []recdomain.UserItemFeedback{
			{UserID: "u1", ItemID: "a"},
			{UserID: "u1", ItemID: "b"},
			{UserID: "u2", ItemID: "a"},
			{UserID: "u2", ItemID: "c"},
			{UserID: "u3", ItemID: "b"},
			{UserID: "u3", ItemID: "c"},
		},
		digest: "v1",
	}, &fakePreparedModelCacheRepository{}, offlineconfig.RecommendConfig{
		PositiveFeedbackTypes: []string{"like"},
	}, offlineconfig.ItemToItemUsersRecallerConfig{
		CommonRecallerConfig: offlineconfig.CommonRecallerConfig{
			Category: recallkey.CategoryCF,
			Name:     recallkey.NameItemToItem,
			Type:     "users",
		},
		NeighborCount: 10,
	})
	if err != nil {
		t.Fatalf("new users recaller: %v", err)
	}
	if err := recaller.Prepare(context.Background()); err != nil {
		t.Fatalf("prepare users recaller: %v", err)
	}
	items, err := recaller.Recall(context.Background(), userdomain.User{UserID: "u1"}, 5)
	if err != nil {
		t.Fatalf("recall users recaller: %v", err)
	}
	if len(items) != 1 || items[0].ItemID != "c" {
		t.Fatalf("unexpected users recall %#v", items)
	}
}

func TestItemToItemTagsRecall(t *testing.T) {
	recaller, err := NewRecaller(recallkey.CFItemToItemRecaller("tags"), &fakeSourceRepository{
		items: []recdomain.ItemDocument{
			{ItemID: "a", Categories: []string{"sci-fi", "space"}},
			{ItemID: "b", Categories: []string{"sci-fi", "robot"}},
			{ItemID: "c", Categories: []string{"romance"}},
		},
		edges: []recdomain.UserItemFeedback{
			{UserID: "u1", ItemID: "a"},
		},
		digest: "v1",
	}, &fakePreparedModelCacheRepository{}, offlineconfig.RecommendConfig{
		PositiveFeedbackTypes: []string{"like"},
	}, offlineconfig.ItemToItemTagsRecallerConfig{
		CommonRecallerConfig: offlineconfig.CommonRecallerConfig{
			Category: recallkey.CategoryCF,
			Name:     recallkey.NameItemToItem,
			Type:     "tags",
		},
		NeighborCount: 10,
	})
	if err != nil {
		t.Fatalf("new tags recaller: %v", err)
	}
	if err := recaller.Prepare(context.Background()); err != nil {
		t.Fatalf("prepare tags recaller: %v", err)
	}
	items, err := recaller.Recall(context.Background(), userdomain.User{UserID: "u1"}, 5)
	if err != nil {
		t.Fatalf("recall tags recaller: %v", err)
	}
	if len(items) != 1 || items[0].ItemID != "b" {
		t.Fatalf("unexpected tags recall %#v", items)
	}
}

func TestItemToItemEmbeddingRecall(t *testing.T) {
	recaller, err := NewRecaller(recallkey.CFItemToItemRecaller("embedding"), &fakeSourceRepository{
		items: []recdomain.ItemDocument{
			{ItemID: "a", Labels: mustJSON(t, map[string]any{"embedding": []float64{1, 0}})},
			{ItemID: "b", Labels: mustJSON(t, map[string]any{"embedding": []float64{0.9, 0.1}})},
			{ItemID: "c", Labels: mustJSON(t, map[string]any{"embedding": []float64{0, 1}})},
		},
		edges: []recdomain.UserItemFeedback{
			{UserID: "u1", ItemID: "a"},
		},
		digest: "v1",
	}, &fakePreparedModelCacheRepository{}, offlineconfig.RecommendConfig{
		PositiveFeedbackTypes: []string{"like"},
	}, offlineconfig.ItemToItemEmbeddingRecallerConfig{
		CommonRecallerConfig: offlineconfig.CommonRecallerConfig{
			Category: recallkey.CategoryCF,
			Name:     recallkey.NameItemToItem,
			Type:     "embedding",
		},
		NeighborCount: 10,
		LabelPath:     "embedding",
		MinScore:      0.1,
	})
	if err != nil {
		t.Fatalf("new embedding recaller: %v", err)
	}
	if err := recaller.Prepare(context.Background()); err != nil {
		t.Fatalf("prepare embedding recaller: %v", err)
	}
	items, err := recaller.Recall(context.Background(), userdomain.User{UserID: "u1"}, 5)
	if err != nil {
		t.Fatalf("recall embedding recaller: %v", err)
	}
	if len(items) != 1 || items[0].ItemID != "b" {
		t.Fatalf("unexpected embedding recall %#v", items)
	}
}

func TestItemToItemAutoRecall(t *testing.T) {
	recaller, err := NewRecaller(recallkey.CFItemToItemRecaller("auto"), &fakeSourceRepository{
		items: []recdomain.ItemDocument{
			{ItemID: "a", Categories: []string{"sci-fi"}},
			{ItemID: "b", Categories: []string{"sci-fi"}},
			{ItemID: "c", Categories: []string{"action"}},
		},
		edges: []recdomain.UserItemFeedback{
			{UserID: "u1", ItemID: "a"},
			{UserID: "u2", ItemID: "a"},
			{UserID: "u2", ItemID: "c"},
		},
		digest: "v1",
	}, &fakePreparedModelCacheRepository{}, offlineconfig.RecommendConfig{
		PositiveFeedbackTypes: []string{"like"},
	}, offlineconfig.ItemToItemAutoRecallerConfig{
		CommonRecallerConfig: offlineconfig.CommonRecallerConfig{
			Category: recallkey.CategoryCF,
			Name:     recallkey.NameItemToItem,
			Type:     "auto",
		},
		NeighborCount: 10,
		UserWeight:    1,
		TagWeight:     1,
	})
	if err != nil {
		t.Fatalf("new auto recaller: %v", err)
	}
	if err := recaller.Prepare(context.Background()); err != nil {
		t.Fatalf("prepare auto recaller: %v", err)
	}
	items, err := recaller.Recall(context.Background(), userdomain.User{UserID: "u1"}, 5)
	if err != nil {
		t.Fatalf("recall auto recaller: %v", err)
	}
	if len(items) != 2 || items[0].ItemID != "b" || items[1].ItemID != "c" {
		t.Fatalf("unexpected auto recall %#v", items)
	}
}

func TestItemToItemPrepareUsesCachedNeighbors(t *testing.T) {
	cache := &fakePreparedModelCacheRepository{}
	source := &fakeSourceRepository{
		items: []recdomain.ItemDocument{
			{ItemID: "a"},
			{ItemID: "b"},
		},
		edges: []recdomain.UserItemFeedback{
			{UserID: "u1", ItemID: "a"},
			{UserID: "u2", ItemID: "a"},
			{UserID: "u2", ItemID: "b"},
		},
		digest: "same",
	}
	recaller, err := NewRecaller(recallkey.CFItemToItemRecaller("users"), source, cache, offlineconfig.RecommendConfig{
		PositiveFeedbackTypes: []string{"like"},
	}, offlineconfig.ItemToItemUsersRecallerConfig{
		CommonRecallerConfig: offlineconfig.CommonRecallerConfig{
			Category: recallkey.CategoryCF,
			Name:     recallkey.NameItemToItem,
			Type:     "users",
		},
		NeighborCount: 10,
	})
	if err != nil {
		t.Fatalf("new users recaller: %v", err)
	}
	if err := recaller.Prepare(context.Background()); err != nil {
		t.Fatalf("prepare users recaller: %v", err)
	}
	source.items = nil
	if err := recaller.Prepare(context.Background()); err != nil {
		t.Fatalf("prepare users recaller from cache: %v", err)
	}
	items, err := recaller.Recall(context.Background(), userdomain.User{UserID: "u1"}, 5)
	if err != nil {
		t.Fatalf("recall users recaller from cache: %v", err)
	}
	if len(items) != 1 || items[0].ItemID != "b" {
		t.Fatalf("unexpected cached recall %#v", items)
	}
}

func mustJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return data
}
