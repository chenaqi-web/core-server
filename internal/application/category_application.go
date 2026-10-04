package application

import (
	"context"
	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/model/dto"
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

func (s *CategoryService) CreateType(ctx context.Context, req *dto.CreateTypeRequest) error {
	// 有唯一索引
	if err := s.repo.CreateType(ctx, &entity.Category{
		ParentID: entity.RootCategoryParentID,
		Name:     req.Name,
	}); err != nil {
		s.log.Error("CreateType Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *CategoryService) DeleteType(ctx context.Context, req *dto.DeleteTypeRequest) error {
	if err := s.repo.DeleteType(ctx, req.ID); err != nil {
		s.log.Error("DeleteType Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *CategoryService) ListTypes(ctx context.Context) (*dto.ListTypesResponse, error) {
	res, err := s.repo.ListType(ctx)
	if err != nil {
		s.log.Error("ListTypes Error", zap.Error(err))
		return nil, err
	}
	return dto.ToListTypesResponse(res), nil
}

// =====================================================================================================================
// 子分类

func (s *CategoryService) CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) error {
	if err := s.repo.CreateCate(ctx, &entity.Category{
		ParentID: req.ParentID,
		Name:     req.Name,
	}); err != nil {
		s.log.Error("CreateCategory Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, req *dto.DeleteCategoryRequest) error {
	if err := s.repo.DeleteCate(ctx, req.ID); err != nil {
		s.log.Error("DeleteCategory Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *CategoryService) ListCategories(ctx context.Context, req *dto.ListCategoriesRequest) (*dto.ListCategoriesResponse, error) {
	res, err := s.repo.ListCate(ctx, req.ParentID)
	if err != nil {
		s.log.Error("ListCategories Error", zap.Error(err))
		return nil, err
	}
	return dto.ToListCategoriesResponse(res), nil
}
