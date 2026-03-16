package httpport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	itemapp "github.com/kidyme/nexus/control/internal/application/item"
	itemdomain "github.com/kidyme/nexus/control/internal/domain/item"
)

type fakeHTTPItemRepository struct {
	createFn      func(context.Context, itemdomain.Item) error
	createBatchFn func(context.Context, []itemdomain.Item) error
	updateFn      func(context.Context, itemdomain.Item) error
	updateBatchFn func(context.Context, []itemdomain.Item) error
	deleteFn      func(context.Context, string) error
	deleteBatchFn func(context.Context, []string) error
	findFn        func(context.Context, string) (*itemdomain.Item, error)
	listFn        func(context.Context) ([]itemdomain.Item, error)
	listPageFn    func(context.Context, int, int) ([]itemdomain.Item, int64, error)
}

func (f *fakeHTTPItemRepository) Create(ctx context.Context, item itemdomain.Item) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, item)
}

func (f *fakeHTTPItemRepository) CreateBatch(ctx context.Context, items []itemdomain.Item) error {
	if f.createBatchFn == nil {
		return nil
	}
	return f.createBatchFn(ctx, items)
}

func (f *fakeHTTPItemRepository) Update(ctx context.Context, item itemdomain.Item) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, item)
}

func (f *fakeHTTPItemRepository) UpdateBatch(ctx context.Context, items []itemdomain.Item) error {
	if f.updateBatchFn == nil {
		return nil
	}
	return f.updateBatchFn(ctx, items)
}

func (f *fakeHTTPItemRepository) Delete(ctx context.Context, itemID string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, itemID)
}

func (f *fakeHTTPItemRepository) DeleteBatch(ctx context.Context, itemIDs []string) error {
	if f.deleteBatchFn == nil {
		return nil
	}
	return f.deleteBatchFn(ctx, itemIDs)
}

func (f *fakeHTTPItemRepository) Find(ctx context.Context, itemID string) (*itemdomain.Item, error) {
	if f.findFn == nil {
		return nil, nil
	}
	return f.findFn(ctx, itemID)
}

func (f *fakeHTTPItemRepository) List(ctx context.Context) ([]itemdomain.Item, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx)
}

func (f *fakeHTTPItemRepository) ListPage(ctx context.Context, page, size int) ([]itemdomain.Item, int64, error) {
	if f.listPageFn == nil {
		return nil, 0, nil
	}
	return f.listPageFn(ctx, page, size)
}

func TestGetItemRouteNotFound(t *testing.T) {
	router := NewRouter(Handlers{Item: NewItemHandler(itemapp.NewService(&fakeHTTPItemRepository{
		findFn: func(context.Context, string) (*itemdomain.Item, error) {
			return nil, itemdomain.ErrItemNotFound
		},
	}))})

	req := httptest.NewRequest(http.MethodGet, "/api/items/missing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestPatchItemsRoute(t *testing.T) {
	router := NewRouter(Handlers{Item: NewItemHandler(itemapp.NewService(&fakeHTTPItemRepository{
		findFn: func(_ context.Context, itemID string) (*itemdomain.Item, error) {
			return &itemdomain.Item{ItemID: itemID, Comment: "old"}, nil
		},
		updateBatchFn: func(_ context.Context, items []itemdomain.Item) error {
			if len(items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(items))
			}
			if items[0].Comment != "new" {
				t.Fatalf("expected merged comment to be updated, got %q", items[0].Comment)
			}
			return nil
		},
	}))})

	req := httptest.NewRequest(http.MethodPatch, "/api/items", strings.NewReader(`[{"item_id":"i-1","comment":"new"}]`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"item_id":"i-1"`) || !strings.Contains(rec.Body.String(), `"comment":"new"`) {
		t.Fatalf("expected response to contain updated item, got %s", rec.Body.String())
	}
}

func TestDeleteItemsRoute(t *testing.T) {
	router := NewRouter(Handlers{Item: NewItemHandler(itemapp.NewService(&fakeHTTPItemRepository{
		deleteBatchFn: func(_ context.Context, itemIDs []string) error {
			if len(itemIDs) != 2 {
				t.Fatalf("expected 2 item ids, got %d", len(itemIDs))
			}
			return nil
		},
	}))})

	req := httptest.NewRequest(http.MethodDelete, "/api/items", strings.NewReader(`["i-1","i-2"]`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func TestListItemsRouteWithPagination(t *testing.T) {
	router := NewRouter(Handlers{Item: NewItemHandler(itemapp.NewService(&fakeHTTPItemRepository{
		listPageFn: func(_ context.Context, page, size int) ([]itemdomain.Item, int64, error) {
			if page != 3 || size != 10 {
				t.Fatalf("expected page=3 size=10, got page=%d size=%d", page, size)
			}
			return []itemdomain.Item{{ItemID: "i-1", Comment: "ok"}}, 88, nil
		},
	}))})

	req := httptest.NewRequest(http.MethodGet, "/api/items?page=3&size=10", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"page":3`) || !strings.Contains(body, `"size":10`) || !strings.Contains(body, `"total":88`) {
		t.Fatalf("expected paged response metadata, got %s", body)
	}
	if !strings.Contains(body, `"item_id":"i-1"`) {
		t.Fatalf("expected response to contain item, got %s", body)
	}
}

func TestListItemsRouteRejectsInvalidPagination(t *testing.T) {
	router := NewRouter(Handlers{Item: NewItemHandler(itemapp.NewService(&fakeHTTPItemRepository{}))})

	req := httptest.NewRequest(http.MethodGet, "/api/items?page=1&size=0", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
