package main

import (
	"log/slog"
	"os"

	"github.com/kidyme/nexus/common/config"
	"github.com/kidyme/nexus/common/log"
)

func main() {
	log.Init(config.IsProd())
	app, cleanup, err := InitializeApp()
	if err != nil {
		log.Error("control 服务初始化失败", slog.Any("error", err))
		os.Exit(1)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		log.Error("control 服务退出", slog.Any("error", err))
		os.Exit(1)
	}
}
