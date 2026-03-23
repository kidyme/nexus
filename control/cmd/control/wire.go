//go:build wireinject

//go:generate wire gen .

package main

import (
	"github.com/google/wire"
	control "github.com/kidyme/nexus/control/internal"
	feedbackapp "github.com/kidyme/nexus/control/internal/application/feedback"
	itemapp "github.com/kidyme/nexus/control/internal/application/item"
	nodeapp "github.com/kidyme/nexus/control/internal/application/node"
	userapp "github.com/kidyme/nexus/control/internal/application/user"
	feedbackinfra "github.com/kidyme/nexus/control/internal/infrastructure/feedback"
	iteminfra "github.com/kidyme/nexus/control/internal/infrastructure/item"
	nodeinfra "github.com/kidyme/nexus/control/internal/infrastructure/node"
	refreshmetainfra "github.com/kidyme/nexus/control/internal/infrastructure/refreshmeta"
	userinfra "github.com/kidyme/nexus/control/internal/infrastructure/user"
	httpport "github.com/kidyme/nexus/control/internal/port/http"
)

// InitializeApp 创建完成依赖注入的 control 运行时。
func InitializeApp() (*control.App, func(), error) {
	wire.Build(
		control.ProviderSet,
		nodeinfra.ProviderSet,
		nodeapp.ProviderSet,
		refreshmetainfra.ProviderSet,
		userinfra.ProviderSet,
		userapp.ProviderSet,
		iteminfra.ProviderSet,
		itemapp.ProviderSet,
		feedbackinfra.ProviderSet,
		feedbackapp.ProviderSet,
		httpport.ProviderSet,
	)
	return nil, nil, nil
}
