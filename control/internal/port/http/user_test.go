package httpport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	userapp "github.com/kidyme/nexus/control/internal/application/user"
	userdomain "github.com/kidyme/nexus/control/internal/domain/user"
)

type fakeHTTPUserRepository struct {
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

func (f *fakeHTTPUserRepository) Create(ctx context.Context, user userdomain.User) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, user)
}

func (f *fakeHTTPUserRepository) CreateBatch(ctx context.Context, users []userdomain.User) error {
	if f.createBatchFn == nil {
		return nil
	}
	return f.createBatchFn(ctx, users)
}

func (f *fakeHTTPUserRepository) Update(ctx context.Context, user userdomain.User) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, user)
}

func (f *fakeHTTPUserRepository) UpdateBatch(ctx context.Context, users []userdomain.User) error {
	if f.updateBatchFn == nil {
		return nil
	}
	return f.updateBatchFn(ctx, users)
}

func (f *fakeHTTPUserRepository) Delete(ctx context.Context, userID string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, userID)
}

func (f *fakeHTTPUserRepository) DeleteBatch(ctx context.Context, userIDs []string) error {
	if f.deleteBatchFn == nil {
		return nil
	}
	return f.deleteBatchFn(ctx, userIDs)
}

func (f *fakeHTTPUserRepository) Find(ctx context.Context, userID string) (*userdomain.User, error) {
	if f.findFn == nil {
		return nil, nil
	}
	return f.findFn(ctx, userID)
}

func (f *fakeHTTPUserRepository) List(ctx context.Context) ([]userdomain.User, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx)
}

func (f *fakeHTTPUserRepository) ListPage(ctx context.Context, page, size int) ([]userdomain.User, int64, error) {
	if f.listPageFn == nil {
		return nil, 0, nil
	}
	return f.listPageFn(ctx, page, size)
}

func TestCreateUsersSingleRoute(t *testing.T) {
	router := NewRouter(Handlers{User: NewUserHandler(userapp.NewService(&fakeHTTPUserRepository{
		findFn: func(_ context.Context, userID string) (*userdomain.User, error) {
			return &userdomain.User{UserID: userID, Comment: "ok"}, nil
		},
	}))})

	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"user_id":"u-1","comment":"ok"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"user_id":"u-1"`) {
		t.Fatalf("expected response to contain user id, got %s", rec.Body.String())
	}
	if rec.Header().Get("Location") != "/api/users/u-1" {
		t.Fatalf("expected location header, got %q", rec.Header().Get("Location"))
	}
}

func TestCreateUsersBatchRoute(t *testing.T) {
	router := NewRouter(Handlers{User: NewUserHandler(userapp.NewService(&fakeHTTPUserRepository{
		createBatchFn: func(_ context.Context, users []userdomain.User) error {
			if len(users) != 2 {
				t.Fatalf("expected 2 users, got %d", len(users))
			}
			return nil
		},
	}))})

	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`[{"user_id":"u-1"},{"user_id":"u-2"}]`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"user_id":"u-1"`) || !strings.Contains(rec.Body.String(), `"user_id":"u-2"`) {
		t.Fatalf("expected response to contain batched user ids, got %s", rec.Body.String())
	}
}

func TestPatchUserRoute(t *testing.T) {
	comment := "old"
	router := NewRouter(Handlers{User: NewUserHandler(userapp.NewService(&fakeHTTPUserRepository{
		findFn: func(_ context.Context, userID string) (*userdomain.User, error) {
			return &userdomain.User{UserID: userID, Comment: comment}, nil
		},
		updateBatchFn: func(_ context.Context, users []userdomain.User) error {
			if len(users) != 1 {
				t.Fatalf("expected 1 user, got %d", len(users))
			}
			if users[0].Comment != "new" {
				t.Fatalf("expected merged comment to be updated, got %q", users[0].Comment)
			}
			comment = users[0].Comment
			return nil
		},
	}))})

	req := httptest.NewRequest(http.MethodPatch, "/api/users/u-1", strings.NewReader(`{"comment":"new"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"comment":"new"`) {
		t.Fatalf("expected response to contain updated comment, got %s", rec.Body.String())
	}
}

func TestDeleteUsersRoute(t *testing.T) {
	router := NewRouter(Handlers{User: NewUserHandler(userapp.NewService(&fakeHTTPUserRepository{
		deleteBatchFn: func(_ context.Context, userIDs []string) error {
			if len(userIDs) != 2 {
				t.Fatalf("expected 2 user ids, got %d", len(userIDs))
			}
			return nil
		},
	}))})

	req := httptest.NewRequest(http.MethodDelete, "/api/users", strings.NewReader(`["u-1","u-2"]`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func TestListUsersRouteWithPagination(t *testing.T) {
	router := NewRouter(Handlers{User: NewUserHandler(userapp.NewService(&fakeHTTPUserRepository{
		listPageFn: func(_ context.Context, page, size int) ([]userdomain.User, int64, error) {
			if page != 2 || size != 5 {
				t.Fatalf("expected page=2 size=5, got page=%d size=%d", page, size)
			}
			return []userdomain.User{{UserID: "u-1", Comment: "ok"}}, 135, nil
		},
	}))})

	req := httptest.NewRequest(http.MethodGet, "/api/users?page=2&size=5", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"page":2`) || !strings.Contains(body, `"size":5`) || !strings.Contains(body, `"total":135`) {
		t.Fatalf("expected paged response metadata, got %s", body)
	}
	if !strings.Contains(body, `"user_id":"u-1"`) {
		t.Fatalf("expected response to contain user, got %s", body)
	}
}

func TestListUsersRouteRejectsInvalidPagination(t *testing.T) {
	router := NewRouter(Handlers{User: NewUserHandler(userapp.NewService(&fakeHTTPUserRepository{}))})

	req := httptest.NewRequest(http.MethodGet, "/api/users?page=0&size=20", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
