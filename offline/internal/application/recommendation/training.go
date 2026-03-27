package recommendation

import (
	"context"
	"fmt"

	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
)

// CPUBackend 是默认训练执行后端。
type CPUBackend struct{}

// Name 返回后端名称。
func (b CPUBackend) Name() string { return "cpu" }

// Run 执行训练。
func (b CPUBackend) Run(ctx context.Context, trainer recdomain.Trainer) error {
	return trainer.Train(ctx)
}

// GPUBackend 为后续 GPU 训练预留接口。
type GPUBackend struct{}

// Name 返回后端名称。
func (b GPUBackend) Name() string { return "gpu" }

// Run 执行训练。
func (b GPUBackend) Run(ctx context.Context, trainer recdomain.Trainer) error {
	return trainer.Train(ctx)
}

// TrainingCoordinator 编排训练器与执行后端。
type TrainingCoordinator struct {
	backend recdomain.Backend
}

// NewTrainingCoordinator 创建训练协调器。
func NewTrainingCoordinator(backend recdomain.Backend) *TrainingCoordinator {
	return &TrainingCoordinator{backend: backend}
}

// Run 执行单个训练器。
func (c *TrainingCoordinator) Run(ctx context.Context, trainer recdomain.Trainer) error {
	if c.backend == nil {
		return fmt.Errorf("offline training: backend is required")
	}
	return c.backend.Run(ctx, trainer)
}
