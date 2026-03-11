// Package config 定义 control 服务配置。
package config

import (
	commonconfig "github.com/kidyme/nexus/common/config"
	"github.com/kidyme/nexus/common/registry"
)

// Config 定义 control 服务配置模型。
type Config struct {
	HTTP    HTTPConfig      `mapstructure:"http"`
	Service ServiceConfig   `mapstructure:"service"`
	ETCD    registry.Config `mapstructure:"etcd"`
}

// HTTPConfig 定义 HTTP 监听地址与对外通告地址。
type HTTPConfig struct {
	Addr          string `mapstructure:"addr"`
	AdvertiseAddr string `mapstructure:"advertiseAddr"`
}

// ServiceConfig 定义服务级运行时元信息。
type ServiceConfig struct {
	Version string `mapstructure:"version"`
}

// Load 加载 control 服务配置并填充默认值。
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
