package rpc

import (
	"context"

	"backend/core-server/internal/application"
	"backend/core-server/internal/model/entity"
	"backend/core-server/internal/rpc/categorypb"
)

type CategoryRPC struct {
	categorypb.UnimplementedCategoryServiceServer
	CategoryService *application.CategoryService
}

func NewCategoryRPC(categoryService *application.CategoryService) *CategoryRPC {
	return &CategoryRPC{CategoryService: categoryService}
}

func (c *CategoryRPC) CreateType(ctx context.Context, request *categorypb.CreateTypeRequest) (*categorypb.CreateTypeResponse, error) {
	if err := c.CategoryService.CreateType(ctx, request.GetName()); err != nil {
		return nil, err
	}
	return &categorypb.CreateTypeResponse{Success: true}, nil
}

func (c *CategoryRPC) DeleteType(ctx context.Context, request *categorypb.DeleteTypeRequest) (*categorypb.DeleteTypeResponse, error) {
	if err := c.CategoryService.DeleteType(ctx, request.GetId()); err != nil {
		return nil, err
	}
	return &categorypb.DeleteTypeResponse{Success: true}, nil
}

func (c *CategoryRPC) ListTypes(ctx context.Context, _ *categorypb.ListTypesRequest) (*categorypb.ListTypesResponse, error) {
	types, err := c.CategoryService.ListTypes(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*categorypb.CategoryType, 0, len(types))
	for _, categoryType := range types {
		items = append(items, toCategoryTypePB(categoryType))
	}
	return &categorypb.ListTypesResponse{Types: items}, nil
}

// 上面type
// =====================================================================================================================
// 下面category

func (c *CategoryRPC) CreateCategory(ctx context.Context, request *categorypb.CreateCategoryRequest) (*categorypb.CreateCategoryResponse, error) {
	if err := c.CategoryService.CreateCategory(ctx, request.GetTypeID(), request.GetName()); err != nil {
		return nil, err
	}
	return &categorypb.CreateCategoryResponse{Success: true}, nil
}

func (c *CategoryRPC) DeleteCategory(ctx context.Context, request *categorypb.DeleteCategoryRequest) (*categorypb.DeleteCategoryResponse, error) {
	if err := c.CategoryService.DeleteCategory(ctx, request.GetId()); err != nil {
		return nil, err
	}
	return &categorypb.DeleteCategoryResponse{Success: true}, nil
}

func (c *CategoryRPC) ListCategories(ctx context.Context, request *categorypb.ListCategoriesRequest) (*categorypb.ListCategoriesResponse, error) {
	categories, err := c.CategoryService.ListCategories(ctx, request.GetTypeID())
	if err != nil {
		return nil, err
	}

	items := make([]*categorypb.Category, 0, len(categories))
	for _, category := range categories {
		items = append(items, toCategoryPB(category))
	}
	return &categorypb.ListCategoriesResponse{Categories: items}, nil
}

func toCategoryTypePB(categoryType *entity.CategoryType) *categorypb.CategoryType {
	return &categorypb.CategoryType{
		Id:   categoryType.ID,
		Name: categoryType.Name,
	}
}

func toCategoryPB(category *entity.Category) *categorypb.Category {
	return &categorypb.Category{
		Id:     category.ID,
		TypeID: category.TypeID,
		Name:   category.Name,
	}
}
