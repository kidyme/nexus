package offline

import (
	"strings"
	"testing"

	offlineconfig "github.com/kidyme/nexus/offline/config"
)

func TestBuildRecallersRejectsUnsupportedRecaller(t *testing.T) {
	_, err := buildRecallers(&offlineconfig.Config{
		Recommend: offlineconfig.RecommendConfig{
			Recallers: []offlineconfig.RecallerConfig{
				{Name: "popular", Enabled: true, Quota: 1},
				{Name: "unknown", Enabled: true, Quota: 1},
			},
		},
	}, nil)
	if err == nil {
		t.Fatal("expected unsupported recaller error")
	}
	if !strings.Contains(err.Error(), "unsupported recaller: unknown") {
		t.Fatalf("unexpected error %v", err)
	}
}
