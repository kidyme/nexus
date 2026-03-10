package httpport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	nodedomain "github.com/kidyme/nexus/control/internal/control/domain/node"
)

type fakeNodeRepository struct {
	listFn          func(ctx context.Context) ([]nodedomain.Node, error)
	listByServiceFn func(ctx context.Context, serviceName string) ([]nodedomain.Node, error)
	findFn          func(ctx context.Context, nodeID string) (*nodedomain.Node, error)
}

func (f *fakeNodeRepository) List(ctx context.Context) ([]nodedomain.Node, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx)
}

func (f *fakeNodeRepository) ListByService(ctx context.Context, serviceName string) ([]nodedomain.Node, error) {
	if f.listByServiceFn == nil {
		return nil, nil
	}
	return f.listByServiceFn(ctx, serviceName)
}

func (f *fakeNodeRepository) Find(ctx context.Context, nodeID string) (*nodedomain.Node, error) {
	if f.findFn == nil {
		return nil, nil
	}
	return f.findFn(ctx, nodeID)
}

func TestHealthz(t *testing.T) {
	router := NewRouter(Handlers{Common: NewCommonHandler()})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
