// Package config provides shared configuration helpers.
package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	defaultConfigPath = "config"
	defaultConfigName = "config"
	defaultConfigType = "yaml"
)

// IsProd 判断当前是否生产环境。
func IsProd() bool {
	return os.Getenv("NEX_ENV") == "prod"
}

// LoadServiceConfig 加载当前服务模块下的 config/config.yaml。
func LoadServiceConfig() (*viper.Viper, error) {
	return Load(defaultConfigPath, defaultConfigName)
}

// Load 加载配置：优先从 env 覆盖，支持 YAML/TOML。
// configPath 为配置文件所在目录，configName 为文件名（不含扩展名）。
func Load(configPath, configName string) (*viper.Viper, error) {
	v := viper.New()
	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	v.SetConfigName(configName)
	v.SetConfigType(defaultConfigType)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 无配置文件时仅用默认 + 环境变量。
			return v, nil
		}
		return nil, err
	}
	return v, nil
}

// LoadServiceConfigInto 加载当前服务模块配置并解码到 target。
func LoadServiceConfigInto(target any) error {
	v, err := LoadServiceConfig()
	if err != nil {
		return err
	}
	return v.Unmarshal(target)
}
