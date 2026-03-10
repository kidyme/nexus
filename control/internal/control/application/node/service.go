// Package node provides node query use cases for control.
package node

import (
	"context"
	"errors"
	"strings"

	nodedomain "github.com/kidyme/nexus/control/internal/control/domain/node"
)

var ErrServiceNameRequired = errors.New("node application: service name is required")
var ErrNodeIDRequired = errors.New("node application: node id is required")

// Service orchestrates node query use cases.
type Service struct {
	repository nodedomain.Repository
}

// NewService creates a node query service.
func NewService(repository nodedomain.Repository) *Service {
	return &Service{repository: repository}
}

// List returns all nodes.
func (s *Service) List(ctx context.Context) ([]nodedomain.Node, error) {
	return s.repository.List(ctx)
}

// ListByService returns all nodes for a service.
func (s *Service) ListByService(ctx context.Context, serviceName string) ([]nodedomain.Node, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return nil, ErrServiceNameRequired
	}

	return s.repository.ListByService(ctx, serviceName)
}

// Find returns the node info for a single node.
func (s *Service) Find(ctx context.Context, nodeID string) (*nodedomain.Node, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return nil, ErrNodeIDRequired
	}

	return s.repository.Find(ctx, nodeID)
}
