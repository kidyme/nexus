package item

import (
	"math"

	offlineconfig "github.com/kidyme/nexus/offline/config"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
)

type usersStrategy struct {
	config offlineconfig.ItemToItemUsersRecallerConfig
}

func (s usersStrategy) build(ds dataset, source string) (map[string][]recdomain.Candidate, error) {
	pairCounts := make(map[string]map[string]float64)
	for _, items := range ds.userItems {
		if len(items) < 2 {
			continue
		}
		for i := 0; i < len(items); i++ {
			for j := i + 1; j < len(items); j++ {
				a, b := items[i], items[j]
				addPairCount(pairCounts, a, b, 1)
				addPairCount(pairCounts, b, a, 1)
			}
		}
	}

	neighbors := make(map[string][]recdomain.Candidate, len(ds.items))
	for itemID := range ds.items {
		scores := make(map[string]float64)
		for otherID, common := range pairCounts[itemID] {
			denominator := math.Sqrt(float64(len(ds.itemUsers[itemID]) * len(ds.itemUsers[otherID])))
			if denominator == 0 {
				continue
			}
			score := common / denominator
			if s.config.Shrinkage > 0 {
				score *= common / (common + s.config.Shrinkage)
			}
			if score > 0 {
				scores[otherID] = score
			}
		}
		neighbors[itemID] = topCandidatesFromScores(scores, s.config.NeighborCount, source)
	}
	return neighbors, nil
}

func addPairCount(pairCounts map[string]map[string]float64, itemID, otherID string, score float64) {
	if _, ok := pairCounts[itemID]; !ok {
		pairCounts[itemID] = make(map[string]float64)
	}
	pairCounts[itemID][otherID] += score
}
