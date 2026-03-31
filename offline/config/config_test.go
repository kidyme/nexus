package config

import (
	"os"
	"strings"
	"testing"

	"github.com/kidyme/nexus/offline/internal/recallkey"
)

func TestConfigValidateRejectsMissingRequiredFields(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validate error")
	}
}

func TestConfigValidateAcceptsCompleteConfig(t *testing.T) {
	cfg := &Config{
		MySQL: MySQLConfig{
			Host:            "127.0.0.1",
			Port:            3306,
			Database:        "nexus",
			Username:        "nexus",
			Password:        "nexus",
			Charset:         "utf8mb4",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: "30m",
		},
		Redis: RedisConfig{
			Host:         "127.0.0.1",
			Port:         6379,
			DB:           0,
			DialTimeout:  "3s",
			ReadTimeout:  "3s",
			WriteTimeout: "3s",
		},
		Worker: WorkerConfig{
			TickInterval: "1m",
		},
		Recommend: RecommendConfig{
			CacheSize:             100,
			CacheExpire:           "24h",
			ActiveUserTTL:         "720h",
			PositiveFeedbackTypes: []string{"like", "star"},
			Recallers: []RecallerConfig{
				PopularRecallerConfig{CommonRecallerConfig: CommonRecallerConfig{
					Category: recallkey.CategoryNonPersonal,
					Name:     recallkey.NamePopular,
					Enabled:  true,
					Quota:    1,
				}},
			},
		},
		Training: TrainingConfig{
			Backend: "cpu",
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate config: %v", err)
	}
	if len(cfg.Recommend.Recallers) != 1 {
		t.Fatalf("expected 1 recaller, got %d", len(cfg.Recommend.Recallers))
	}
	if _, ok := cfg.Recommend.Recallers[0].(PopularRecallerConfig); !ok {
		t.Fatalf("unexpected recaller type %T", cfg.Recommend.Recallers[0])
	}
}

func TestConfigValidateRejectsDuplicateRecallerNames(t *testing.T) {
	cfg := &Config{
		MySQL: MySQLConfig{
			Host:            "127.0.0.1",
			Port:            3306,
			Database:        "nexus",
			Username:        "nexus",
			Password:        "nexus",
			Charset:         "utf8mb4",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: "30m",
		},
		Redis: RedisConfig{
			Host:         "127.0.0.1",
			Port:         6379,
			DB:           0,
			DialTimeout:  "3s",
			ReadTimeout:  "3s",
			WriteTimeout: "3s",
		},
		Worker: WorkerConfig{
			TickInterval: "1m",
		},
		Recommend: RecommendConfig{
			CacheSize:             100,
			CacheExpire:           "24h",
			ActiveUserTTL:         "720h",
			PositiveFeedbackTypes: []string{"like", "star"},
			Recallers: []RecallerConfig{
				PopularRecallerConfig{CommonRecallerConfig: CommonRecallerConfig{
					Category: recallkey.CategoryNonPersonal,
					Name:     recallkey.NamePopular,
					Enabled:  true,
					Quota:    1,
				}},
				PopularRecallerConfig{CommonRecallerConfig: CommonRecallerConfig{
					Category: recallkey.CategoryNonPersonal,
					Name:     recallkey.NamePopular,
					Enabled:  true,
					Quota:    2,
				}},
			},
		},
		Training: TrainingConfig{
			Backend: "cpu",
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected duplicate recaller validation error")
	}
	if !strings.Contains(err.Error(), `duplicate recommend.recallers key "non_personal/popular"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadFindsConfigFromAnyWorkingDirectory(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config from arbitrary cwd: %v", err)
	}
	if cfg.Redis.Host != "127.0.0.1" || cfg.Redis.Port != 6379 {
		t.Fatalf("unexpected redis config %#v", cfg.Redis)
	}
	if cfg.Worker.TickInterval != "1m" {
		t.Fatalf("unexpected tick interval %q", cfg.Worker.TickInterval)
	}
}
