package node

import "github.com/google/wire"

// ProviderSet 提供 node 应用服务依赖。
var ProviderSet = wire.NewSet(NewService)
