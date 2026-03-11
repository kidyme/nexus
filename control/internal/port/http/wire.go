package httpport

import "github.com/google/wire"

// ProviderSet 提供 HTTP handler 与路由装配依赖。
var ProviderSet = wire.NewSet(
	NewCommonHandler,
	NewNodeHandler,
	wire.Struct(new(Handlers), "*"),
	NewRouter,
)
