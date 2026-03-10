package httpport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kidyme/nexus/common/registry"
	nodeapp "github.com/kidyme/nexus/control/internal/control/application/node"
	nodedomain "github.com/kidyme/nexus/control/internal/control/domain/node"
)

func TestListNodesRoute(t *testing.T) {
	router := NewRouter(Handlers{Node: NewNodeHandler(nodeapp.NewService(&fakeNodeRepository{
		listByServiceFn: func(context.Context, string) ([]nodedomain.Node, error) {
			return []nodedomain.Node{
				{
					NodeID:      "online-01",
					ServiceName: "online",
					Endpoint:    "127.0.0.1:8082",
				},
			}, nil
		},
	}))})

	req := httptest.NewRequest(http.MethodGet, "/api/meta/nodes?service=online", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"node_id":"online-01"`) {
		t.Fatalf("expected response to contain node id, got %s", rec.Body.String())
	}
}

func TestListNodesRouteAll(t *testing.T) {
	router := NewRouter(Handlers{Node: NewNodeHandler(nodeapp.NewService(&fakeNodeRepository{
		listFn: func(context.Context) ([]nodedomain.Node, error) {
			return []nodedomain.Node{
				{NodeID: "online-01", ServiceName: "online", Endpoint: "127.0.0.1:8082"},
				{NodeID: "offline-01", ServiceName: "offline", Endpoint: "127.0.0.1:8081"},
			}, nil
		},
	}))})

	req := httptest.NewRequest(http.MethodGet, "/api/meta/nodes", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"node_id":"offline-01"`) {
		t.Fatalf("expected response to contain all nodes, got %s", rec.Body.String())
	}
}

func TestGetNodeRoute(t *testing.T) {
	router := NewRouter(Handlers{Node: NewNodeHandler(nodeapp.NewService(&fakeNodeRepository{
		findFn: func(context.Context, string) (*nodedomain.Node, error) {
			return &nodedomain.Node{
				NodeID:      "online-01",
				ServiceName: "online",
				Endpoint:    "127.0.0.1:8082",
			}, nil
		},
	}))})

	req := httptest.NewRequest(http.MethodGet, "/api/meta/nodes/online-01", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"service_name":"online"`) {
		t.Fatalf("expected response to contain service name, got %s", rec.Body.String())
	}
}

func TestGetNodeRouteNotFound(t *testing.T) {
	router := NewRouter(Handlers{Node: NewNodeHandler(nodeapp.NewService(&fakeNodeRepository{
		findFn: func(context.Context, string) (*nodedomain.Node, error) {
			return nil, registry.ErrNodeNotFound
		},
	}))})

	req := httptest.NewRequest(http.MethodGet, "/api/meta/nodes/missing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}
