package httpport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	feedbackapp "github.com/kidyme/nexus/control/internal/application/feedback"
	feedbackdomain "github.com/kidyme/nexus/control/internal/domain/feedback"
)

type fakeHTTPFeedbackRepository struct {
	createFn      func(context.Context, feedbackdomain.Feedback) error
	createBatchFn func(context.Context, []feedbackdomain.Feedback) error
	updateFn      func(context.Context, feedbackdomain.Feedback) error
	updateBatchFn func(context.Context, []feedbackdomain.Feedback) error
	deleteFn      func(context.Context, string, string, string) error
	deleteBatchFn func(context.Context, []feedbackdomain.Key) error
	findFn        func(context.Context, string, string, string) (*feedbackdomain.Feedback, error)
	listFn        func(context.Context, feedbackdomain.Filter) ([]feedbackdomain.Feedback, error)
	listPageFn    func(context.Context, feedbackdomain.Filter, int, int) ([]feedbackdomain.Feedback, int64, error)
}

func (f *fakeHTTPFeedbackRepository) Create(ctx context.Context, feedback feedbackdomain.Feedback) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, feedback)
}

func (f *fakeHTTPFeedbackRepository) CreateBatch(ctx context.Context, feedbacks []feedbackdomain.Feedback) error {
	if f.createBatchFn == nil {
		return nil
	}
	return f.createBatchFn(ctx, feedbacks)
}

func (f *fakeHTTPFeedbackRepository) Update(ctx context.Context, feedback feedbackdomain.Feedback) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, feedback)
}

func (f *fakeHTTPFeedbackRepository) UpdateBatch(ctx context.Context, feedbacks []feedbackdomain.Feedback) error {
	if f.updateBatchFn == nil {
		return nil
	}
	return f.updateBatchFn(ctx, feedbacks)
}

func (f *fakeHTTPFeedbackRepository) Delete(ctx context.Context, feedbackType, userID, itemID string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, feedbackType, userID, itemID)
}

func (f *fakeHTTPFeedbackRepository) DeleteBatch(ctx context.Context, keys []feedbackdomain.Key) error {
	if f.deleteBatchFn == nil {
		return nil
	}
	return f.deleteBatchFn(ctx, keys)
}

func (f *fakeHTTPFeedbackRepository) Find(ctx context.Context, feedbackType, userID, itemID string) (*feedbackdomain.Feedback, error) {
	if f.findFn == nil {
		return nil, nil
	}
	return f.findFn(ctx, feedbackType, userID, itemID)
}

func (f *fakeHTTPFeedbackRepository) List(ctx context.Context, filter feedbackdomain.Filter) ([]feedbackdomain.Feedback, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, filter)
}

func (f *fakeHTTPFeedbackRepository) ListPage(ctx context.Context, filter feedbackdomain.Filter, page, size int) ([]feedbackdomain.Feedback, int64, error) {
	if f.listPageFn == nil {
		return nil, 0, nil
	}
	return f.listPageFn(ctx, filter, page, size)
}

func TestListFeedbackRoute(t *testing.T) {
	now := time.Now().UTC()
	router := NewRouter(Handlers{Feedback: NewFeedbackHandler(feedbackapp.NewService(&fakeHTTPFeedbackRepository{
		listPageFn: func(_ context.Context, filter feedbackdomain.Filter, page, size int) ([]feedbackdomain.Feedback, int64, error) {
			if filter.UserID != "u-1" {
				t.Fatalf("expected filter user id u-1, got %q", filter.UserID)
			}
			if page != 1 || size != 20 {
				t.Fatalf("expected default page=1 size=20, got page=%d size=%d", page, size)
			}
			return []feedbackdomain.Feedback{
				{FeedbackType: "like", UserID: "u-1", ItemID: "i-1", Timestamp: now, Value: 1},
			}, 135, nil
		},
	}))})

	req := httptest.NewRequest(http.MethodGet, "/api/feedback?user_id=u-1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"feedback_type":"like"`) {
		t.Fatalf("expected response to contain feedback type, got %s", rec.Body.String())
	}
	if !strings.Contains(body, `"page":1`) || !strings.Contains(body, `"size":20`) || !strings.Contains(body, `"total":135`) {
		t.Fatalf("expected paged response metadata, got %s", body)
	}
}

func TestPatchFeedbackCollectionRoute(t *testing.T) {
	now := time.Now().UTC()
	router := NewRouter(Handlers{Feedback: NewFeedbackHandler(feedbackapp.NewService(&fakeHTTPFeedbackRepository{
		findFn: func(_ context.Context, feedbackType, userID, itemID string) (*feedbackdomain.Feedback, error) {
			return &feedbackdomain.Feedback{FeedbackType: feedbackType, UserID: userID, ItemID: itemID, Timestamp: now}, nil
		},
		updateBatchFn: func(_ context.Context, feedbacks []feedbackdomain.Feedback) error {
			if len(feedbacks) != 1 {
				t.Fatalf("expected 1 feedback, got %d", len(feedbacks))
			}
			if feedbacks[0].Value != 2 {
				t.Fatalf("expected merged value to be updated, got %v", feedbacks[0].Value)
			}
			return nil
		},
	}))})

	req := httptest.NewRequest(http.MethodPatch, "/api/feedback", strings.NewReader(`[{"feedback_type":"like","user_id":"u-1","item_id":"i-1","value":2}]`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"value":2`) {
		t.Fatalf("expected response to contain updated value, got %s", rec.Body.String())
	}
}

func TestDeleteFeedbackCollectionRoute(t *testing.T) {
	router := NewRouter(Handlers{Feedback: NewFeedbackHandler(feedbackapp.NewService(&fakeHTTPFeedbackRepository{
		deleteBatchFn: func(_ context.Context, keys []feedbackdomain.Key) error {
			if len(keys) != 1 {
				t.Fatalf("expected 1 key, got %d", len(keys))
			}
			if keys[0].FeedbackType != "like" || keys[0].UserID != "u-1" || keys[0].ItemID != "i-1" {
				t.Fatalf("unexpected key: %#v", keys[0])
			}
			return nil
		},
	}))})

	req := httptest.NewRequest(http.MethodDelete, "/api/feedback", strings.NewReader(`[{"feedback_type":"like","user_id":"u-1","item_id":"i-1"}]`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func TestListFeedbackRouteRejectsInvalidPagination(t *testing.T) {
	router := NewRouter(Handlers{Feedback: NewFeedbackHandler(feedbackapp.NewService(&fakeHTTPFeedbackRepository{}))})

	req := httptest.NewRequest(http.MethodGet, "/api/feedback?page=0&size=20", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
