// Package config defines control service configuration.
package config

import (
	commonconfig "github.com/kidyme/nexus/common/config"
	"github.com/kidyme/nexus/common/registry"
)

// Config defines the control service configuration model.
type Config struct {
	HTTP    HTTPConfig      `mapstructure:"http"`
	Service ServiceConfig   `mapstructure:"service"`
	ETCD    registry.Config `mapstructure:"etcd"`
}

// HTTPConfig defines HTTP listener and advertise addresses.
type HTTPConfig struct {
	Addr          string `mapstructure:"addr"`
	AdvertiseAddr string `mapstructure:"advertiseAddr"`
}

// ServiceConfig defines service-level runtime metadata.
type ServiceConfig struct {
	Version string `mapstructure:"version"`
}

// Load loads the control service configuration with defaults.
func Load() (*Config, error) {
	cfg := &Config{
		HTTP:    HTTPConfig{Addr: ":8080"},
		Service: ServiceConfig{Version: "dev"},
	}
	v, err := commonconfig.Load("configs", "config")
	if err != nil {
		return nil, err
	}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8080"
	}
	return cfg, nil
}
