package main

import (
	"log/slog"
	"os"

	"github.com/kidyme/nexus/common/config"
	"github.com/kidyme/nexus/common/log"
	online "github.com/kidyme/nexus/online/internal/online"
)

func main() {
	log.Init(config.IsProd())
	if err := online.Run(); err != nil {
		log.Error("online 服务退出", slog.Any("error", err))
		os.Exit(1)
	}
}
