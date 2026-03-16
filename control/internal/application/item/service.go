// Package item 提供物品应用服务。
package item

import (
	"context"
	"errors"
	"strings"
	"time"

	itemdomain "github.com/kidyme/nexus/control/internal/domain/item"
)

var ErrInvalidPage = errors.New("item application: page must be greater than 0")
var ErrInvalidSize = errors.New("item application: size must be greater than 0")

// Service 编排物品用例。
type Service struct {
	repository itemdomain.Repository
}

// NewService 创建物品应用服务。
func NewService(repository itemdomain.Repository) *Service {
	return &Service{repository: repository}
}

// Create 创建物品。
func (s *Service) Create(ctx context.Context, item itemdomain.Item) error {
	return s.CreateBatch(ctx, []itemdomain.Item{item})
}

// CreateBatch 批量创建物品。
func (s *Service) CreateBatch(ctx context.Context, items []itemdomain.Item) error {
	for i := range items {
		items[i].ItemID = strings.TrimSpace(items[i].ItemID)
		if items[i].Timestamp.IsZero() {
			items[i].Timestamp = time.Now().UTC()
		}
		if err := items[i].Validate(); err != nil {
			return err
		}
	}
	return s.repository.CreateBatch(ctx, items)
}

// Update 更新物品。
func (s *Service) Update(ctx context.Context, item itemdomain.Item) error {
	return s.UpdateBatch(ctx, []itemdomain.Item{item})
}

// UpdateBatch 批量更新物品。
func (s *Service) UpdateBatch(ctx context.Context, items []itemdomain.Item) error {
	for i := range items {
		items[i].ItemID = strings.TrimSpace(items[i].ItemID)
		if items[i].Timestamp.IsZero() {
			items[i].Timestamp = time.Now().UTC()
		}
		if err := items[i].Validate(); err != nil {
			return err
		}
	}
	return s.repository.UpdateBatch(ctx, items)
}

// Delete 删除物品。
func (s *Service) Delete(ctx context.Context, itemID string) error {
	return s.DeleteBatch(ctx, []string{itemID})
}

// DeleteBatch 批量删除物品。
func (s *Service) DeleteBatch(ctx context.Context, itemIDs []string) error {
	for i := range itemIDs {
		itemIDs[i] = strings.TrimSpace(itemIDs[i])
		if itemIDs[i] == "" {
			return itemdomain.ErrItemIDRequired
		}
	}
	return s.repository.DeleteBatch(ctx, itemIDs)
}

// Find 查询单个物品。
func (s *Service) Find(ctx context.Context, itemID string) (*itemdomain.Item, error) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return nil, itemdomain.ErrItemIDRequired
	}
	return s.repository.Find(ctx, itemID)
}

// List 查询全部物品。
func (s *Service) List(ctx context.Context) ([]itemdomain.Item, error) {
	return s.repository.List(ctx)
}

// ListPage 分页查询物品。
func (s *Service) ListPage(ctx context.Context, page, size int) ([]itemdomain.Item, int64, error) {
	if page <= 0 {
		return nil, 0, ErrInvalidPage
	}
	if size <= 0 {
		return nil, 0, ErrInvalidSize
	}
	return s.repository.ListPage(ctx, page, size)
}
