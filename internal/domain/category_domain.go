package domain

import (
	"context"

	"backend/core-server/internal/model/entity"
)

type CategoryRepo interface {
	CreateType(ctx context.Context, categoryType *entity.CategoryType) error
	DeleteTypeByID(ctx context.Context, id uint64) error
	QueryTypeByID(ctx context.Context, id uint64) (*entity.CategoryType, error)
	ListTypes(ctx context.Context) ([]*entity.CategoryType, error)

	Create(ctx context.Context, category *entity.Category) error
	DeleteByID(ctx context.Context, id uint64) error
	DeleteByTypeID(ctx context.Context, typeID uint64) error
	ListByTypeID(ctx context.Context, typeID uint64) ([]*entity.Category, error)
}

type CategoryRepoDomain interface {
	ITransaction
	CategoryRepo
}
