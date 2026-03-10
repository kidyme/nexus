package httpport

import "github.com/google/wire"

// ProviderSet provides HTTP handlers and router assembly.
var ProviderSet = wire.NewSet(
	NewCommonHandler,
	NewNodeHandler,
	wire.Struct(new(Handlers), "*"),
	NewRouter,
)
