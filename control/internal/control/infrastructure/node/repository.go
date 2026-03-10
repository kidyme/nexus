// Package node provides node repository implementations.
package node

import (
	"context"

	"github.com/kidyme/nexus/common/registry"
	nodedomain "github.com/kidyme/nexus/control/internal/control/domain/node"
)

// Repository is a registry-backed node repository.
type Repository struct {
	registry registry.Registry
}

// NewRepository creates a node repository backed by common/registry.
func NewRepository(nodeRegistry registry.Registry) *Repository {
	return &Repository{registry: nodeRegistry}
}

// List returns all nodes.
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

// ListByService returns all nodes for a service.
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

// Find returns a single node by its globally unique ID.
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
