// Package user 提供用户应用服务。
package user

import (
	"context"
	"errors"
	"strings"

	refreshmetadomain "github.com/kidyme/nexus/control/internal/domain/refreshmeta"
	userdomain "github.com/kidyme/nexus/control/internal/domain/user"
)

var ErrInvalidPage = errors.New("user application: page must be greater than 0")
var ErrInvalidSize = errors.New("user application: size must be greater than 0")

// Service 编排用户用例。
type Service struct {
	repository  userdomain.Repository
	refreshMeta refreshmetadomain.Repository
}

// NewService 创建用户应用服务。
func NewService(repository userdomain.Repository, refreshMeta refreshmetadomain.Repository) *Service {
	return &Service{repository: repository, refreshMeta: refreshMeta}
}

// Create 创建用户。
func (s *Service) Create(ctx context.Context, user userdomain.User) error {
	return s.CreateBatch(ctx, []userdomain.User{user})
}

// CreateBatch 批量创建用户。
func (s *Service) CreateBatch(ctx context.Context, users []userdomain.User) error {
	for i := range users {
		users[i].UserID = strings.TrimSpace(users[i].UserID)
		if err := users[i].Validate(); err != nil {
			return err
		}
	}
	if err := s.repository.CreateBatch(ctx, users); err != nil {
		return err
	}
	return s.refreshMeta.TouchUsers(ctx, collectUserIDs(users))
}

// Update 更新用户。
func (s *Service) Update(ctx context.Context, user userdomain.User) error {
	return s.UpdateBatch(ctx, []userdomain.User{user})
}

// UpdateBatch 批量更新用户。
func (s *Service) UpdateBatch(ctx context.Context, users []userdomain.User) error {
	for i := range users {
		users[i].UserID = strings.TrimSpace(users[i].UserID)
		if err := users[i].Validate(); err != nil {
			return err
		}
	}
	if err := s.repository.UpdateBatch(ctx, users); err != nil {
		return err
	}
	return s.refreshMeta.TouchUsers(ctx, collectUserIDs(users))
}

// Delete 删除用户。
func (s *Service) Delete(ctx context.Context, userID string) error {
	return s.DeleteBatch(ctx, []string{userID})
}

// DeleteBatch 批量删除用户。
func (s *Service) DeleteBatch(ctx context.Context, userIDs []string) error {
	for i := range userIDs {
		userIDs[i] = strings.TrimSpace(userIDs[i])
		if userIDs[i] == "" {
			return userdomain.ErrUserIDRequired
		}
	}
	return s.repository.DeleteBatch(ctx, userIDs)
}

// Find 查询单个用户。
func (s *Service) Find(ctx context.Context, userID string) (*userdomain.User, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, userdomain.ErrUserIDRequired
	}
	return s.repository.Find(ctx, userID)
}

// List 查询全部用户。
func (s *Service) List(ctx context.Context) ([]userdomain.User, error) {
	return s.repository.List(ctx)
}

// ListPage 分页查询用户。
func (s *Service) ListPage(ctx context.Context, page, size int) ([]userdomain.User, int64, error) {
	if page <= 0 {
		return nil, 0, ErrInvalidPage
	}
	if size <= 0 {
		return nil, 0, ErrInvalidSize
	}
	return s.repository.ListPage(ctx, page, size)
}

func collectUserIDs(users []userdomain.User) []string {
	result := make([]string, 0, len(users))
	for _, user := range users {
		if user.UserID != "" {
			result = append(result, user.UserID)
		}
	}
	return result
}
