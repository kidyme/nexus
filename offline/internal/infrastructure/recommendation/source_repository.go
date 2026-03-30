// Package recommendation 提供召回源数据访问实现。
package recommendation

import (
	"context"
	"database/sql"
	"strings"

	recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"
	"github.com/kidyme/nexus/offline/internal/recallkey"
)

// SourceRepository 是基于 MySQL 的召回源数据仓储实现。
type SourceRepository struct {
	db *sql.DB
}

// NewSourceRepository 创建召回源数据仓储。
func NewSourceRepository(db *sql.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

// ListPopularItems 返回热门物品。
func (r *SourceRepository) ListPopularItems(ctx context.Context, feedbackTypes []string, limit int) ([]recdomain.Candidate, error) {
	query := `SELECT item_id, COUNT(*) AS score FROM feedback`
	args := make([]any, 0, len(feedbackTypes)+1)
	if len(feedbackTypes) > 0 {
		query += ` WHERE feedback_type IN (` + placeholders(len(feedbackTypes)) + `)`
		for _, feedbackType := range feedbackTypes {
			args = append(args, feedbackType)
		}
	}
	query += ` GROUP BY item_id ORDER BY score DESC, item_id LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]recdomain.Candidate, 0, limit)
	for rows.Next() {
		var candidate recdomain.Candidate
		if err := rows.Scan(&candidate.ItemID, &candidate.Score); err != nil {
			return nil, err
		}
		candidate.Source = recallkey.RecallerPopular
		result = append(result, candidate)
	}
	return result, rows.Err()
}

// ListLatestItems 返回最新物品。
func (r *SourceRepository) ListLatestItems(ctx context.Context, limit int) ([]recdomain.Candidate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT item_id, timestamp FROM items WHERE is_hidden = FALSE ORDER BY timestamp DESC, item_id LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]recdomain.Candidate, 0, limit)
	for rows.Next() {
		var candidate recdomain.Candidate
		var itemTime sql.NullTime
		if err := rows.Scan(&candidate.ItemID, &itemTime); err != nil {
			return nil, err
		}
		if itemTime.Valid {
			candidate.Score = float64(itemTime.Time.UTC().Unix())
		}
		candidate.Source = recallkey.RecallerLatest
		result = append(result, candidate)
	}
	return result, rows.Err()
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = "?"
	}
	return strings.Join(parts, ", ")
}
