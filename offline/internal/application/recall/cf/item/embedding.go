package item

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	offlineconfig "github.com/kidyme/nexus/offline/config"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
)

type embeddingStrategy struct {
	config offlineconfig.ItemToItemEmbeddingRecallerConfig
}

func (s embeddingStrategy) build(ds dataset, source string) (map[string][]recdomain.Candidate, error) {
	type vector struct {
		itemID string
		values []float64
		norm   float64
	}

	vectors := make([]vector, 0, len(ds.items))
	for itemID, item := range ds.items {
		values, err := extractEmbedding(item.Labels, s.config.LabelPath)
		if err != nil {
			return nil, fmt.Errorf("extract embedding for item %s: %w", itemID, err)
		}
		if len(values) == 0 {
			continue
		}
		norm := vectorNorm(values)
		if norm == 0 {
			continue
		}
		vectors = append(vectors, vector{
			itemID: itemID,
			values: values,
			norm:   norm,
		})
	}

	scoreMap := make(map[string]map[string]float64, len(vectors))
	for i := 0; i < len(vectors); i++ {
		for j := i + 1; j < len(vectors); j++ {
			if len(vectors[i].values) != len(vectors[j].values) {
				continue
			}
			score := cosineSimilarity(vectors[i].values, vectors[j].values, vectors[i].norm, vectors[j].norm)
			if score < s.config.MinScore {
				continue
			}
			addPairCount(scoreMap, vectors[i].itemID, vectors[j].itemID, score)
			addPairCount(scoreMap, vectors[j].itemID, vectors[i].itemID, score)
		}
	}

	neighbors := make(map[string][]recdomain.Candidate, len(ds.items))
	for itemID := range ds.items {
		neighbors[itemID] = topCandidatesFromScores(scoreMap[itemID], s.config.NeighborCount, source)
	}
	return neighbors, nil
}

func extractEmbedding(raw json.RawMessage, path string) ([]float64, error) {
	if len(raw) == 0 || strings.TrimSpace(path) == "" {
		return nil, nil
	}
	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	values := resolvePath(payload, path)
	if len(values) == 0 {
		return nil, nil
	}
	vector, ok := values[0].([]any)
	if !ok {
		return nil, fmt.Errorf("path %q is not a numeric array", path)
	}
	result := make([]float64, 0, len(vector))
	for _, value := range vector {
		number, ok := value.(float64)
		if !ok {
			return nil, fmt.Errorf("path %q contains non-numeric value", path)
		}
		result = append(result, number)
	}
	return result, nil
}

func vectorNorm(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value * value
	}
	return math.Sqrt(sum)
}

func cosineSimilarity(left, right []float64, leftNorm, rightNorm float64) float64 {
	if len(left) != len(right) || leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	var dot float64
	for i := range left {
		dot += left[i] * right[i]
	}
	return dot / (leftNorm * rightNorm)
}
