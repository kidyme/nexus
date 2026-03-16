// Package feedback 定义反馈业务模型。
package feedback

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrFeedbackTypeRequired = errors.New("feedback domain: feedback type is required")
var ErrUserIDRequired = errors.New("feedback domain: user id is required")
var ErrItemIDRequired = errors.New("feedback domain: item id is required")
var ErrFeedbackNotFound = errors.New("feedback domain: feedback not found")

// Feedback 表示用户与物品之间的交互事件。
type Feedback struct {
	FeedbackType string    `json:"feedback_type"`
	UserID       string    `json:"user_id"`
	ItemID       string    `json:"item_id"`
	Value        float64   `json:"value"`
	Timestamp    time.Time `json:"timestamp"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Key 表示反馈记录的复合主键。
type Key struct {
	FeedbackType string `json:"feedback_type"`
	UserID       string `json:"user_id"`
	ItemID       string `json:"item_id"`
}

// Filter 定义反馈列表过滤条件。
type Filter struct {
	FeedbackType string
	UserID       string
	ItemID       string
}

// Validate 校验反馈数据是否合法。
func (f Feedback) Validate() error {
	switch {
	case strings.TrimSpace(f.FeedbackType) == "":
		return ErrFeedbackTypeRequired
	case strings.TrimSpace(f.UserID) == "":
		return ErrUserIDRequired
	case strings.TrimSpace(f.ItemID) == "":
		return ErrItemIDRequired
	}
	return nil
}

// Validate 校验反馈主键是否合法。
func (k Key) Validate() error {
	return Feedback{
		FeedbackType: k.FeedbackType,
		UserID:       k.UserID,
		ItemID:       k.ItemID,
	}.Validate()
}

// Repository 定义反馈数据访问接口。
type Repository interface {
	Create(ctx context.Context, feedback Feedback) error
	CreateBatch(ctx context.Context, feedbacks []Feedback) error
	Update(ctx context.Context, feedback Feedback) error
	UpdateBatch(ctx context.Context, feedbacks []Feedback) error
	Delete(ctx context.Context, feedbackType, userID, itemID string) error
	DeleteBatch(ctx context.Context, keys []Key) error
	Find(ctx context.Context, feedbackType, userID, itemID string) (*Feedback, error)
	List(ctx context.Context, filter Filter) ([]Feedback, error)
	ListPage(ctx context.Context, filter Filter, page, size int) ([]Feedback, int64, error)
}
