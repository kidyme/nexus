// Package config provides shared configuration helpers.
package config

import (
	"os"

	"github.com/spf13/viper"
)

// IsProd 判断当前是否生产环境。
func IsProd() bool {
	return os.Getenv("NEX_ENV") == "prod"
}

// Load 加载配置：优先从 env 覆盖，支持 YAML/TOML。
// configPath 为配置文件所在目录，configName 为文件名（不含扩展名）。
func Load(configPath, configName string) (*viper.Viper, error) {
	v := viper.New()
	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
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
