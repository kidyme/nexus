package item

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	offlineconfig "github.com/kidyme/nexus/offline/config"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
)

type tagsStrategy struct {
	config offlineconfig.ItemToItemTagsRecallerConfig
}

func (s tagsStrategy) build(ds dataset, source string) (map[string][]recdomain.Candidate, error) {
	itemTags := make(map[string][]string, len(ds.items))
	tagIndex := make(map[string][]string)
	for itemID, item := range ds.items {
		tags, err := extractItemTags(item, s.config.LabelPaths)
		if err != nil {
			return nil, fmt.Errorf("extract tags for item %s: %w", itemID, err)
		}
		if len(tags) == 0 {
			continue
		}
		itemTags[itemID] = tags
		for _, tag := range tags {
			tagIndex[tag] = append(tagIndex[tag], itemID)
		}
	}

	commonCounts := make(map[string]map[string]int)
	for _, itemIDs := range tagIndex {
		slices.Sort(itemIDs)
		for i := 0; i < len(itemIDs); i++ {
			for j := i + 1; j < len(itemIDs); j++ {
				a, b := itemIDs[i], itemIDs[j]
				addCommonTag(commonCounts, a, b)
				addCommonTag(commonCounts, b, a)
			}
		}
	}

	neighbors := make(map[string][]recdomain.Candidate, len(ds.items))
	for itemID := range ds.items {
		scores := make(map[string]float64)
		for otherID, common := range commonCounts[itemID] {
			if s.config.MinCommonTags > 0 && common < s.config.MinCommonTags {
				continue
			}
			union := len(itemTags[itemID]) + len(itemTags[otherID]) - common
			if union <= 0 {
				continue
			}
			score := float64(common) / float64(union)
			if score > 0 {
				scores[otherID] = score
			}
		}
		neighbors[itemID] = topCandidatesFromScores(scores, s.config.NeighborCount, source)
	}
	return neighbors, nil
}

func extractItemTags(item recdomain.ItemDocument, labelPaths []string) ([]string, error) {
	tagSet := make(map[string]struct{})
	for _, category := range item.Categories {
		if normalized := normalizeTag(category); normalized != "" {
			tagSet[normalized] = struct{}{}
		}
	}
	if len(item.Labels) == 0 {
		return setToSortedSlice(tagSet), nil
	}
	var payload any
	if err := json.Unmarshal(item.Labels, &payload); err != nil {
		return nil, err
	}
	for _, path := range labelPaths {
		for _, value := range resolvePath(payload, path) {
			collectTagValues(value, tagSet)
		}
	}
	return setToSortedSlice(tagSet), nil
}

func resolvePath(value any, path string) []any {
	if strings.TrimSpace(path) == "" {
		return []any{value}
	}
	parts := strings.Split(path, ".")
	current := []any{value}
	for _, part := range parts {
		next := make([]any, 0)
		for _, item := range current {
			switch typed := item.(type) {
			case map[string]any:
				if child, ok := typed[part]; ok {
					next = append(next, child)
				}
			case []any:
				for _, elem := range typed {
					if childMap, ok := elem.(map[string]any); ok {
						if child, exists := childMap[part]; exists {
							next = append(next, child)
						}
					}
				}
			}
		}
		current = next
	}
	return current
}

func collectTagValues(value any, tagSet map[string]struct{}) {
	switch typed := value.(type) {
	case string:
		if normalized := normalizeTag(typed); normalized != "" {
			tagSet[normalized] = struct{}{}
		}
	case float64:
		tagSet[strconv.FormatFloat(typed, 'f', -1, 64)] = struct{}{}
	case bool:
		tagSet[strconv.FormatBool(typed)] = struct{}{}
	case []any:
		for _, elem := range typed {
			collectTagValues(elem, tagSet)
		}
	}
}

func normalizeTag(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func setToSortedSlice(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	slices.Sort(result)
	return result
}

func addCommonTag(commonCounts map[string]map[string]int, itemID, otherID string) {
	if _, ok := commonCounts[itemID]; !ok {
		commonCounts[itemID] = make(map[string]int)
	}
	commonCounts[itemID][otherID]++
}
