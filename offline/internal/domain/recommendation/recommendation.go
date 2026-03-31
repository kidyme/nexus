// Package recommendation 定义离线推荐领域模型。
package recommendation

import (
	"context"
	"encoding/json"
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

// ItemDocument 表示用于召回构建的物品特征快照。
type ItemDocument struct {
	ItemID     string
	Categories []string
	Labels     json.RawMessage
	Timestamp  time.Time
}

// UserItemFeedback 表示用户与物品的正反馈关联。
type UserItemFeedback struct {
	UserID string
	ItemID string
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

// PreparedModelCacheRepository 定义离线预计算模型缓存接口。
type PreparedModelCacheRepository interface {
	GetItemToItemNeighbors(ctx context.Context, key string) (map[string][]Candidate, string, error)
	SaveItemToItemNeighbors(ctx context.Context, key string, neighbors map[string][]Candidate, digest string) error
}

// SourceRepository 定义召回源数据读取接口。
type SourceRepository interface {
	ListPopularItems(ctx context.Context, feedbackTypes []string, limit int) ([]Candidate, error)
	ListLatestItems(ctx context.Context, limit int) ([]Candidate, error)
	ListItemDocuments(ctx context.Context) ([]ItemDocument, error)
	ListPositiveUserItemFeedback(ctx context.Context, feedbackTypes []string) ([]UserItemFeedback, error)
	GetItemToItemDigest(ctx context.Context, feedbackTypes []string) (string, error)
}

// Recaller 定义单个召回器策略。
type Recaller interface {
	Name() string
	Prepare(ctx context.Context) error
	Recall(ctx context.Context, user userdomain.User, limit int) ([]Candidate, error)
}
