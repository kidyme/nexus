// Package recommendation 定义离线推荐领域模型。
package recommendation

import (
	"context"
	"time"

	userdomain "github.com/kidyme/nexus/offline/internal/domain/user"
)

// Candidate 表示召回候选。
type Candidate struct {
	ItemID  string    `json:"item_id"`
	Score   float64   `json:"score"`
	Source  string    `json:"source"`
	Created time.Time `json:"created_at"`
}

// Meta 表示用户推荐缓存元数据。
type Meta struct {
	UpdateTime time.Time
	Digest     string
}

// CacheRepository 定义推荐缓存读写接口。
type CacheRepository interface {
	GetLastModifyUserTime(ctx context.Context, userID string) (time.Time, error)
	GetMeta(ctx context.Context, userID string) (Meta, error)
	Save(ctx context.Context, userID string, items []Candidate, meta Meta) error
}

// SourceRepository 定义召回源数据读取接口。
type SourceRepository interface {
	ListPopularItems(ctx context.Context, feedbackTypes []string, limit int) ([]Candidate, error)
	ListLatestItems(ctx context.Context, limit int) ([]Candidate, error)
}

// Recaller 定义单个召回器策略。
type Recaller interface {
	Name() string
	Prepare(ctx context.Context) error
	Recall(ctx context.Context, user userdomain.User, limit int) ([]Candidate, error)
}
