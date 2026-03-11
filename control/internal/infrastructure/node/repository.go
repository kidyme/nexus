// Package node 提供 node 仓储实现。
package node

import (
	"context"

	"github.com/kidyme/nexus/common/registry"
	nodedomain "github.com/kidyme/nexus/control/internal/domain/node"
)

// Repository 是基于 registry 的 node 仓储实现。
type Repository struct {
	registry registry.Registry
}

// NewRepository 创建基于 common/registry 的 node 仓储。
func NewRepository(nodeRegistry registry.Registry) *Repository {
	return &Repository{registry: nodeRegistry}
}

// List 返回全部节点。
func (r *Repository) List(ctx context.Context) ([]nodedomain.Node, error) {
	nodes, err := r.registry.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]nodedomain.Node, 0, len(nodes))
	for _, item := range nodes {
		result = append(result, toNode(item))
	}
	return result, nil
}

// ListByService 返回指定服务下的全部节点。
func (r *Repository) ListByService(ctx context.Context, serviceName string) ([]nodedomain.Node, error) {
	nodes, err := r.registry.Discover(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	result := make([]nodedomain.Node, 0, len(nodes))
	for _, item := range nodes {
		result = append(result, toNode(item))
	}
	return result, nil
}

// Find 根据全局唯一 node ID 返回单个节点。
func (r *Repository) Find(ctx context.Context, nodeID string) (*nodedomain.Node, error) {
	nodes, err := r.registry.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range nodes {
		if item.NodeID == nodeID {
			node := toNode(item)
			return &node, nil
		}
	}
	return nil, registry.ErrNodeNotFound
}

func toNode(src registry.Node) nodedomain.Node {
	return nodedomain.Node{
		NodeID:      src.NodeID,
		ServiceName: src.ServiceName,
		Endpoint:    src.Endpoint,
		Status:      src.Status,
		Version:     src.Version,
		HeartbeatAt: src.HeartbeatAt,
	}
}
