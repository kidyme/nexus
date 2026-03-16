// Package item 提供物品仓储实现。
package item

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	itemdomain "github.com/kidyme/nexus/control/internal/domain/item"
)

// Repository 是基于 MySQL 的物品仓储实现。
type Repository struct {
	db *sql.DB
}

// NewRepository 创建物品仓储。
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建物品。
func (r *Repository) Create(ctx context.Context, item itemdomain.Item) error {
	return r.CreateBatch(ctx, []itemdomain.Item{item})
}

// CreateBatch 批量创建物品。
func (r *Repository) CreateBatch(ctx context.Context, items []itemdomain.Item) error {
	if len(items) == 0 {
		return nil
	}

	query := `INSERT INTO items (item_id, is_hidden, categories, timestamp, labels, comment) VALUES `
	args := make([]any, 0, len(items)*6)

	for i, item := range items {
		categories, err := json.Marshal(item.Categories)
		if err != nil {
			return err
		}

		if i > 0 {
			query += `, `
		}
		query += `(?, ?, ?, ?, ?, ?)`
		args = append(args, item.ItemID, item.IsHidden, nullableJSON(categories), item.Timestamp, nullableJSON(item.Labels), item.Comment)
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// Update 更新物品。
func (r *Repository) Update(ctx context.Context, item itemdomain.Item) error {
	return r.UpdateBatch(ctx, []itemdomain.Item{item})
}

// UpdateBatch 批量更新物品。
func (r *Repository) UpdateBatch(ctx context.Context, items []itemdomain.Item) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, item := range items {
		categories, err := json.Marshal(item.Categories)
		if err != nil {
			return err
		}

		result, err := tx.ExecContext(ctx,
			`UPDATE items SET is_hidden = ?, categories = ?, timestamp = ?, labels = ?, comment = ? WHERE item_id = ?`,
			item.IsHidden,
			nullableJSON(categories),
			item.Timestamp,
			nullableJSON(item.Labels),
			item.Comment,
			item.ItemID,
		)
		if err != nil {
			return err
		}
		if err := expectRows(result, itemdomain.ErrItemNotFound); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete 删除物品。
func (r *Repository) Delete(ctx context.Context, itemID string) error {
	return r.DeleteBatch(ctx, []string{itemID})
}

// DeleteBatch 批量删除物品。
func (r *Repository) DeleteBatch(ctx context.Context, itemIDs []string) error {
	if len(itemIDs) == 0 {
		return nil
	}

	itemIDs = uniqueStrings(itemIDs)
	query := `DELETE FROM items WHERE item_id IN (`
	args := make([]any, 0, len(itemIDs))

	for i, itemID := range itemIDs {
		if i > 0 {
			query += `, `
		}
		query += `?`
		args = append(args, itemID)
	}
	query += `)`

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected < int64(len(itemIDs)) {
		return itemdomain.ErrItemNotFound
	}

	return nil
}

// Find 查询单个物品。
func (r *Repository) Find(ctx context.Context, itemID string) (*itemdomain.Item, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT item_id, is_hidden, categories, timestamp, labels, comment, created_at, updated_at FROM items WHERE item_id = ?`,
		itemID,
	)

	var item itemdomain.Item
	var categories []byte
	var labels []byte
	if err := row.Scan(&item.ItemID, &item.IsHidden, &categories, &item.Timestamp, &labels, &item.Comment, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, itemdomain.ErrItemNotFound
		}
		return nil, err
	}
	if len(categories) > 0 {
		if err := json.Unmarshal(categories, &item.Categories); err != nil {
			return nil, err
		}
	}
	item.Labels = labels
	return &item, nil
}

// List 列出全部物品。
func (r *Repository) List(ctx context.Context) ([]itemdomain.Item, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT item_id, is_hidden, categories, timestamp, labels, comment, created_at, updated_at FROM items ORDER BY item_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []itemdomain.Item
	for rows.Next() {
		var item itemdomain.Item
		var categories []byte
		var labels []byte
		if err := rows.Scan(&item.ItemID, &item.IsHidden, &categories, &item.Timestamp, &labels, &item.Comment, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if len(categories) > 0 {
			if err := json.Unmarshal(categories, &item.Categories); err != nil {
				return nil, err
			}
		}
		item.Labels = labels
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListPage 分页列出物品。
func (r *Repository) ListPage(ctx context.Context, page, size int) ([]itemdomain.Item, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM items`).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := r.db.QueryContext(ctx,
		`SELECT item_id, is_hidden, categories, timestamp, labels, comment, created_at, updated_at FROM items ORDER BY item_id LIMIT ? OFFSET ?`,
		size,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []itemdomain.Item
	for rows.Next() {
		var item itemdomain.Item
		var categories []byte
		var labels []byte
		if err := rows.Scan(&item.ItemID, &item.IsHidden, &categories, &item.Timestamp, &labels, &item.Comment, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if len(categories) > 0 {
			if err := json.Unmarshal(categories, &item.Categories); err != nil {
				return nil, 0, err
			}
		}
		item.Labels = labels
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func nullableJSON(data []byte) any {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	return data
}

func expectRows(result sql.Result, notFound error) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return notFound
	}
	return nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
