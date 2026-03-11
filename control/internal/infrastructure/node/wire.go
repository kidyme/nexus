package node

import (
	"github.com/google/wire"
	nodedomain "github.com/kidyme/nexus/control/internal/domain/node"
)

// ProviderSet 提供 node 仓储实现依赖。
var ProviderSet = wire.NewSet(
	NewRepository,
	wire.Bind(new(nodedomain.Repository), new(*Repository)),
)
