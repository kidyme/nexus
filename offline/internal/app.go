package offline

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kidyme/nexus/common/log"
)

// Refresher 定义周期刷新入口。
type Refresher interface {
	RefreshAll(ctx context.Context) error
}

// App 持有 offline 运行时依赖。
type App struct {
	refresher    Refresher
	tickInterval time.Duration
}

// NewApp 创建 offline 运行时。
func NewApp(refresher Refresher, tickInterval time.Duration) *App {
	return &App{
		refresher:    refresher,
		tickInterval: tickInterval,
	}
}

// Run 启动 offline worker。
func (a *App) Run() error {
	if a.tickInterval <= 0 {
		a.tickInterval = time.Minute
	}

	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.runLoop(ctx, errCh)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			return err
		}
	case <-signalCh:
	}
	cancel()
	return nil
}

func (a *App) runLoop(ctx context.Context, errCh chan<- error) {
	runOnce := func() bool {
		if err := a.refresher.RefreshAll(ctx); err != nil {
			log.Error("offline 刷新失败", "error", err)
			select {
			case errCh <- err:
			default:
			}
			return false
		}
		return true
	}

	if !runOnce() {
		return
	}
	for {
		timer := time.NewTimer(a.tickInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return
		case <-timer.C:
		}

		if !runOnce() {
			return
		}
	}
}
