package recommendation

import (
	"fmt"

	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
)

// Builder 定义召回器构造函数。
type Builder func() recdomain.Recaller

// Registry 管理召回器注册与构建。
type Registry struct {
	builders map[string]Builder
}

// NewRegistry 创建召回器注册表。
func NewRegistry() *Registry {
	return &Registry{builders: make(map[string]Builder)}
}

// Register 注册召回器。
func (r *Registry) Register(name string, builder Builder) {
	if name == "" || builder == nil {
		return
	}
	r.builders[name] = builder
}

// Build 按顺序构建召回器。
func (r *Registry) Build(names []string) ([]recdomain.Recaller, error) {
	result := make([]recdomain.Recaller, 0, len(names))
	for _, name := range names {
		builder, ok := r.builders[name]
		if !ok {
			return nil, fmt.Errorf("unsupported recaller: %s", name)
		}
		result = append(result, builder())
	}
	return result, nil
}
