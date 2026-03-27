// Package config 定义 control 服务配置。
package config

import (
	"fmt"
	commonconfig "github.com/kidyme/nexus/common/config"
	"github.com/kidyme/nexus/common/registry"
)

// Config 定义 control 服务配置模型。
type Config struct {
	HTTP      HTTPConfig      `mapstructure:"http"`
	Service   ServiceConfig   `mapstructure:"service"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Recommend RecommendConfig `mapstructure:"recommend"`
	ETCD      registry.Config `mapstructure:"etcd"`
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

// MySQLConfig 定义 MySQL 连接配置。
type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Database        string `mapstructure:"database"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Charset         string `mapstructure:"charset"`
	MaxOpenConns    int    `mapstructure:"maxOpenConns"`
	MaxIdleConns    int    `mapstructure:"maxIdleConns"`
	ConnMaxLifetime string `mapstructure:"connMaxLifetime"`
}

// RedisConfig 定义 Redis 连接配置。
type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	DialTimeout  string `mapstructure:"dialTimeout"`
	ReadTimeout  string `mapstructure:"readTimeout"`
	WriteTimeout string `mapstructure:"writeTimeout"`
}

// RecommendConfig 定义推荐语义相关配置。
type RecommendConfig struct {
	PositiveFeedbackTypes []string `mapstructure:"positiveFeedbackTypes"`
	ReadFeedbackTypes     []string `mapstructure:"readFeedbackTypes"`
}

// DSN 返回 MySQL 连接串。
func (c MySQLConfig) DSN() string {
	charset := c.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=UTC&tls=false",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		charset,
	)
}

// Load 加载 control 服务配置并填充默认值。
func Load() (*Config, error) {
	cfg := &Config{
		HTTP:    HTTPConfig{Addr: ":8080"},
		Service: ServiceConfig{Version: "dev"},
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
			Addr:         "127.0.0.1:6379",
			DB:           0,
			DialTimeout:  "3s",
			ReadTimeout:  "3s",
			WriteTimeout: "3s",
		},
		Recommend: RecommendConfig{
			PositiveFeedbackTypes: []string{"like", "star"},
			ReadFeedbackTypes:     []string{"read"},
		},
	}
	v, err := commonconfig.LoadFromCaller("configs", "config")
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
