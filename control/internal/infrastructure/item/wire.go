package item

import (
	"github.com/google/wire"
	itemdomain "github.com/kidyme/nexus/control/internal/domain/item"
)

// ProviderSet 提供物品仓储实现依赖。
var ProviderSet = wire.NewSet(
	NewRepository,
	wire.Bind(new(itemdomain.Repository), new(*Repository)),
)
