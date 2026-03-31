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
	"github.com/kidyme/nexus/offline/internal/recallkey"
)

// Config 定义 offline 运行时配置。
type Config struct {
	Service   ServiceConfig
	MySQL     MySQLConfig
	Redis     RedisConfig
	Worker    WorkerConfig
	Recommend RecommendConfig
	Training  TrainingConfig
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

// CommonRecallerConfig 定义召回器实例的通用配置。
type CommonRecallerConfig struct {
	Category string
	Name     string
	Type     string
	Enabled  bool
	Quota    int
}

// Key 返回召回器唯一标识。
func (c CommonRecallerConfig) Key() string {
	return recallkey.Key(c.Category, c.Name, c.Type)
}

// RecallerConfig 定义公开的强类型召回器配置接口。
type RecallerConfig interface {
	Common() CommonRecallerConfig
	Key() string
}

// ItemToItemRecallerConfig 定义 item-to-item 配置接口。
type ItemToItemRecallerConfig interface {
	RecallerConfig
	itemToItemRecallerConfig()
}

// PopularRecallerConfig 定义 popular 召回器配置。
type PopularRecallerConfig struct {
	CommonRecallerConfig
}

func (c PopularRecallerConfig) Common() CommonRecallerConfig { return c.CommonRecallerConfig }
func (c PopularRecallerConfig) Key() string                  { return c.CommonRecallerConfig.Key() }

// LatestRecallerConfig 定义 latest 召回器配置。
type LatestRecallerConfig struct {
	CommonRecallerConfig
}

func (c LatestRecallerConfig) Common() CommonRecallerConfig { return c.CommonRecallerConfig }
func (c LatestRecallerConfig) Key() string                  { return c.CommonRecallerConfig.Key() }

// UserToUserRecallerConfig 定义 user_to_user 召回器配置。
type UserToUserRecallerConfig struct {
	CommonRecallerConfig
}

func (c UserToUserRecallerConfig) Common() CommonRecallerConfig { return c.CommonRecallerConfig }
func (c UserToUserRecallerConfig) Key() string                  { return c.CommonRecallerConfig.Key() }

// MFRecallerConfig 定义 mf 召回器配置。
type MFRecallerConfig struct {
	CommonRecallerConfig
}

func (c MFRecallerConfig) Common() CommonRecallerConfig { return c.CommonRecallerConfig }
func (c MFRecallerConfig) Key() string                  { return c.CommonRecallerConfig.Key() }

// ExternalRecallerConfig 定义 external 召回器配置。
type ExternalRecallerConfig struct {
	CommonRecallerConfig
}

func (c ExternalRecallerConfig) Common() CommonRecallerConfig { return c.CommonRecallerConfig }
func (c ExternalRecallerConfig) Key() string                  { return c.CommonRecallerConfig.Key() }

// ItemToItemUsersRecallerConfig 定义 users 型 item-to-item 配置。
type ItemToItemUsersRecallerConfig struct {
	CommonRecallerConfig
	NeighborCount int
	Shrinkage     float64
}

func (c ItemToItemUsersRecallerConfig) Common() CommonRecallerConfig { return c.CommonRecallerConfig }
func (c ItemToItemUsersRecallerConfig) Key() string                  { return c.CommonRecallerConfig.Key() }
func (ItemToItemUsersRecallerConfig) itemToItemRecallerConfig()      {}

// ItemToItemTagsRecallerConfig 定义 tags 型 item-to-item 配置。
type ItemToItemTagsRecallerConfig struct {
	CommonRecallerConfig
	NeighborCount int
	LabelPaths    []string
	MinCommonTags int
}

func (c ItemToItemTagsRecallerConfig) Common() CommonRecallerConfig { return c.CommonRecallerConfig }
func (c ItemToItemTagsRecallerConfig) Key() string                  { return c.CommonRecallerConfig.Key() }
func (ItemToItemTagsRecallerConfig) itemToItemRecallerConfig()      {}

// ItemToItemEmbeddingRecallerConfig 定义 embedding 型 item-to-item 配置。
type ItemToItemEmbeddingRecallerConfig struct {
	CommonRecallerConfig
	NeighborCount int
	LabelPath     string
	MinScore      float64
}

func (c ItemToItemEmbeddingRecallerConfig) Common() CommonRecallerConfig {
	return c.CommonRecallerConfig
}
func (c ItemToItemEmbeddingRecallerConfig) Key() string             { return c.CommonRecallerConfig.Key() }
func (ItemToItemEmbeddingRecallerConfig) itemToItemRecallerConfig() {}

// ItemToItemAutoRecallerConfig 定义 auto 型 item-to-item 配置。
type ItemToItemAutoRecallerConfig struct {
	CommonRecallerConfig
	NeighborCount int
	LabelPaths    []string
	MinCommonTags int
	UserWeight    float64
	TagWeight     float64
}

func (c ItemToItemAutoRecallerConfig) Common() CommonRecallerConfig { return c.CommonRecallerConfig }
func (c ItemToItemAutoRecallerConfig) Key() string                  { return c.CommonRecallerConfig.Key() }
func (ItemToItemAutoRecallerConfig) itemToItemRecallerConfig()      {}

// RecommendConfig 定义离线推荐配置。
type RecommendConfig struct {
	CacheSize             int
	CacheExpire           string
	ActiveUserTTL         string
	PositiveFeedbackTypes []string
	Recallers             []RecallerConfig
}

// TrainingConfig 定义训练配置。
type TrainingConfig struct {
	Backend string `mapstructure:"backend"`
}

type rawConfig struct {
	Service   ServiceConfig      `mapstructure:"service"`
	MySQL     MySQLConfig        `mapstructure:"mysql"`
	Redis     RedisConfig        `mapstructure:"redis"`
	Worker    WorkerConfig       `mapstructure:"worker"`
	Recommend rawRecommendConfig `mapstructure:"recommend"`
	Training  TrainingConfig     `mapstructure:"training"`
}

type rawRecommendConfig struct {
	CacheSize             int                 `mapstructure:"cacheSize"`
	CacheExpire           string              `mapstructure:"cacheExpire"`
	ActiveUserTTL         string              `mapstructure:"activeUserTTL"`
	PositiveFeedbackTypes []string            `mapstructure:"positiveFeedbackTypes"`
	Recallers             []rawRecallerConfig `mapstructure:"recallers"`
}

type rawRecallerConfig struct {
	Category      string                       `mapstructure:"category"`
	Name          string                       `mapstructure:"name"`
	Type          string                       `mapstructure:"type"`
	Enabled       bool                         `mapstructure:"enabled"`
	Quota         int                          `mapstructure:"quota"`
	NeighborCount int                          `mapstructure:"neighborCount"`
	Users         rawItemToItemUsersConfig     `mapstructure:"users"`
	Tags          rawItemToItemTagsConfig      `mapstructure:"tags"`
	Embedding     rawItemToItemEmbeddingConfig `mapstructure:"embedding"`
	Auto          rawItemToItemAutoConfig      `mapstructure:"auto"`
}

type rawItemToItemUsersConfig struct {
	Shrinkage float64 `mapstructure:"shrinkage"`
}

type rawItemToItemTagsConfig struct {
	LabelPaths    []string `mapstructure:"labelPaths"`
	MinCommonTags int      `mapstructure:"minCommonTags"`
}

type rawItemToItemEmbeddingConfig struct {
	LabelPath string  `mapstructure:"labelPath"`
	MinScore  float64 `mapstructure:"minScore"`
}

type rawItemToItemAutoConfig struct {
	UserWeight float64 `mapstructure:"userWeight"`
	TagWeight  float64 `mapstructure:"tagWeight"`
}

func (r rawConfig) toConfig() (*Config, error) {
	recallers := make([]RecallerConfig, 0, len(r.Recommend.Recallers))
	for _, rawRecaller := range r.Recommend.Recallers {
		recaller, err := rawRecaller.toConfig()
		if err != nil {
			return nil, err
		}
		recallers = append(recallers, recaller)
	}
	return &Config{
		Service: r.Service,
		MySQL:   r.MySQL,
		Redis:   r.Redis,
		Worker:  r.Worker,
		Recommend: RecommendConfig{
			CacheSize:             r.Recommend.CacheSize,
			CacheExpire:           r.Recommend.CacheExpire,
			ActiveUserTTL:         r.Recommend.ActiveUserTTL,
			PositiveFeedbackTypes: append([]string(nil), r.Recommend.PositiveFeedbackTypes...),
			Recallers:             recallers,
		},
		Training: r.Training,
	}, nil
}

func (r rawRecallerConfig) toConfig() (RecallerConfig, error) {
	common := CommonRecallerConfig{
		Category: r.Category,
		Name:     r.Name,
		Type:     r.Type,
		Enabled:  r.Enabled,
		Quota:    r.Quota,
	}
	switch common.Category {
	case recallkey.CategoryNonPersonal:
		switch common.Name {
		case recallkey.NamePopular:
			return PopularRecallerConfig{CommonRecallerConfig: common}, nil
		case recallkey.NameLatest:
			return LatestRecallerConfig{CommonRecallerConfig: common}, nil
		default:
			return nil, fmt.Errorf("offline config: unsupported non_personal recaller %q", common.Name)
		}
	case recallkey.CategoryCF:
		switch common.Name {
		case recallkey.NameItemToItem:
			switch common.Type {
			case "users":
				return ItemToItemUsersRecallerConfig{
					CommonRecallerConfig: common,
					NeighborCount:        r.NeighborCount,
					Shrinkage:            r.Users.Shrinkage,
				}, nil
			case "tags":
				return ItemToItemTagsRecallerConfig{
					CommonRecallerConfig: common,
					NeighborCount:        r.NeighborCount,
					LabelPaths:           append([]string(nil), r.Tags.LabelPaths...),
					MinCommonTags:        r.Tags.MinCommonTags,
				}, nil
			case "embedding":
				return ItemToItemEmbeddingRecallerConfig{
					CommonRecallerConfig: common,
					NeighborCount:        r.NeighborCount,
					LabelPath:            r.Embedding.LabelPath,
					MinScore:             r.Embedding.MinScore,
				}, nil
			case "auto":
				return ItemToItemAutoRecallerConfig{
					CommonRecallerConfig: common,
					NeighborCount:        r.NeighborCount,
					LabelPaths:           append([]string(nil), r.Tags.LabelPaths...),
					MinCommonTags:        r.Tags.MinCommonTags,
					UserWeight:           r.Auto.UserWeight,
					TagWeight:            r.Auto.TagWeight,
				}, nil
			default:
				return nil, fmt.Errorf("offline config: recommend.recallers[%s].type %q is unsupported", common.Key(), common.Type)
			}
		case recallkey.NameUserToUser:
			return UserToUserRecallerConfig{CommonRecallerConfig: common}, nil
		case recallkey.NameMF:
			return MFRecallerConfig{CommonRecallerConfig: common}, nil
		default:
			return nil, fmt.Errorf("offline config: unsupported cf recaller name %q", common.Name)
		}
	case recallkey.CategoryExternal:
		return ExternalRecallerConfig{CommonRecallerConfig: common}, nil
	default:
		return nil, fmt.Errorf("offline config: unsupported recaller category %q", common.Category)
	}
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
		if !recaller.Common().Enabled || strings.TrimSpace(recaller.Common().Name) == "" {
			continue
		}
		result = append(result, recaller)
	}
	return result
}

