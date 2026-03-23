// Package feedback 提供反馈应用服务。
package feedback

import (
	"context"
	"errors"
	"strings"
	"time"

	feedbackdomain "github.com/kidyme/nexus/control/internal/domain/feedback"
	refreshmetadomain "github.com/kidyme/nexus/control/internal/domain/refreshmeta"
)

var ErrInvalidPage = errors.New("feedback application: page must be greater than 0")
var ErrInvalidSize = errors.New("feedback application: size must be greater than 0")

// Service 编排反馈用例。
type Service struct {
	repository  feedbackdomain.Repository
	refreshMeta refreshmetadomain.Repository
}

// NewService 创建反馈应用服务。
func NewService(repository feedbackdomain.Repository, refreshMeta refreshmetadomain.Repository) *Service {
	return &Service{repository: repository, refreshMeta: refreshMeta}
}

// Create 创建反馈。
func (s *Service) Create(ctx context.Context, feedback feedbackdomain.Feedback) error {
	return s.CreateBatch(ctx, []feedbackdomain.Feedback{feedback})
}

// CreateBatch 批量创建反馈。
func (s *Service) CreateBatch(ctx context.Context, feedbacks []feedbackdomain.Feedback) error {
	for i := range feedbacks {
		feedbacks[i] = normalize(feedbacks[i])
		if feedbacks[i].Timestamp.IsZero() {
			feedbacks[i].Timestamp = time.Now().UTC()
		}
		if err := feedbacks[i].Validate(); err != nil {
			return err
		}
	}
	if err := s.repository.CreateBatch(ctx, feedbacks); err != nil {
		return err
	}
	return s.touchRelatedEntities(ctx, feedbacks)
}

// Update 更新反馈。
func (s *Service) Update(ctx context.Context, feedback feedbackdomain.Feedback) error {
	return s.UpdateBatch(ctx, []feedbackdomain.Feedback{feedback})
}

// UpdateBatch 批量更新反馈。
func (s *Service) UpdateBatch(ctx context.Context, feedbacks []feedbackdomain.Feedback) error {
	for i := range feedbacks {
		feedbacks[i] = normalize(feedbacks[i])
		if feedbacks[i].Timestamp.IsZero() {
			feedbacks[i].Timestamp = time.Now().UTC()
		}
		if err := feedbacks[i].Validate(); err != nil {
			return err
		}
	}
	if err := s.repository.UpdateBatch(ctx, feedbacks); err != nil {
		return err
	}
	return s.touchRelatedEntities(ctx, feedbacks)
}

// Delete 删除反馈。
func (s *Service) Delete(ctx context.Context, feedbackType, userID, itemID string) error {
	return s.DeleteBatch(ctx, []feedbackdomain.Key{{
		FeedbackType: feedbackType,
		UserID:       userID,
		ItemID:       itemID,
	}})
}

// DeleteBatch 批量删除反馈。
func (s *Service) DeleteBatch(ctx context.Context, keys []feedbackdomain.Key) error {
	for i := range keys {
		keys[i] = normalizeKey(keys[i])
		if err := keys[i].Validate(); err != nil {
			return err
		}
	}
	return s.repository.DeleteBatch(ctx, keys)
}

// Find 查询单个反馈。
func (s *Service) Find(ctx context.Context, feedbackType, userID, itemID string) (*feedbackdomain.Feedback, error) {
	feedback := normalize(feedbackdomain.Feedback{
		FeedbackType: feedbackType,
		UserID:       userID,
		ItemID:       itemID,
	})
	if err := feedback.Validate(); err != nil {
		return nil, err
	}
	return s.repository.Find(ctx, feedback.FeedbackType, feedback.UserID, feedback.ItemID)
}

// List 查询反馈列表。
func (s *Service) List(ctx context.Context, filter feedbackdomain.Filter) ([]feedbackdomain.Feedback, error) {
	filter.FeedbackType = strings.TrimSpace(filter.FeedbackType)
	filter.UserID = strings.TrimSpace(filter.UserID)
	filter.ItemID = strings.TrimSpace(filter.ItemID)
	return s.repository.List(ctx, filter)
}

// ListPage 分页查询反馈列表。
func (s *Service) ListPage(ctx context.Context, filter feedbackdomain.Filter, page, size int) ([]feedbackdomain.Feedback, int64, error) {
	filter.FeedbackType = strings.TrimSpace(filter.FeedbackType)
	filter.UserID = strings.TrimSpace(filter.UserID)
	filter.ItemID = strings.TrimSpace(filter.ItemID)
	if page <= 0 {
		return nil, 0, ErrInvalidPage
	}
	if size <= 0 {
		return nil, 0, ErrInvalidSize
	}
	return s.repository.ListPage(ctx, filter, page, size)
}

func normalize(feedback feedbackdomain.Feedback) feedbackdomain.Feedback {
	feedback.FeedbackType = strings.TrimSpace(feedback.FeedbackType)
	feedback.UserID = strings.TrimSpace(feedback.UserID)
	feedback.ItemID = strings.TrimSpace(feedback.ItemID)
	return feedback
}

func normalizeKey(key feedbackdomain.Key) feedbackdomain.Key {
	key.FeedbackType = strings.TrimSpace(key.FeedbackType)
	key.UserID = strings.TrimSpace(key.UserID)
	key.ItemID = strings.TrimSpace(key.ItemID)
	return key
}

func (s *Service) touchRelatedEntities(ctx context.Context, feedbacks []feedbackdomain.Feedback) error {
	userIDs := make([]string, 0, len(feedbacks))
	itemIDs := make([]string, 0, len(feedbacks))
	for _, feedback := range feedbacks {
		if feedback.UserID != "" {
			userIDs = append(userIDs, feedback.UserID)
		}
		if feedback.ItemID != "" {
			itemIDs = append(itemIDs, feedback.ItemID)
		}
	}
	if err := s.refreshMeta.TouchUsers(ctx, userIDs); err != nil {
		return err
	}
	return s.refreshMeta.TouchItems(ctx, itemIDs)
}
