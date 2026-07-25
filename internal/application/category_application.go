package application

import (
	"context"
	"errors"

	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/clog"
	"backend/core-server/internal/infras/repo"
	"backend/core-server/internal/model/entity"
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
// 一级类型

func (s *CategoryService) CreateType(ctx context.Context, name string) error {
	return s.repo.CreateType(ctx, &entity.CategoryType{Name: name})
}

func (s *CategoryService) DeleteType(ctx context.Context, id uint64) error {
	return s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.DeleteByTypeID(ctx, id); err != nil {
			return err
		}
		if err := s.repo.DeleteTypeByID(ctx, id); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return ErrCategoryTypeNotFound
			}
			return err
		}
		return nil
	})
}

func (s *CategoryService) ListTypes(ctx context.Context) ([]*entity.CategoryType, error) {
	return s.repo.ListTypes(ctx)
}

// =====================================================================================================================
// 二级分类

func (s *CategoryService) CreateCategory(ctx context.Context, typeID uint64, name string) error {
	categoryType, err := s.repo.QueryTypeByID(ctx, typeID)
	if err != nil {
		return err
	}
	if categoryType == nil {
		return ErrCategoryTypeNotFound
	}

	return s.repo.Create(ctx, &entity.Category{
		TypeID: typeID,
		Name:   name,
	})
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id uint64) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return nil
}

func (s *CategoryService) ListCategories(ctx context.Context, typeID uint64) ([]*entity.Category, error) {
	return s.repo.ListByTypeID(ctx, typeID)
}
