package feedback

import "github.com/google/wire"

// ProviderSet 提供反馈应用服务依赖。
var ProviderSet = wire.NewSet(NewService)
