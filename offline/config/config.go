// Package config 定义 offline 服务配置。
package config

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	commonconfig "github.com/kidyme/nexus/common/config"
)

// Config 定义 offline 运行时配置。
type Config struct {
	Service   ServiceConfig   `mapstructure:"service"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Worker    WorkerConfig    `mapstructure:"worker"`
	Recommend RecommendConfig `mapstructure:"recommend"`
	Training  TrainingConfig  `mapstructure:"training"`
}

// ServiceConfig 定义服务级配置。
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

// RedisConfig 定义 Redis 连接配置。
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	DialTimeout  string `mapstructure:"dialTimeout"`
	ReadTimeout  string `mapstructure:"readTimeout"`
	WriteTimeout string `mapstructure:"writeTimeout"`
}

// Addr 返回 Redis 连接地址。
func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// WorkerConfig 定义 worker 调度配置。
type WorkerConfig struct {
	TickInterval string `mapstructure:"tickInterval"`
}

// RecallerConfig 定义单个召回器配置。
type RecallerConfig struct {
	Name    string `mapstructure:"name"`
	Enabled bool   `mapstructure:"enabled"`
	Quota   int    `mapstructure:"quota"`
}

// RecommendConfig 定义离线推荐配置。
type RecommendConfig struct {
	CacheSize             int              `mapstructure:"cacheSize"`
	CacheExpire           string           `mapstructure:"cacheExpire"`
	ActiveUserTTL         string           `mapstructure:"activeUserTTL"`
	PositiveFeedbackTypes []string         `mapstructure:"positiveFeedbackTypes"`
	Recallers             []RecallerConfig `mapstructure:"recallers"`
}

// Hash 返回推荐配置摘要。
func (c RecommendConfig) Hash() string {
	payload, _ := json.Marshal(c)
	sum := sha1.Sum(payload)
	return hex.EncodeToString(sum[:])
}

// EnabledRecallers 返回启用的召回器配置，顺序与配置文件一致。
func (c RecommendConfig) EnabledRecallers() []RecallerConfig {
	result := make([]RecallerConfig, 0, len(c.Recallers))
	for _, recaller := range c.Recallers {
		if !recaller.Enabled || strings.TrimSpace(recaller.Name) == "" {
			continue
		}
		result = append(result, recaller)
	}
	return result
}

// EnabledRecallerNames 返回启用召回器名称列表。
func (c RecommendConfig) EnabledRecallerNames() []string {
	recallers := c.EnabledRecallers()
	names := make([]string, 0, len(recallers))
	for _, recaller := range recallers {
		names = append(names, recaller.Name)
	}
	return names
}

// RecallersByNames 返回指定名称列表对应的启用召回器配置，顺序与 names 一致。
func (c RecommendConfig) RecallersByNames(names []string) []RecallerConfig {
	if len(names) == 0 {
		return nil
	}
	enabled := make(map[string]RecallerConfig, len(c.Recallers))
	for _, recaller := range c.EnabledRecallers() {
		enabled[recaller.Name] = recaller
	}
	result := make([]RecallerConfig, 0, len(names))
	for _, name := range names {
		recaller, ok := enabled[name]
		if !ok {
			continue
		}
		result = append(result, recaller)
	}
	return result
}

// AllocateRecallerQuotas 按整数比值为召回器分配配额。
func AllocateRecallerQuotas(total int, recallers []RecallerConfig) map[string]int {
	result := make(map[string]int)
	if total <= 0 {
		return result
	}
	if len(recallers) == 0 {
		return result
	}

	type remainder struct {
		name  string
		value int
	}

	totalWeight := 0
	for _, recaller := range recallers {
		if recaller.Quota > 0 {
			totalWeight += recaller.Quota
		}
	}
	if totalWeight <= 0 {
		return result
	}

	remaining := total
	remainders := make([]remainder, 0, len(recallers))
	for _, recaller := range recallers {
		weight := recaller.Quota
		if weight < 0 {
			weight = 0
		}
		numerator := total * weight
		quota := 0
		if totalWeight > 0 {
			quota = numerator / totalWeight
		}
		if quota > remaining {
			quota = remaining
		}
		result[recaller.Name] = quota
		remaining -= quota
		remainders = append(remainders, remainder{
			name:  recaller.Name,
			value: numerator % totalWeight,
		})
	}

	for remaining > 0 {
		bestIndex := -1
		bestValue := -1
		for i, item := range remainders {
			if item.value > bestValue {
				bestValue = item.value
				bestIndex = i
			}
		}
		if bestIndex == -1 || bestValue <= 0 {
			break
		}
		result[remainders[bestIndex].name]++
		remainders[bestIndex].value = 0
		remaining--
	}

	if remaining <= 0 {
		return result
	}

	targetNames := make([]string, 0, len(recallers))
	for _, recaller := range recallers {
		if recaller.Quota > 0 {
			targetNames = append(targetNames, recaller.Name)
		}
	}
	for i := 0; remaining > 0 && len(targetNames) > 0; i++ {
		name := targetNames[i%len(targetNames)]
		result[name]++
		remaining--
	}
	return result
}

// TrainingConfig 定义训练配置。
type TrainingConfig struct {
	Backend string `mapstructure:"backend"`
}

// MustLoad 加载 offline 配置，失败时直接 panic。
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

// Load 严格加载 offline 配置。
func Load() (*Config, error) {
	v, err := commonconfig.LoadFromCaller("configs", "config")
	if err != nil {
		return nil, err
	}
	if v.ConfigFileUsed() == "" {
		return nil, errors.New("offline config: configs/config.yaml is required")
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate 校验 offline 必填配置。
func (c *Config) Validate() error {
	if strings.TrimSpace(c.MySQL.Host) == "" {
		return errors.New("offline config: mysql.host is required")
	}
	if c.MySQL.Port <= 0 {
		return errors.New("offline config: mysql.port must be greater than 0")
	}
	if strings.TrimSpace(c.MySQL.Database) == "" {
		return errors.New("offline config: mysql.database is required")
	}
	if strings.TrimSpace(c.MySQL.Username) == "" {
		return errors.New("offline config: mysql.username is required")
	}
	if strings.TrimSpace(c.MySQL.Charset) == "" {
		return errors.New("offline config: mysql.charset is required")
	}
	if c.MySQL.MaxOpenConns <= 0 {
		return errors.New("offline config: mysql.maxOpenConns must be greater than 0")
	}
	if c.MySQL.MaxIdleConns <= 0 {
		return errors.New("offline config: mysql.maxIdleConns must be greater than 0")
	}
	if _, err := parseRequiredDuration("mysql.connMaxLifetime", c.MySQL.ConnMaxLifetime); err != nil {
		return err
	}

	if strings.TrimSpace(c.Redis.Host) == "" {
		return errors.New("offline config: redis.host is required")
	}
	if c.Redis.Port <= 0 {
		return errors.New("offline config: redis.port must be greater than 0")
	}
	if _, err := parseRequiredDuration("redis.dialTimeout", c.Redis.DialTimeout); err != nil {
		return err
	}
	if _, err := parseRequiredDuration("redis.readTimeout", c.Redis.ReadTimeout); err != nil {
		return err
	}
	if _, err := parseRequiredDuration("redis.writeTimeout", c.Redis.WriteTimeout); err != nil {
		return err
	}

	if _, err := parseRequiredDuration("worker.tickInterval", c.Worker.TickInterval); err != nil {
		return err
	}
	if c.Recommend.CacheSize <= 0 {
		return errors.New("offline config: recommend.cacheSize must be greater than 0")
	}
	if _, err := parseRequiredDuration("recommend.cacheExpire", c.Recommend.CacheExpire); err != nil {
		return err
	}
	if _, err := parseRequiredDuration("recommend.activeUserTTL", c.Recommend.ActiveUserTTL); err != nil {
		return err
	}
	if len(c.Recommend.PositiveFeedbackTypes) == 0 {
		return errors.New("offline config: recommend.positiveFeedbackTypes is required")
	}
	if len(c.Recommend.EnabledRecallers()) == 0 {
		return errors.New("offline config: at least one enabled recommend.recallers entry is required")
	}
	seenRecallers := make(map[string]struct{}, len(c.Recommend.EnabledRecallers()))
	for _, recaller := range c.Recommend.EnabledRecallers() {
		if _, ok := seenRecallers[recaller.Name]; ok {
			return fmt.Errorf("offline config: duplicate recommend.recallers name %q", recaller.Name)
		}
		seenRecallers[recaller.Name] = struct{}{}
		if recaller.Quota <= 0 {
			return fmt.Errorf("offline config: recommend.recallers[%s].quota must be greater than 0", recaller.Name)
		}
	}

	if strings.TrimSpace(c.Training.Backend) == "" {
		return errors.New("offline config: training.backend is required")
	}
	return nil
}

// MustParseDuration 解析 duration，调用方已保证配置合法。
func MustParseDuration(raw string) time.Duration {
	value, err := parseRequiredDuration("duration", raw)
	if err != nil {
		panic(err)
	}
	return value
}

func parseRequiredDuration(field, raw string) (time.Duration, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, fmt.Errorf("offline config: %s is required", field)
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("offline config: invalid %s: %w", field, err)
	}
	return value, nil
}
