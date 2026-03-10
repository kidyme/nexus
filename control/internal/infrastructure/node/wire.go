package node

import (
	"github.com/google/wire"
	nodedomain "github.com/kidyme/nexus/control/internal/control/domain/node"
)

// ProviderSet provides node repository implementations.
var ProviderSet = wire.NewSet(
	NewRepository,
	wire.Bind(new(nodedomain.Repository), new(*Repository)),
)
