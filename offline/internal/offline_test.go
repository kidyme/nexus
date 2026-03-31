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
				offlineconfig.PopularRecallerConfig{CommonRecallerConfig: offlineconfig.CommonRecallerConfig{
					Category: recallkey.CategoryNonPersonal,
					Name:     recallkey.NamePopular,
					Enabled:  true,
					Quota:    1,
				}},
				offlineconfig.ExternalRecallerConfig{CommonRecallerConfig: offlineconfig.CommonRecallerConfig{
					Category: "unknown",
					Name:     "mystery",
					Enabled:  true,
					Quota:    1,
				}},
			},
		},
	}, nil, nil)
	if err == nil {
		t.Fatal("expected unsupported recaller error")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unexpected error %v", err)
	}
}
