package registry

import "testing"

func TestNewEtcdRegistryRequiresEndpoints(t *testing.T) {
	_, err := NewEtcdRegistry(Config{})
	if err != ErrEndpointsRequired {
		t.Fatalf("expected ErrEndpointsRequired, got %v", err)
	}
}

func TestNodeKey(t *testing.T) {
	got := nodeKey("/nexus/meta", "online", "online-01")
	want := "/nexus/meta/nodes/online/online-01"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestServicePrefix(t *testing.T) {
	got := servicePrefix("/nexus/meta", "control")
	want := "/nexus/meta/nodes/control/"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestNodesPrefix(t *testing.T) {
	got := nodesPrefix("/nexus/meta")
	want := "/nexus/meta/nodes/"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestValidateNode(t *testing.T) {
	err := validateNode(Node{
		NodeID:      "offline-01",
		ServiceName: "offline",
		Endpoint:    "127.0.0.1:8081",
	})
	if err != nil {
		t.Fatalf("expected valid node, got %v", err)
	}
}
