package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsProd(t *testing.T) {
	t.Setenv("NEX_ENV", "prod")

	if !IsProd() {
		t.Fatal("expected IsProd to return true")
	}
}

func TestLoadReadsConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := []byte("http:\n  addr: \":8080\"\n")

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	v, err := Load(dir, "config")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got := v.GetString("http.addr"); got != ":8080" {
		t.Fatalf("expected http.addr to be :8080, got %q", got)
	}
}

func TestLoadWithoutConfigFileReturnsEmptyViper(t *testing.T) {
	dir := t.TempDir()

	v, err := Load(dir, "config")
	if err != nil {
		t.Fatalf("load config without file: %v", err)
	}

	if v == nil {
		t.Fatal("expected non-nil viper")
	}
}
