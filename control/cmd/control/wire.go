//go:build wireinject

//go:generate wire gen .

package main

import (
	"github.com/google/wire"
	control "github.com/kidyme/nexus/control/internal/control"
	nodeapp "github.com/kidyme/nexus/control/internal/control/application/node"
	nodeinfra "github.com/kidyme/nexus/control/internal/control/infrastructure/node"
	httpport "github.com/kidyme/nexus/control/internal/control/port/http"
)

// InitializeApp creates the fully wired control runtime.
func InitializeApp() (*control.App, func(), error) {
	wire.Build(
		control.ProviderSet,
		nodeinfra.ProviderSet,
		nodeapp.ProviderSet,
		httpport.ProviderSet,
	)
	return nil, nil, nil
}
