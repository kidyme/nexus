package feedback

import (
	"github.com/google/wire"
	feedbackdomain "github.com/kidyme/nexus/control/internal/domain/feedback"
)

// ProviderSet 提供反馈仓储实现依赖。
var ProviderSet = wire.NewSet(
	NewRepository,
	wire.Bind(new(feedbackdomain.Repository), new(*Repository)),
)
