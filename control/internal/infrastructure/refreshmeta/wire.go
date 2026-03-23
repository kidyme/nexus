package refreshmeta

import (
	"github.com/google/wire"
	refreshmetadomain "github.com/kidyme/nexus/control/internal/domain/refreshmeta"
)

// ProviderSet 提供刷新元数据仓储依赖。
var ProviderSet = wire.NewSet(
	NewRepository,
	wire.Bind(new(refreshmetadomain.Repository), new(*Repository)),
)
