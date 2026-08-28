package rpc

import (
	"context"

	"core-server/internal/application"
	"core-server/internal/model/dto"
	"core-server/internal/rpc/categorypb"
)

type CategoryRPC struct {
	categorypb.UnimplementedCategoryServiceServer
	CategoryService *application.CategoryService
}

func NewCategoryRPC(categoryService *application.CategoryService) *CategoryRPC {
	return &CategoryRPC{CategoryService: categoryService}
}

func (c *CategoryRPC) CreateType(ctx context.Context, request *categorypb.CreateTypeRequest) (*categorypb.CreateTypeResponse, error) {
	if err := c.CategoryService.CreateType(ctx, &dto.CreateTypeRequest{Name: request.GetName()}); err != nil {
		return nil, err
	}
	return &categorypb.CreateTypeResponse{Success: true}, nil
}

func (c *CategoryRPC) DeleteType(ctx context.Context, request *categorypb.DeleteTypeRequest) (*categorypb.DeleteTypeResponse, error) {
	if err := c.CategoryService.DeleteType(ctx, &dto.DeleteTypeRequest{ID: request.GetId()}); err != nil {
		return nil, err
	}
	return &categorypb.DeleteTypeResponse{Success: true}, nil
}

func (c *CategoryRPC) ListTypes(ctx context.Context, _ *categorypb.ListTypesRequest) (*categorypb.ListTypesResponse, error) {
	types, err := c.CategoryService.ListTypes(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*categorypb.CategoryType, 0, len(types.Types))
	for _, category := range types.Types {
		items = append(items, &categorypb.CategoryType{Id: category.ID, Name: category.Name})
	}
	return &categorypb.ListTypesResponse{Types: items}, nil
}

// 上面 type
// =====================================================================================================================
// 下面 category

func (c *CategoryRPC) CreateCategory(ctx context.Context, request *categorypb.CreateCategoryRequest) (*categorypb.CreateCategoryResponse, error) {
	if err := c.CategoryService.CreateCategory(ctx, &dto.CreateCategoryRequest{ParentID: request.GetParentID(), Name: request.GetName()}); err != nil {
		return nil, err
	}
	return &categorypb.CreateCategoryResponse{Success: true}, nil
}

func (c *CategoryRPC) DeleteCategory(ctx context.Context, request *categorypb.DeleteCategoryRequest) (*categorypb.DeleteCategoryResponse, error) {
	if err := c.CategoryService.DeleteCategory(ctx, &dto.DeleteCategoryRequest{ID: request.GetId()}); err != nil {
		return nil, err
	}
	return &categorypb.DeleteCategoryResponse{Success: true}, nil
}

func (c *CategoryRPC) ListCategories(ctx context.Context, request *categorypb.ListCategoriesRequest) (*categorypb.ListCategoriesResponse, error) {
	categories, err := c.CategoryService.ListCategories(ctx, &dto.ListCategoriesRequest{ParentID: request.GetParentID()})
	if err != nil {
		return nil, err
	}

	items := make([]*categorypb.Category, 0, len(categories.Categories))
	for _, category := range categories.Categories {
		items = append(items, &categorypb.Category{Id: category.ID, ParentID: category.ParentID, Name: category.Name})
	}
	return &categorypb.ListCategoriesResponse{Categories: items}, nil
}
