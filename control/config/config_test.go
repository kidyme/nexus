package config

import (
	"os"
	"strings"
	"testing"
)

func TestMySQLConfigDSNIncludesDriverOptions(t *testing.T) {
	cfg := MySQLConfig{
		Host:     "127.0.0.1",
		Port:     3306,
		Database: "nexus",
		Username: "nexus",
		Password: "nexus",
	}

	dsn := cfg.DSN()

	if !strings.Contains(dsn, "tls=false") {
		t.Fatalf("expected dsn to include tls=false, got %q", dsn)
	}
	if strings.Contains(dsn, "allowPublicKeyRetrieval=true") {
		t.Fatalf("expected dsn not to include JDBC-only allowPublicKeyRetrieval, got %q", dsn)
	}
	if strings.Contains(dsn, "useSSL=false") {
		t.Fatalf("expected dsn not to include JDBC-only useSSL, got %q", dsn)
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
	if cfg.HTTP.AdvertiseAddr != "127.0.0.1:8080" {
		t.Fatalf("unexpected advertise addr %q", cfg.HTTP.AdvertiseAddr)
	}
	if cfg.Redis.Addr != "127.0.0.1:6379" {
		t.Fatalf("unexpected redis addr %q", cfg.Redis.Addr)
	}
}
