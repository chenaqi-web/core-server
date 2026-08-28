package dto

import "core-server/internal/model/entity"

type CreateTypeRequest struct {
	Name string
}
type DeleteTypeRequest struct {
	ID uint64
}
type CreateCategoryRequest struct {
	ParentID uint64
	Name     string
}
type DeleteCategoryRequest struct {
	ID uint64
}
type ListCategoriesRequest struct {
	ParentID uint64
}

type CategoryType struct {
	ID   uint64
	Name string
}
type ListTypesResponse struct {
	Types []*CategoryType
}
type Category struct {
	ID       uint64
	ParentID uint64
	Name     string
}
type ListCategoriesResponse struct {
	Categories []*Category
}

func CategoryTypeFromEntity(c *entity.Category) *CategoryType {
	if c == nil {
		return nil
	}
	return &CategoryType{ID: c.ID, Name: c.Name}
}

func ToListTypesResponse(items []*entity.Category) *ListTypesResponse {
	r := &ListTypesResponse{Types: make([]*CategoryType, 0, len(items))}
	for _, item := range items {
		r.Types = append(r.Types, CategoryTypeFromEntity(item))
	}
	return r
}

func ToCategory(c *entity.Category) *Category {
	if c == nil {
		return nil
	}
	return &Category{ID: c.ID, ParentID: c.ParentID, Name: c.Name}
}

func ToListCategoriesResponse(items []*entity.Category) *ListCategoriesResponse {
	r := &ListCategoriesResponse{Categories: make([]*Category, 0, len(items))}
	for _, item := range items {
		r.Categories = append(r.Categories, ToCategory(item))
	}
	return r
}
