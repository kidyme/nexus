package user

import (
	"context"
	"testing"

	userdomain "github.com/kidyme/nexus/control/internal/domain/user"
)

type fakeRefreshMetaRepository struct {
	touchUsersFn func(context.Context, []string) error
	touchItemsFn func(context.Context, []string) error
}

func (f *fakeRefreshMetaRepository) TouchUsers(ctx context.Context, userIDs []string) error {
	if f.touchUsersFn == nil {
		return nil
	}
	return f.touchUsersFn(ctx, userIDs)
}

func (f *fakeRefreshMetaRepository) TouchItems(ctx context.Context, itemIDs []string) error {
	if f.touchItemsFn == nil {
		return nil
	}
	return f.touchItemsFn(ctx, itemIDs)
}

type fakeUserRepository struct {
	createFn      func(context.Context, userdomain.User) error
	createBatchFn func(context.Context, []userdomain.User) error
	updateFn      func(context.Context, userdomain.User) error
	updateBatchFn func(context.Context, []userdomain.User) error
	deleteFn      func(context.Context, string) error
	deleteBatchFn func(context.Context, []string) error
	findFn        func(context.Context, string) (*userdomain.User, error)
	listFn        func(context.Context) ([]userdomain.User, error)
	listPageFn    func(context.Context, int, int) ([]userdomain.User, int64, error)
}

func (f *fakeUserRepository) Create(ctx context.Context, user userdomain.User) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, user)
}

func (f *fakeUserRepository) CreateBatch(ctx context.Context, users []userdomain.User) error {
	if f.createBatchFn == nil {
		return nil
	}
	return f.createBatchFn(ctx, users)
}

func (f *fakeUserRepository) Update(ctx context.Context, user userdomain.User) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, user)
}

func (f *fakeUserRepository) UpdateBatch(ctx context.Context, users []userdomain.User) error {
	if f.updateBatchFn == nil {
		return nil
	}
	return f.updateBatchFn(ctx, users)
}

func (f *fakeUserRepository) Delete(ctx context.Context, userID string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, userID)
}

func (f *fakeUserRepository) DeleteBatch(ctx context.Context, userIDs []string) error {
	if f.deleteBatchFn == nil {
		return nil
	}
	return f.deleteBatchFn(ctx, userIDs)
}

func (f *fakeUserRepository) Find(ctx context.Context, userID string) (*userdomain.User, error) {
	if f.findFn == nil {
		return nil, nil
	}
	return f.findFn(ctx, userID)
}

func (f *fakeUserRepository) List(ctx context.Context) ([]userdomain.User, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx)
}

func (f *fakeUserRepository) ListPage(ctx context.Context, page, size int) ([]userdomain.User, int64, error) {
	if f.listPageFn == nil {
		return nil, 0, nil
	}
	return f.listPageFn(ctx, page, size)
}

func TestCreateTrimsUserID(t *testing.T) {
	service := NewService(&fakeUserRepository{
		createFn: func(_ context.Context, user userdomain.User) error {
			if user.UserID != "u-1" {
				t.Fatalf("expected trimmed user id, got %q", user.UserID)
			}
			return nil
		},
	}, &fakeRefreshMetaRepository{})

	if err := service.Create(context.Background(), userdomain.User{UserID: "  u-1  "}); err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func TestCreateBatchTrimsUserIDs(t *testing.T) {
	service := NewService(&fakeUserRepository{
		createBatchFn: func(_ context.Context, users []userdomain.User) error {
			if len(users) != 2 {
				t.Fatalf("expected 2 users, got %d", len(users))
			}
			if users[0].UserID != "u-1" || users[1].UserID != "u-2" {
				t.Fatalf("expected trimmed user ids, got %#v", users)
			}
			return nil
		},
	}, &fakeRefreshMetaRepository{})

	if err := service.CreateBatch(context.Background(), []userdomain.User{
		{UserID: "  u-1  "},
		{UserID: " u-2 "},
	}); err != nil {
		t.Fatalf("create batch users: %v", err)
	}
}

func TestListPageValidatesPagination(t *testing.T) {
	service := NewService(&fakeUserRepository{}, &fakeRefreshMetaRepository{})

	if _, _, err := service.ListPage(context.Background(), 0, 20); err != ErrInvalidPage {
		t.Fatalf("expected ErrInvalidPage, got %v", err)
	}
	if _, _, err := service.ListPage(context.Background(), 1, 0); err != ErrInvalidSize {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}

func TestCreateBatchTouchesUsers(t *testing.T) {
	service := NewService(&fakeUserRepository{
		createBatchFn: func(_ context.Context, users []userdomain.User) error {
			if len(users) != 2 {
				t.Fatalf("expected 2 users, got %d", len(users))
			}
			return nil
		},
	}, &fakeRefreshMetaRepository{
		touchUsersFn: func(_ context.Context, userIDs []string) error {
			if len(userIDs) != 2 || userIDs[0] != "u-1" || userIDs[1] != "u-2" {
				t.Fatalf("expected touched users [u-1 u-2], got %#v", userIDs)
			}
			return nil
		},
	})

	if err := service.CreateBatch(context.Background(), []userdomain.User{
		{UserID: "u-1"},
		{UserID: "u-2"},
	}); err != nil {
		t.Fatalf("create batch users: %v", err)
	}
}
