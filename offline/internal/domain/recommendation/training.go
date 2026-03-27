package recommendation

import "context"

// Trainer 定义单个模型训练器。
type Trainer interface {
	Name() string
	Train(ctx context.Context) error
}

// Backend 定义训练执行后端，后续可扩展 CPU / GPU。
type Backend interface {
	Name() string
	Run(ctx context.Context, trainer Trainer) error
}
