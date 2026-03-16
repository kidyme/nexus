// Package item 定义物品业务模型。
package item

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrItemIDRequired = errors.New("item domain: item id is required")
var ErrItemNotFound = errors.New("item domain: item not found")

// Item 表示可推荐物品。
type Item struct {
	ItemID     string          `json:"item_id"`
	Labels     json.RawMessage `json:"labels"`
	Categories []string        `json:"categories"`
	Comment    string          `json:"comment"`
	IsHidden   bool            `json:"is_hidden"`
	Timestamp  time.Time       `json:"timestamp"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// Validate 校验物品数据是否合法。
func (i Item) Validate() error {
	if strings.TrimSpace(i.ItemID) == "" {
		return ErrItemIDRequired
	}
	return nil
}

// Repository 定义物品数据访问接口。
type Repository interface {
	Create(ctx context.Context, item Item) error
	CreateBatch(ctx context.Context, items []Item) error
	Update(ctx context.Context, item Item) error
	UpdateBatch(ctx context.Context, items []Item) error
	Delete(ctx context.Context, itemID string) error
	DeleteBatch(ctx context.Context, itemIDs []string) error
	Find(ctx context.Context, itemID string) (*Item, error)
	List(ctx context.Context) ([]Item, error)
	ListPage(ctx context.Context, page, size int) ([]Item, int64, error)
}
