package item

import "github.com/google/wire"

// ProviderSet 提供物品应用服务依赖。
var ProviderSet = wire.NewSet(NewService)
