package user

import "github.com/google/wire"

// ProviderSet 提供用户应用服务依赖。
var ProviderSet = wire.NewSet(NewService)