// HasEnabledRecaller 返回指定标识的召回器是否启用。
func (c RecommendConfig) HasEnabledRecaller(key string) bool {
	for _, recaller := range c.EnabledRecallers() {
		if recaller.Key() == key {
			return true
		}
	}
	return false
}

// EnabledRecallerKeys 返回启用召回器标识列表。
func (c RecommendConfig) EnabledRecallerKeys() []string {
	recallers := c.EnabledRecallers()
	keys := make([]string, 0, len(recallers))
	for _, recaller := range recallers {
		keys = append(keys, recaller.Key())
	}
	return keys
}

// RecallersByKeys 返回指定标识列表对应的启用召回器配置，顺序与 keys 一致。
func (c RecommendConfig) RecallersByKeys(keys []string) []RecallerConfig {
	if len(keys) == 0 {
		return nil
	}
	enabled := make(map[string]RecallerConfig, len(c.Recallers))
	for _, recaller := range c.EnabledRecallers() {
		enabled[recaller.Key()] = recaller
	}
	result := make([]RecallerConfig, 0, len(keys))
	for _, key := range keys {
		recaller, ok := enabled[key]
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
	if total <= 0 || len(recallers) == 0 {
		return result
	}

	type remainder struct {
		name  string
		value int
	}

	totalWeight := 0
	for _, recaller := range recallers {
		if recaller.Common().Quota > 0 {
			totalWeight += recaller.Common().Quota
		}
	}
	if totalWeight <= 0 {
		return result
	}

	remaining := total
	remainders := make([]remainder, 0, len(recallers))
	for _, recaller := range recallers {
		weight := recaller.Common().Quota
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
		result[recaller.Key()] = quota
		remaining -= quota
		remainders = append(remainders, remainder{
			name:  recaller.Key(),
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
		if recaller.Common().Quota > 0 {
			targetNames = append(targetNames, recaller.Key())
		}
	}
	for i := 0; remaining > 0 && len(targetNames) > 0; i++ {
		name := targetNames[i%len(targetNames)]
		result[name]++
		remaining--
	}
	return result
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

	raw := &rawConfig{}
	if err := v.Unmarshal(raw); err != nil {
		return nil, err
	}
	cfg, err := raw.toConfig()
	if err != nil {
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
		common := recaller.Common()
		if strings.TrimSpace(common.Category) == "" {
			return errors.New("offline config: recommend.recallers.category is required")
		}
		if strings.TrimSpace(common.Name) == "" {
			return errors.New("offline config: recommend.recallers.name is required")
		}
		if common.Quota <= 0 {
			return fmt.Errorf("offline config: recommend.recallers[%s].quota must be greater than 0", recaller.Key())
		}
		if _, ok := seenRecallers[recaller.Key()]; ok {
			return fmt.Errorf("offline config: duplicate recommend.recallers key %q", recaller.Key())
		}
		seenRecallers[recaller.Key()] = struct{}{}
		if err := validateRecallerConfig(recaller); err != nil {
			return err
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

func validateRecallerConfig(recaller RecallerConfig) error {
	common := recaller.Common()
	switch typed := recaller.(type) {
	case PopularRecallerConfig:
		return validateNonPersonalRecaller(common, recallkey.NamePopular)
	case LatestRecallerConfig:
		return validateNonPersonalRecaller(common, recallkey.NameLatest)
	case UserToUserRecallerConfig:
		return validateCFRecallerBase(common, recallkey.NameUserToUser)
	case MFRecallerConfig:
		return validateCFRecallerBase(common, recallkey.NameMF)
	case ExternalRecallerConfig:
		if common.Category != recallkey.CategoryExternal {
			return fmt.Errorf("offline config: external recaller %q has invalid category %q", recaller.Key(), common.Category)
		}
		return nil
	case ItemToItemUsersRecallerConfig:
		if err := validateItemToItemCommon(common, "users"); err != nil {
			return err
		}
		if typed.NeighborCount <= 0 {
			return fmt.Errorf("offline config: recommend.recallers[%s].neighborCount must be greater than 0", recaller.Key())
		}
		if typed.Shrinkage < 0 {
			return fmt.Errorf("offline config: recommend.recallers[%s].users.shrinkage must be greater than or equal to 0", recaller.Key())
		}
		return nil
	case ItemToItemTagsRecallerConfig:
		if err := validateItemToItemCommon(common, "tags"); err != nil {
			return err
		}
		if typed.NeighborCount <= 0 {
			return fmt.Errorf("offline config: recommend.recallers[%s].neighborCount must be greater than 0", recaller.Key())
		}
		return nil
	case ItemToItemEmbeddingRecallerConfig:
		if err := validateItemToItemCommon(common, "embedding"); err != nil {
			return err
		}
		if typed.NeighborCount <= 0 {
			return fmt.Errorf("offline config: recommend.recallers[%s].neighborCount must be greater than 0", recaller.Key())
		}
		if strings.TrimSpace(typed.LabelPath) == "" {
			return fmt.Errorf("offline config: recommend.recallers[%s].embedding.labelPath is required for embedding type", recaller.Key())
		}
		return nil
	case ItemToItemAutoRecallerConfig:
		if err := validateItemToItemCommon(common, "auto"); err != nil {
			return err
		}
		if typed.NeighborCount <= 0 {
			return fmt.Errorf("offline config: recommend.recallers[%s].neighborCount must be greater than 0", recaller.Key())
		}
		if typed.UserWeight <= 0 {
			return fmt.Errorf("offline config: recommend.recallers[%s].auto.userWeight must be greater than 0", recaller.Key())
		}
		if typed.TagWeight <= 0 {
			return fmt.Errorf("offline config: recommend.recallers[%s].auto.tagWeight must be greater than 0", recaller.Key())
		}
		return nil
	default:
		return fmt.Errorf("offline config: unsupported recaller type %T", recaller)
	}
}

func validateNonPersonalRecaller(common CommonRecallerConfig, expectedName string) error {
	if common.Category != recallkey.CategoryNonPersonal {
		return fmt.Errorf("offline config: recaller %q has invalid category %q", common.Key(), common.Category)
	}
	if common.Name != expectedName {
		return fmt.Errorf("offline config: unsupported non_personal recaller %q", common.Name)
	}
	if strings.TrimSpace(common.Type) != "" {
		return fmt.Errorf("offline config: non_personal recaller %q must not set type", common.Key())
	}
	return nil
}

func validateCFRecallerBase(common CommonRecallerConfig, expectedName string) error {
	if common.Category != recallkey.CategoryCF {
		return fmt.Errorf("offline config: recaller %q has invalid category %q", common.Key(), common.Category)
	}
	if common.Name != expectedName {
		return fmt.Errorf("offline config: unsupported cf recaller name %q", common.Name)
	}
	return nil
}

func validateItemToItemCommon(common CommonRecallerConfig, expectedType string) error {
	if err := validateCFRecallerBase(common, recallkey.NameItemToItem); err != nil {
		return err
	}
	if strings.TrimSpace(common.Type) == "" {
		return errors.New("offline config: recommend.recallers item_to_item type is required")
	}
	if common.Type != expectedType {
		return fmt.Errorf("offline config: recommend.recallers[%s].type %q is unsupported", common.Key(), common.Type)
	}
	return nil
}
