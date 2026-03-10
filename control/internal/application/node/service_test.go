package node

import (
	"context"
	"testing"
	"time"

	nodedomain "github.com/kidyme/nexus/control/internal/control/domain/node"
)

type fakeRepository struct {
	allNodes  []nodedomain.Node
	nodes     map[string][]nodedomain.Node
	probeNode *nodedomain.Node
	probeErr  error
}

func (f *fakeRepository) List(_ context.Context) ([]nodedomain.Node, error) {
	return f.allNodes, nil
}

func (f *fakeRepository) ListByService(_ context.Context, serviceName string) ([]nodedomain.Node, error) {
	return f.nodes[serviceName], nil
}

func (f *fakeRepository) Find(_ context.Context, _ string) (*nodedomain.Node, error) {
	if f.probeErr != nil {
		return nil, f.probeErr
	}
	return f.probeNode, nil
}

func TestListByService(t *testing.T) {
	service := NewService(&fakeRepository{
		nodes: map[string][]nodedomain.Node{
			"online": {
				{
					NodeID:      "online-01",
					ServiceName: "online",
					Endpoint:    "127.0.0.1:8082",
					Status:      "online",
					Version:     "v0.1.0",
					HeartbeatAt: time.Now().UTC(),
				},
			},
		},
	})

	nodes, err := service.ListByService(context.Background(), "online")
	if err != nil {
		t.Fatalf("list nodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].NodeID != "online-01" {
		t.Fatalf("expected node id online-01, got %s", nodes[0].NodeID)
	}
}

func TestList(t *testing.T) {
	service := NewService(&fakeRepository{
		allNodes: []nodedomain.Node{
			{
				NodeID:      "online-01",
				ServiceName: "online",
				Endpoint:    "127.0.0.1:8082",
			},
			{
				NodeID:      "offline-01",
				ServiceName: "offline",
				Endpoint:    "127.0.0.1:8081",
			},
		},
	})

	nodes, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list all nodes: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestListByServiceRequiresServiceName(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.ListByService(context.Background(), "")
	if err != ErrServiceNameRequired {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

func TestFind(t *testing.T) {
	service := NewService(&fakeRepository{
		probeNode: &nodedomain.Node{
			NodeID:      "online-01",
			ServiceName: "online",
			Endpoint:    "127.0.0.1:8082",
		},
	})

	node, err := service.Find(context.Background(), "online-01")
	if err != nil {
		t.Fatalf("get node: %v", err)
	}
	if node.NodeID != "online-01" {
		t.Fatalf("expected node id online-01, got %s", node.NodeID)
	}
}

func TestFindRequiresNodeID(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.Find(context.Background(), "")
	if err != ErrNodeIDRequired {
		t.Fatalf("expected ErrNodeIDRequired, got %v", err)
	}
}
