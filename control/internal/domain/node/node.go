// Package node 定义 control 节点查询的读模型。
package node

import (
	"context"
	"time"
)

const (
	// StatusAvailable 表示节点可用。
	StatusAvailable = "available"
)

// Node 是 control 暴露的节点读模型。
type Node struct {
	NodeID      string    `json:"node_id"`
	ServiceName string    `json:"service_name"`
	Endpoint    string    `json:"endpoint"`
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	HeartbeatAt time.Time `json:"heartbeat_at"`
}

// IsAvailable 判断节点是否可用。
func (n Node) IsAvailable() bool {
	return n.Status == StatusAvailable
}

// Repository 定义节点元数据的读取接口。
type Repository interface {
	List(ctx context.Context) ([]Node, error)
	ListByService(ctx context.Context, serviceName string) ([]Node, error)
	Find(ctx context.Context, nodeID string) (*Node, error)
}
