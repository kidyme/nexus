package offline

import (
	"strings"
	"testing"

	offlineconfig "github.com/kidyme/nexus/offline/config"
	"github.com/kidyme/nexus/offline/internal/recallkey"
)

func TestBuildRecallersRejectsUnsupportedRecaller(t *testing.T) {
	_, err := buildRecallers(&offlineconfig.Config{
		Recommend: offlineconfig.RecommendConfig{
			Recallers: []offlineconfig.RecallerConfig{
				{Category: recallkey.CategoryNonPersonal, Name: recallkey.NamePopular, Enabled: true, Quota: 1},
				{Category: recallkey.CategoryCF, Name: recallkey.NameItemToItem, Enabled: true, Quota: 1},
			},
		},
	}, nil)
	if err == nil {
		t.Fatal("expected unsupported recaller error")
	}
	if !strings.Contains(err.Error(), "unsupported recaller: cf/item_to_item") {
		t.Fatalf("unexpected error %v", err)
	}
}
