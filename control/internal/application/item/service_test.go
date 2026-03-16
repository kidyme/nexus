package item

import (
	"context"
	"testing"

	itemdomain "github.com/kidyme/nexus/control/internal/domain/item"
)

type fakeItemRepository struct {
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

func (f *fakeItemRepository) Create(ctx context.Context, item itemdomain.Item) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, item)
}

func (f *fakeItemRepository) CreateBatch(ctx context.Context, items []itemdomain.Item) error {
	if f.createBatchFn == nil {
		return nil
	}
	return f.createBatchFn(ctx, items)
}

func (f *fakeItemRepository) Update(ctx context.Context, item itemdomain.Item) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, item)
}

func (f *fakeItemRepository) UpdateBatch(ctx context.Context, items []itemdomain.Item) error {
	if f.updateBatchFn == nil {
		return nil
	}
	return f.updateBatchFn(ctx, items)
}

func (f *fakeItemRepository) Delete(ctx context.Context, itemID string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, itemID)
}

func (f *fakeItemRepository) DeleteBatch(ctx context.Context, itemIDs []string) error {
	if f.deleteBatchFn == nil {
		return nil
	}
	return f.deleteBatchFn(ctx, itemIDs)
}

func (f *fakeItemRepository) Find(ctx context.Context, itemID string) (*itemdomain.Item, error) {
	if f.findFn == nil {
		return nil, nil
	}
	return f.findFn(ctx, itemID)
}

func (f *fakeItemRepository) List(ctx context.Context) ([]itemdomain.Item, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx)
}

func (f *fakeItemRepository) ListPage(ctx context.Context, page, size int) ([]itemdomain.Item, int64, error) {
	if f.listPageFn == nil {
		return nil, 0, nil
	}
	return f.listPageFn(ctx, page, size)
}

func TestCreateDefaultsTimestamp(t *testing.T) {
	service := NewService(&fakeItemRepository{
		createFn: func(_ context.Context, item itemdomain.Item) error {
			if item.Timestamp.IsZero() {
				t.Fatal("expected timestamp to be defaulted")
			}
			return nil
		},
	})

	if err := service.Create(context.Background(), itemdomain.Item{ItemID: "i-1"}); err != nil {
		t.Fatalf("create item: %v", err)
	}
}

func TestCreateBatchDefaultsTimestamp(t *testing.T) {
	service := NewService(&fakeItemRepository{
		createBatchFn: func(_ context.Context, items []itemdomain.Item) error {
			if len(items) != 2 {
				t.Fatalf("expected 2 items, got %d", len(items))
			}
			for i, item := range items {
				if item.Timestamp.IsZero() {
					t.Fatalf("expected item %d timestamp to be defaulted", i)
				}
			}
			return nil
		},
	})

	if err := service.CreateBatch(context.Background(), []itemdomain.Item{
		{ItemID: "i-1"},
		{ItemID: "i-2"},
	}); err != nil {
		t.Fatalf("create batch items: %v", err)
	}
}

func TestListPageValidatesPagination(t *testing.T) {
	service := NewService(&fakeItemRepository{})

	if _, _, err := service.ListPage(context.Background(), 0, 20); err != ErrInvalidPage {
		t.Fatalf("expected ErrInvalidPage, got %v", err)
	}
	if _, _, err := service.ListPage(context.Background(), 1, 0); err != ErrInvalidSize {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}
