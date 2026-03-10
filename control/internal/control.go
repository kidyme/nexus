// Package control contains the private runtime of the control service.
package control

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kidyme/nexus/common/log"
	"github.com/kidyme/nexus/common/registry"
	commonserver "github.com/kidyme/nexus/common/server"
)

// App contains the runtime dependencies of the control service.
type App struct {
	httpServer        *commonserver.HTTPServer
	nodeRegistry      registry.Registry
	node              registry.Node
	heartbeatInterval time.Duration
}

// NewApp creates a control runtime with injected dependencies.
func NewApp(httpServer *commonserver.HTTPServer, nodeRegistry registry.Registry, node registry.Node, heartbeatInterval time.Duration) *App {
	return &App{
		httpServer:        httpServer,
		nodeRegistry:      nodeRegistry,
		node:              node,
		heartbeatInterval: heartbeatInterval,
	}
}

// Run starts the control runtime and handles graceful shutdown.
func (a *App) Run() error {
	if err := a.nodeRegistry.Register(context.Background(), a.node); err != nil {
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := a.nodeRegistry.Deregister(ctx); err != nil {
			log.Error("control 节点注销失败", "error", err)
		}
	}()

	errCh := make(chan error, 1)
	go func() {
		log.Info("control 服务启动", "service", "control", "addr", a.httpServer.Addr)
		errCh <- a.httpServer.Start()
	}()
	heartbeatErrCh := make(chan error, 1)
	heartbeatCtx, heartbeatCancel := context.WithCancel(context.Background())
	defer heartbeatCancel()
	go a.runHeartbeat(heartbeatCtx, heartbeatErrCh)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case err := <-heartbeatErrCh:
		if err != nil {
			return err
		}
	case <-signalCh:
	}

	heartbeatCancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.httpServer.Shutdown(ctx)
}

func (a *App) runHeartbeat(ctx context.Context, errCh chan<- error) {
	ticker := time.NewTicker(a.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.nodeRegistry.Heartbeat(ctx); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
		}
	}
}
