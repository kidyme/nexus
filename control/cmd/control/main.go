package main

import (
	"log/slog"
	"os"

	"github.com/kidyme/nexus/common/config"
	"github.com/kidyme/nexus/common/log"
	control "github.com/kidyme/nexus/control/internal/control"
)

func main() {
	log.Init(config.IsProd())
	if err := control.Run(); err != nil {
		log.Error("control 服务退出", slog.Any("error", err))
		os.Exit(1)
	}
}
