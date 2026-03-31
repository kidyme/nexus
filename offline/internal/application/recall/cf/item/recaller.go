// Package item 提供 item-to-item 召回器实现。
package item

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	offlineconfig "github.com/kidyme/nexus/offline/config"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
	userdomain "github.com/kidyme/nexus/offline/internal/domain/user"
)

type strategy interface {
	build(dataset dataset, source string) (map[string][]recdomain.Candidate, error)
}

type dataset struct {
	items     map[string]recdomain.ItemDocument
	userItems map[string][]string
	itemUsers map[string][]string
}

// Recaller 实现 item-to-item 协同过滤召回。
type Recaller struct {
	name          string
	source        recdomain.SourceRepository
	cache         recdomain.PreparedModelCacheRepository
	feedbackTypes []string
	config        offlineconfig.ItemToItemRecallerConfig
	strategy      strategy
	neighbors     map[string][]recdomain.Candidate
	userItems     map[string][]string
}

// NewRecaller 创建 item-to-item 召回器。
func NewRecaller(name string, source recdomain.SourceRepository, cache recdomain.PreparedModelCacheRepository, recommendCfg offlineconfig.RecommendConfig, itemToItemCfg offlineconfig.ItemToItemRecallerConfig) (*Recaller, error) {
	var s strategy
	switch typed := itemToItemCfg.(type) {
	case offlineconfig.ItemToItemUsersRecallerConfig:
		s = usersStrategy{config: typed}
	case offlineconfig.ItemToItemTagsRecallerConfig:
		s = tagsStrategy{config: typed}
	case offlineconfig.ItemToItemEmbeddingRecallerConfig:
		s = embeddingStrategy{config: typed}
	case offlineconfig.ItemToItemAutoRecallerConfig:
		s = autoStrategy{config: typed}
	default:
		return nil, fmt.Errorf("unsupported item_to_item config type: %T", itemToItemCfg)
	}
	return &Recaller{
		name:          name,
		source:        source,
		cache:         cache,
		feedbackTypes: append([]string(nil), recommendCfg.PositiveFeedbackTypes...),
		config:        itemToItemCfg,
		strategy:      s,
	}, nil
}

// Name 返回召回器标识。
func (r *Recaller) Name() string {
	return r.name
}

// Prepare 构建本轮 item 邻居索引。
func (r *Recaller) Prepare(ctx context.Context) error {
	edges, err := r.source.ListPositiveUserItemFeedback(ctx, r.feedbackTypes)
	if err != nil {
		return err
	}
	r.userItems = buildUserItemsFromEdges(edges)

	digest, err := r.buildDigest(ctx)
	if err != nil {
		return err
	}
	cacheKey := r.cacheKey()
	if r.cache != nil {
		neighbors, cachedDigest, err := r.cache.GetItemToItemNeighbors(ctx, cacheKey)
		if err != nil {
			return err
		}
		if cachedDigest == digest && neighbors != nil {
			r.neighbors = neighbors
			return nil
		}
	}

	items, err := r.source.ListItemDocuments(ctx)
	if err != nil {
		return err
	}
	ds := buildDataset(items, edges)
	neighbors, err := r.strategy.build(ds, r.Name())
	if err != nil {
		return err
	}
	r.neighbors = neighbors
	if r.cache != nil {
		if err := r.cache.SaveItemToItemNeighbors(ctx, cacheKey, neighbors, digest); err != nil {
			return err
		}
	}
	return nil
}

// Recall 基于用户历史物品聚合相似物品。
func (r *Recaller) Recall(_ context.Context, user userdomain.User, limit int) ([]recdomain.Candidate, error) {
	history := r.userItems[user.UserID]
	if len(history) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{}, len(history))
	for _, itemID := range history {
		seen[itemID] = struct{}{}
	}
	scores := make(map[string]float64)
	for _, itemID := range history {
		for _, neighbor := range r.neighbors[itemID] {
			if _, ok := seen[neighbor.ItemID]; ok {
				continue
			}
			scores[neighbor.ItemID] += neighbor.Score
		}
	}
	return topCandidatesFromScores(scores, limit, r.Name()), nil
}

func buildDataset(items []recdomain.ItemDocument, edges []recdomain.UserItemFeedback) dataset {
	ds := dataset{
		items:     make(map[string]recdomain.ItemDocument, len(items)),
		userItems: make(map[string][]string),
		itemUsers: make(map[string][]string),
	}
	for _, item := range items {
		ds.items[item.ItemID] = item
	}

	userSeen := make(map[string]map[string]struct{})
	itemSeen := make(map[string]map[string]struct{})
	for _, edge := range edges {
		if _, ok := ds.items[edge.ItemID]; !ok {
			continue
		}
		if _, ok := userSeen[edge.UserID]; !ok {
			userSeen[edge.UserID] = make(map[string]struct{})
		}
		if _, ok := userSeen[edge.UserID][edge.ItemID]; !ok {
			userSeen[edge.UserID][edge.ItemID] = struct{}{}
			ds.userItems[edge.UserID] = append(ds.userItems[edge.UserID], edge.ItemID)
		}
		if _, ok := itemSeen[edge.ItemID]; !ok {
			itemSeen[edge.ItemID] = make(map[string]struct{})
		}
		if _, ok := itemSeen[edge.ItemID][edge.UserID]; !ok {
			itemSeen[edge.ItemID][edge.UserID] = struct{}{}
			ds.itemUsers[edge.ItemID] = append(ds.itemUsers[edge.ItemID], edge.UserID)
		}
	}

	for userID := range ds.userItems {
		sort.Strings(ds.userItems[userID])
	}
	for itemID := range ds.itemUsers {
		sort.Strings(ds.itemUsers[itemID])
	}
	return ds
}

func buildUserItemsFromEdges(edges []recdomain.UserItemFeedback) map[string][]string {
	userItems := make(map[string][]string)
	userSeen := make(map[string]map[string]struct{})
	for _, edge := range edges {
		if _, ok := userSeen[edge.UserID]; !ok {
			userSeen[edge.UserID] = make(map[string]struct{})
		}
		if _, ok := userSeen[edge.UserID][edge.ItemID]; ok {
			continue
		}
		userSeen[edge.UserID][edge.ItemID] = struct{}{}
		userItems[edge.UserID] = append(userItems[edge.UserID], edge.ItemID)
	}
	for userID := range userItems {
		sort.Strings(userItems[userID])
	}
	return userItems
}

func topCandidatesFromScores(scores map[string]float64, limit int, source string) []recdomain.Candidate {
	result := make([]recdomain.Candidate, 0, len(scores))
	for itemID, score := range scores {
		if itemID == "" || score <= 0 {
			continue
		}
		result = append(result, recdomain.Candidate{
			ItemID: itemID,
			Score:  score,
			Source: source,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].ItemID < result[j].ItemID
		}
		return result[i].Score > result[j].Score
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func (r *Recaller) cacheKey() string {
	return r.Name()
}

func (r *Recaller) buildDigest(ctx context.Context) (string, error) {
	dataDigest, err := r.source.GetItemToItemDigest(ctx, r.feedbackTypes)
	if err != nil {
		return "", err
	}
	configPayload, err := json.Marshal(r.config)
	if err != nil {
		return "", err
	}
	sum := sha1.Sum([]byte(dataDigest + ":" + string(configPayload)))
	return hex.EncodeToString(sum[:]), nil
}
