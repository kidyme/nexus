package config

import (
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
