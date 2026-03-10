package control

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/google/wire"
	"github.com/kidyme/nexus/common/log"
	"github.com/kidyme/nexus/common/registry"
	commonserver "github.com/kidyme/nexus/common/server"
	controlconfig "github.com/kidyme/nexus/control/config"
)

// ProviderSet provides control runtime dependencies.
var ProviderSet = wire.NewSet(
	ProvideConfig,
	ProvideRegistry,
	ProvideSelfNode,
	ProvideHeartbeatInterval,
	ProvideHTTPServer,
	NewApp,
)

// ProvideConfig loads the runtime config for control.
func ProvideConfig() (*controlconfig.Config, error) {
	return controlconfig.Load()
}

// ProvideRegistry creates the service registry and its cleanup function.
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

// ProvideSelfNode creates the current control node metadata for registry registration.
func ProvideSelfNode(cfg *controlconfig.Config) (registry.Node, error) {
	if cfg.HTTP.AdvertiseAddr == "" {
		return registry.Node{}, fmt.Errorf("control: http.advertiseAddr is required")
	}
	if cfg.Service.Version == "" {
		return registry.Node{}, fmt.Errorf("control: service.version is required")
	}
	return registry.Node{
		NodeID:      uuid.NewString(),
		ServiceName: "control",
		Endpoint:    cfg.HTTP.AdvertiseAddr,
		Status:      "available",
		Version:     cfg.Service.Version,
	}, nil
}

// ProvideHeartbeatInterval derives the runtime heartbeat interval from the lease TTL.
func ProvideHeartbeatInterval(cfg *controlconfig.Config) time.Duration {
	ttl := cfg.ETCD.LeaseTTL
	if ttl <= 1 {
		return time.Second
	}
	return time.Duration(ttl/2) * time.Second
}

// ProvideHTTPServer creates the HTTP server for control.
func ProvideHTTPServer(cfg *controlconfig.Config, router *gin.Engine) *commonserver.HTTPServer {
	return commonserver.NewHTTPServer(cfg.HTTP.Addr, router)
}
