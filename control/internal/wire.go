package control

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/google/wire"
	"github.com/kidyme/nexus/common/log"
	"github.com/kidyme/nexus/common/redisx"
	"github.com/kidyme/nexus/common/registry"
	commonserver "github.com/kidyme/nexus/common/server"
	controlconfig "github.com/kidyme/nexus/control/config"
)

// ProviderSet 提供 control 运行时依赖。
var ProviderSet = wire.NewSet(
	ProvideConfig,
	ProvideRegistry,
	ProvideMySQLDB,
	ProvideRedisClient,
	ProvideSelfNode,
	ProvideHeartbeatInterval,
	ProvideHTTPServer,
	NewApp,
)

// ProvideConfig 加载 control 运行时配置。
func ProvideConfig() (*controlconfig.Config, error) {
	return controlconfig.Load()
}

// ProvideRegistry 创建服务注册中心及其清理函数。
func ProvideRegistry(cfg *controlconfig.Config) (registry.Registry, func(), error) {
	nodeRegistry, err := registry.NewEtcdRegistry(cfg.ETCD)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		if closeErr := nodeRegistry.Close(); closeErr != nil {
			log.Error("关闭 registry 失败", "error", closeErr)
		}
	}
	return nodeRegistry, cleanup, nil
}

// ProvideMySQLDB 创建 MySQL 连接及其清理函数。
func ProvideMySQLDB(cfg *controlconfig.Config) (*sql.DB, func(), error) {
	db, err := sql.Open("mysql", cfg.MySQL.DSN())
	if err != nil {
		return nil, nil, err
	}
	if cfg.MySQL.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	}
	if cfg.MySQL.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	}
	if cfg.MySQL.ConnMaxLifetime != "" {
		d, err := time.ParseDuration(cfg.MySQL.ConnMaxLifetime)
		if err != nil {
			_ = db.Close()
			return nil, nil, err
		}
		db.SetConnMaxLifetime(d)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return db, func() {
		if err := db.Close(); err != nil {
			log.Error("关闭 mysql 连接失败", "error", err)
		}
	}, nil
}

// ProvideRedisClient 创建 Redis 客户端。
func ProvideRedisClient(cfg *controlconfig.Config) (*redisx.Client, error) {
	dialTimeout, err := time.ParseDuration(cfg.Redis.DialTimeout)
	if err != nil {
		return nil, err
	}
	readTimeout, err := time.ParseDuration(cfg.Redis.ReadTimeout)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := time.ParseDuration(cfg.Redis.WriteTimeout)
	if err != nil {
		return nil, err
	}

	client := redisx.NewClient(redisx.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	})
	if err := client.Ping(context.Background()); err != nil {
		return nil, err
	}
	return client, nil
}

// ProvideSelfNode 创建当前 control 节点的注册元信息。
func ProvideSelfNode(cfg *controlconfig.Config) (registry.Node, error) {
	if cfg.HTTP.AdvertiseAddr == "" {
		return registry.Node{}, fmt.Errorf("control: http.advertiseAddr is required")
	}
	if cfg.Service.Version == "" {
		return registry.Node{}, fmt.Errorf("control: service.version is required")
	}
	return registry.Node{
		NodeID:      uuid.NewString()[:8],
		ServiceName: "control",
		Endpoint:    cfg.HTTP.AdvertiseAddr,
		Status:      "available",
		Version:     cfg.Service.Version,
	}, nil
}

// ProvideHeartbeatInterval 根据租约 TTL 推导运行时心跳间隔。
func ProvideHeartbeatInterval(cfg *controlconfig.Config) time.Duration {
	ttl := cfg.ETCD.LeaseTTL
	if ttl <= 1 {
		return time.Second
	}
	return time.Duration(ttl/2) * time.Second
}

// ProvideHTTPServer 创建 control 的 HTTP 服务。
func ProvideHTTPServer(cfg *controlconfig.Config, router *gin.Engine) *commonserver.HTTPServer {
	return commonserver.NewHTTPServer(cfg.HTTP.Addr, router)
}
