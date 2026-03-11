// Package node 提供 control 的节点查询用例。
package node

import (
	"context"
	"errors"
	"strings"

	nodedomain "github.com/kidyme/nexus/control/internal/domain/node"
)

var ErrServiceNameRequired = errors.New("node application: service name is required")
var ErrNodeIDRequired = errors.New("node application: node id is required")

// Service 编排 node 查询用例。
type Service struct {
	repository nodedomain.Repository
}

// NewService 创建 node 查询服务。
func NewService(repository nodedomain.Repository) *Service {
	return &Service{repository: repository}
}

// List 返回全部节点。
func (s *Service) List(ctx context.Context) ([]nodedomain.Node, error) {
	return s.repository.List(ctx)
}

// ListByService 返回指定服务下的全部节点。
func (s *Service) ListByService(ctx context.Context, serviceName string) ([]nodedomain.Node, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return nil, ErrServiceNameRequired
	}

	return s.repository.ListByService(ctx, serviceName)
}

// Find 返回单个节点信息。
func (s *Service) Find(ctx context.Context, nodeID string) (*nodedomain.Node, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return nil, ErrNodeIDRequired
	}

	return s.repository.Find(ctx, nodeID)
}
