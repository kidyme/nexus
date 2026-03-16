package user

import (
	"github.com/google/wire"
	userdomain "github.com/kidyme/nexus/control/internal/domain/user"
)

// ProviderSet 提供用户仓储实现依赖。
var ProviderSet = wire.NewSet(
	NewRepository,
	wire.Bind(new(userdomain.Repository), new(*Repository)),
)
