//go:build integration

package registry

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEtcdRegistryIntegration(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		Endpoints:      integrationEndpoints(),
		DialTimeout:    5 * time.Second,
		RequestTimeout: 3 * time.Second,
		KeyPrefix:      "/nexus/meta/test-" + time.Now().UTC().Format("20060102150405.000000000"),
		LeaseTTL:       10,
	}

	reg, err := NewEtcdRegistry(cfg)
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	defer func() {
		_ = reg.Deregister(context.Background())
		_ = reg.Close()
	}()

	node := Node{
		NodeID:      "online-01",
		ServiceName: "online",
		Endpoint:    "127.0.0.1:8082",
		Status:      "online",
		Version:     "v0.1.0",
	}

	if err := reg.Register(ctx, node); err != nil {
		t.Fatalf("register node: %v", err)
	}

	probed, err := reg.Probe(ctx, node.ServiceName, node.NodeID)
	if err != nil {
		t.Fatalf("probe node after register: %v", err)
	}
	if probed.NodeID != node.NodeID {
		t.Fatalf("expected node id %s, got %s", node.NodeID, probed.NodeID)
	}
	if probed.HeartbeatAt.IsZero() {
		t.Fatal("expected heartbeat_at to be set")
	}

	discovered, err := reg.Discover(ctx, node.ServiceName)
	if err != nil {
		t.Fatalf("discover nodes: %v", err)
	}
	if !containsNode(discovered, node.NodeID) {
		t.Fatalf("expected discovered nodes to contain %s", node.NodeID)
	}

	beforeHeartbeat := probed.HeartbeatAt
	time.Sleep(10 * time.Millisecond)
	if err := reg.Heartbeat(ctx); err != nil {
		t.Fatalf("heartbeat node: %v", err)
	}

	probedAfterHeartbeat, err := reg.Probe(ctx, node.ServiceName, node.NodeID)
	if err != nil {
		t.Fatalf("probe node after heartbeat: %v", err)
	}
	if !probedAfterHeartbeat.HeartbeatAt.After(beforeHeartbeat) {
		t.Fatalf("expected heartbeat_at to move forward, before=%s after=%s", beforeHeartbeat, probedAfterHeartbeat.HeartbeatAt)
	}

	if err := reg.Deregister(ctx); err != nil {
		t.Fatalf("deregister node: %v", err)
	}

	_, err = reg.Probe(ctx, node.ServiceName, node.NodeID)
	if err != ErrNodeNotFound {
		t.Fatalf("expected ErrNodeNotFound after deregister, got %v", err)
	}
}

func integrationEndpoints() []string {
	raw := strings.TrimSpace(os.Getenv("ETCD_ENDPOINTS"))
	if raw == "" {
		return []string{"127.0.0.1:2379"}
	}

	parts := strings.Split(raw, ",")
	endpoints := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			endpoints = append(endpoints, part)
		}
	}
	return endpoints
}

func containsNode(nodes []Node, nodeID string) bool {
	for _, node := range nodes {
		if node.NodeID == nodeID {
			return true
		}
	}
	return false
}
