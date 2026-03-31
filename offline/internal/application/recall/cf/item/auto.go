package item

import (
	offlineconfig "github.com/kidyme/nexus/offline/config"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
)

type autoStrategy struct {
	config offlineconfig.ItemToItemAutoRecallerConfig
}

func (s autoStrategy) build(ds dataset, source string) (map[string][]recdomain.Candidate, error) {
	userNeighbors, err := (usersStrategy{
		config: offlineconfig.ItemToItemUsersRecallerConfig{
			CommonRecallerConfig: s.config.CommonRecallerConfig,
			NeighborCount:        s.config.NeighborCount,
		},
	}).build(ds, source)
	if err != nil {
		return nil, err
	}
	tagNeighbors, err := (tagsStrategy{
		config: offlineconfig.ItemToItemTagsRecallerConfig{
			CommonRecallerConfig: s.config.CommonRecallerConfig,
			NeighborCount:        s.config.NeighborCount,
			LabelPaths:           append([]string(nil), s.config.LabelPaths...),
			MinCommonTags:        s.config.MinCommonTags,
		},
	}).build(ds, source)
	if err != nil {
		return nil, err
	}

	neighbors := make(map[string][]recdomain.Candidate, len(ds.items))
	for itemID := range ds.items {
		merged := make(map[string]float64)
		for _, candidate := range userNeighbors[itemID] {
			merged[candidate.ItemID] += candidate.Score * s.config.UserWeight
		}
		for _, candidate := range tagNeighbors[itemID] {
			merged[candidate.ItemID] += candidate.Score * s.config.TagWeight
		}
		neighbors[itemID] = topCandidatesFromScores(merged, s.config.NeighborCount, source)
	}
	return neighbors, nil
}
