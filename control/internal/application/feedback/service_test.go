package feedback

import (
	"context"
	"testing"

	feedbackdomain "github.com/kidyme/nexus/control/internal/domain/feedback"
)

type fakeFeedbackRepository struct {
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

func (f *fakeFeedbackRepository) Create(ctx context.Context, feedback feedbackdomain.Feedback) error {
	if f.createFn == nil {
		return nil
	}
	return f.createFn(ctx, feedback)
}

func (f *fakeFeedbackRepository) CreateBatch(ctx context.Context, feedbacks []feedbackdomain.Feedback) error {
	if f.createBatchFn == nil {
		return nil
	}
	return f.createBatchFn(ctx, feedbacks)
}

func (f *fakeFeedbackRepository) Update(ctx context.Context, feedback feedbackdomain.Feedback) error {
	if f.updateFn == nil {
		return nil
	}
	return f.updateFn(ctx, feedback)
}

func (f *fakeFeedbackRepository) UpdateBatch(ctx context.Context, feedbacks []feedbackdomain.Feedback) error {
	if f.updateBatchFn == nil {
		return nil
	}
	return f.updateBatchFn(ctx, feedbacks)
}

func (f *fakeFeedbackRepository) Delete(ctx context.Context, feedbackType, userID, itemID string) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, feedbackType, userID, itemID)
}

func (f *fakeFeedbackRepository) DeleteBatch(ctx context.Context, keys []feedbackdomain.Key) error {
	if f.deleteBatchFn == nil {
		return nil
	}
	return f.deleteBatchFn(ctx, keys)
}

func (f *fakeFeedbackRepository) Find(ctx context.Context, feedbackType, userID, itemID string) (*feedbackdomain.Feedback, error) {
	if f.findFn == nil {
		return nil, nil
	}
	return f.findFn(ctx, feedbackType, userID, itemID)
}

func (f *fakeFeedbackRepository) List(ctx context.Context, filter feedbackdomain.Filter) ([]feedbackdomain.Feedback, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, filter)
}

func (f *fakeFeedbackRepository) ListPage(ctx context.Context, filter feedbackdomain.Filter, page, size int) ([]feedbackdomain.Feedback, int64, error) {
	if f.listPageFn == nil {
		return nil, 0, nil
	}
	return f.listPageFn(ctx, filter, page, size)
}

func TestDeleteRequiresCompositeKey(t *testing.T) {
	service := NewService(&fakeFeedbackRepository{})

	err := service.Delete(context.Background(), "", "u-1", "i-1")
	if err != feedbackdomain.ErrFeedbackTypeRequired {
		t.Fatalf("expected ErrFeedbackTypeRequired, got %v", err)
	}
}

func TestDeleteBatchNormalizesCompositeKey(t *testing.T) {
	service := NewService(&fakeFeedbackRepository{
		deleteBatchFn: func(_ context.Context, keys []feedbackdomain.Key) error {
			if len(keys) != 1 {
				t.Fatalf("expected 1 key, got %d", len(keys))
			}
			if keys[0].FeedbackType != "like" || keys[0].UserID != "u-1" || keys[0].ItemID != "i-1" {
				t.Fatalf("expected normalized key, got %#v", keys[0])
			}
			return nil
		},
	})

	if err := service.DeleteBatch(context.Background(), []feedbackdomain.Key{{
		FeedbackType: " like ",
		UserID:       " u-1 ",
		ItemID:       " i-1 ",
	}}); err != nil {
		t.Fatalf("delete batch feedback: %v", err)
	}
}

func TestListPageValidatesPagination(t *testing.T) {
	service := NewService(&fakeFeedbackRepository{})

	if _, _, err := service.ListPage(context.Background(), feedbackdomain.Filter{}, 0, 20); err != ErrInvalidPage {
		t.Fatalf("expected ErrInvalidPage, got %v", err)
	}
	if _, _, err := service.ListPage(context.Background(), feedbackdomain.Filter{}, 1, 0); err != ErrInvalidSize {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}
