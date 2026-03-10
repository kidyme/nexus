// Package node defines the read model for control node queries.
package node

import (
	"context"
	"time"
)

const (
	// StatusAvailable marks a node as available.
	StatusAvailable = "available"
)

// Node is the node read model exposed by control.
type Node struct {
	NodeID      string    `json:"node_id"`
	ServiceName string    `json:"service_name"`
	Endpoint    string    `json:"endpoint"`
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	HeartbeatAt time.Time `json:"heartbeat_at"`
}

// IsAvailable reports whether the node is available.
func (n Node) IsAvailable() bool {
	return n.Status == StatusAvailable
}

// Repository defines read access to node metadata.
type Repository interface {
	List(ctx context.Context) ([]Node, error)
	ListByService(ctx context.Context, serviceName string) ([]Node, error)
	Find(ctx context.Context, nodeID string) (*Node, error)
}
