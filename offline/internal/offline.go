// Package offline 提供 offline 服务的私有运行时实现。
package offline

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kidyme/nexus/common/log"
	"github.com/kidyme/nexus/common/redisx"
	offlineconfig "github.com/kidyme/nexus/offline/config"
	recallerapp "github.com/kidyme/nexus/offline/internal/application/recall"
	recallcfitem "github.com/kidyme/nexus/offline/internal/application/recall/cf/item"
	recapp "github.com/kidyme/nexus/offline/internal/application/recommendation"
	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
	recinfra "github.com/kidyme/nexus/offline/internal/infrastructure/recommendation"
	userinfra "github.com/kidyme/nexus/offline/internal/infrastructure/user"
)

// Run 启动 offline 服务的内部逻辑。
func Run() error {
	cfg := offlineconfig.MustLoad()

	db, cleanupDB, err := provideMySQL(cfg)
	if err != nil {
		return err
	}
	defer cleanupDB()

	redisClient, err := provideRedis(cfg)
	if err != nil {
		return err
	}

	userRepository := userinfra.NewRepository(db)
	cacheRepository := recinfra.NewCacheRepository(redisClient)
	sourceRepository := recinfra.NewSourceRepository(db)

	recallers, err := buildRecallers(cfg, sourceRepository, cacheRepository)
	if err != nil {
		return err
	}
	service := recapp.NewService(userRepository, cacheRepository, recallers, cfg.Recommend)
	app := NewApp(service, offlineconfig.MustParseDuration(cfg.Worker.TickInterval))

	log.Info("offline 服务启动",
		slog.String("service", "offline"),
		slog.String("tick_interval", cfg.Worker.TickInterval),
		slog.String("cache_expire", cfg.Recommend.CacheExpire),
		slog.String("active_user_ttl", cfg.Recommend.ActiveUserTTL),
		slog.Any("recallers", cfg.Recommend.EnabledRecallers()),
		slog.String("training_backend", cfg.Training.Backend))
	return app.Run()
}

func provideMySQL(cfg *offlineconfig.Config) (*sql.DB, func(), error) {
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
			log.Error("关闭 offline mysql 连接失败", "error", err)
		}
	}, nil
}

func provideRedis(cfg *offlineconfig.Config) (*redisx.Client, error) {
	client := redisx.NewClient(redisx.Options{
		Addr:         cfg.Redis.Addr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  offlineconfig.MustParseDuration(cfg.Redis.DialTimeout),
		ReadTimeout:  offlineconfig.MustParseDuration(cfg.Redis.ReadTimeout),
		WriteTimeout: offlineconfig.MustParseDuration(cfg.Redis.WriteTimeout),
	})
	if err := client.Ping(context.Background()); err != nil {
		return nil, err
	}
	return client, nil
}

func buildRecallers(cfg *offlineconfig.Config, source *recinfra.SourceRepository, cacheRepository *recinfra.CacheRepository) ([]recdomain.Recaller, error) {
	registry := recapp.NewRegistry()
	for _, recallerCfg := range cfg.Recommend.Recallers {
		switch typed := recallerCfg.(type) {
		case offlineconfig.PopularRecallerConfig:
			registry.Register(typed.Key(), func() (recdomain.Recaller, error) {
				return recallerapp.NewPopularRecaller(source, cfg.Recommend), nil
			})
		case offlineconfig.LatestRecallerConfig:
			registry.Register(typed.Key(), func() (recdomain.Recaller, error) {
				return recallerapp.NewLatestRecaller(source, cfg.Recommend), nil
			})
		case offlineconfig.ItemToItemUsersRecallerConfig:
			recallerConfig := typed
			registry.Register(recallerConfig.Key(), func() (recdomain.Recaller, error) {
				return recallcfitem.NewRecaller(recallerConfig.Key(), source, cacheRepository, cfg.Recommend, recallerConfig)
			})
		case offlineconfig.ItemToItemTagsRecallerConfig:
			recallerConfig := typed
			registry.Register(recallerConfig.Key(), func() (recdomain.Recaller, error) {
				return recallcfitem.NewRecaller(recallerConfig.Key(), source, cacheRepository, cfg.Recommend, recallerConfig)
			})
		case offlineconfig.ItemToItemEmbeddingRecallerConfig:
			recallerConfig := typed
			registry.Register(recallerConfig.Key(), func() (recdomain.Recaller, error) {
				return recallcfitem.NewRecaller(recallerConfig.Key(), source, cacheRepository, cfg.Recommend, recallerConfig)
			})
		case offlineconfig.ItemToItemAutoRecallerConfig:
			recallerConfig := typed
			registry.Register(recallerConfig.Key(), func() (recdomain.Recaller, error) {
				return recallcfitem.NewRecaller(recallerConfig.Key(), source, cacheRepository, cfg.Recommend, recallerConfig)
			})
		case offlineconfig.UserToUserRecallerConfig, offlineconfig.MFRecallerConfig, offlineconfig.ExternalRecallerConfig:
			// 后续具体实现时在这里注册。
		}
	}
	return registry.Build(cfg.Recommend.EnabledRecallerKeys())
}
