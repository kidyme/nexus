package main

import (
	"log/slog"
	"os"

	"github.com/kidyme/nexus/common/config"
	"github.com/kidyme/nexus/common/log"
	offline "github.com/kidyme/nexus/offline/internal"
)

func main() {
	log.Init(config.IsProd())
	if err := offline.Run(); err != nil {
		log.Error("offline 服务退出", slog.Any("error", err))
		os.Exit(1)
	}
}
