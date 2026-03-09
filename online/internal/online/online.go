package online

import (
	"log/slog"

	"github.com/kidyme/nexus/common/log"
)

// Run 启动 online 服务的内部逻辑。
func Run() error {
	log.Info("online 服务启动", slog.String("service", "online"))
	// TODO: 加载配置、实时召回与重排
	return nil
}
