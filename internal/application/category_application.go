package application

import (
	"context"
	"errors"

	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/infras/repo"
	"core-server/internal/model/entity"

	"go.uber.org/zap"
)

type CategoryService struct {
	cfg  *config.Config
	log  *clog.Log
	repo domain.CategoryRepoDomain
}

func NewCategoryService(
	log *clog.Log,
	repo domain.CategoryRepoDomain,
	cfg *config.Config,
) (*CategoryService, error) {
	return &CategoryService{
		cfg:  cfg,
		log:  log,
		repo: repo,
	}, nil
}

// =====================================================================================================================
// 一级类型（parent_id = 0）

func (s *CategoryService) CreateType(ctx context.Context, name string) error {
	if err := s.repo.Create(ctx, &entity.Category{
		ParentID: entity.RootCategoryParentID,
		Name:     name,
	}); err != nil {
		s.log.Error("CreateType Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *CategoryService) DeleteType(ctx context.Context, id uint64) error {
	return s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.DeleteByParentID(ctx, id); err != nil {
			s.log.Error("DeleteType Error", zap.Error(err))
			return err
		}
		if err := s.repo.DeleteByID(ctx, id); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return ErrCategoryTypeNotFound
			}
			s.log.Error("DeleteType Error", zap.Error(err))
			return err
		}
		return nil
	})
}

func (s *CategoryService) ListTypes(ctx context.Context) ([]*entity.Category, error) {
	res, err := s.repo.ListByParentID(ctx, entity.RootCategoryParentID)
	if err != nil {
		s.log.Error("ListTypes Error", zap.Error(err))
		return nil, err
	}
	return res, nil
}

// =====================================================================================================================
// 子分类

func (s *CategoryService) CreateCategory(ctx context.Context, parentID uint64, name string) error {
	if err := s.repo.Create(ctx, &entity.Category{
		ParentID: parentID,
		Name:     name,
	}); err != nil {
		s.log.Error("CreateCategory Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id uint64) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrCategoryNotFound
		}
		s.log.Error("DeleteCategory Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *CategoryService) ListCategories(ctx context.Context, parentID uint64) ([]*entity.Category, error) {
	res, err := s.repo.ListByParentID(ctx, parentID)
	if err != nil {
		s.log.Error("ListCategories Error", zap.Error(err))
		return nil, err
	}
	return res, nil
}
