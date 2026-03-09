// Package registry provides shared service registration helpers.
package registry

import (
	"context"
	"errors"
	"time"
)

var ErrNodeNotFound = errors.New("registry: node not found")
var ErrEndpointsRequired = errors.New("registry: endpoints are required")

// Config describes shared registry configuration.
type Config struct {
	Endpoints      []string      `mapstructure:"endpoints"`
	Username       string        `mapstructure:"username"`
	Password       string        `mapstructure:"password"`
	DialTimeout    time.Duration `mapstructure:"dialTimeout"`
	RequestTimeout time.Duration `mapstructure:"requestTimeout"`
	KeyPrefix      string        `mapstructure:"keyPrefix"`
	LeaseTTL       int64         `mapstructure:"leaseTTL"`
}

// Node describes a service node stored in the registry.
type Node struct {
	NodeID      string    `json:"node_id"`
	ServiceName string    `json:"service_name"`
	Endpoint    string    `json:"endpoint"`
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	HeartbeatAt time.Time `json:"heartbeat_at"`
}

// Registry defines the shared service registration contract.
type Registry interface {
	Register(ctx context.Context, node Node) error
	Heartbeat(ctx context.Context) error
	Discover(ctx context.Context, serviceName string) ([]Node, error)
	Probe(ctx context.Context, serviceName, nodeID string) (*Node, error)
	Deregister(ctx context.Context) error
	Close() error
}
