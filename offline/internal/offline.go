package offline

import (
	"log/slog"

	"github.com/kidyme/nexus/common/log"
)

// Run 启动 offline 服务的内部逻辑。
func Run() error {
	log.Info("offline 服务启动", slog.String("service", "offline"))
	// 待办：加载配置、离线任务、写入缓存
	return nil
}
