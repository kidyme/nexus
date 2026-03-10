//go:build wireinject

//go:generate wire gen .

package main

import (
	"github.com/google/wire"
	control "github.com/kidyme/nexus/control/internal"
	nodeapp "github.com/kidyme/nexus/control/internal/application/node"
	nodeinfra "github.com/kidyme/nexus/control/internal/infrastructure/node"
	httpport "github.com/kidyme/nexus/control/internal/port/http"
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
