// Package registry 提供共享的服务注册辅助能力。
package registry

import (
	"context"
	"errors"
	"time"
)

var ErrNodeNotFound = errors.New("registry: node not found")
var ErrEndpointsRequired = errors.New("registry: endpoints are required")

// Config 描述共享注册中心配置。
type Config struct {
	Endpoints      []string      `mapstructure:"endpoints"`
	Username       string        `mapstructure:"username"`
	Password       string        `mapstructure:"password"`
	DialTimeout    time.Duration `mapstructure:"dialTimeout"`
	RequestTimeout time.Duration `mapstructure:"requestTimeout"`
	KeyPrefix      string        `mapstructure:"keyPrefix"`
	LeaseTTL       int64         `mapstructure:"leaseTTL"`
}

// Node 描述注册中心中存储的服务节点。
type Node struct {
	NodeID      string    `json:"node_id"`
	ServiceName string    `json:"service_name"`
	Endpoint    string    `json:"endpoint"`
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	HeartbeatAt time.Time `json:"heartbeat_at"`
}

// Registry 定义共享服务注册契约。
type Registry interface {
	Register(ctx context.Context, node Node) error
	Heartbeat(ctx context.Context) error
	List(ctx context.Context) ([]Node, error)
	Discover(ctx context.Context, serviceName string) ([]Node, error)
	Probe(ctx context.Context, serviceName, nodeID string) (*Node, error)
	Deregister(ctx context.Context) error
	Close() error
}
