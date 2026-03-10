package node

import "github.com/google/wire"

// ProviderSet provides node application services.
var ProviderSet = wire.NewSet(NewService)
