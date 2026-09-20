package domain

import (
	"context"

	"core-server/internal/model/entity"
)

type CategoryRepo interface {
	CreateType(ctx context.Context, cate *entity.Category) error
	CreateCate(ctx context.Context, cate *entity.Category) error

	DeleteType(ctx context.Context, id uint64) error
	DeleteCate(ctx context.Context, id uint64) error

	ListType(ctx context.Context) ([]*entity.Category, error)
	ListCate(ctx context.Context, parentID uint64) ([]*entity.Category, error)
}

type CategoryRepoDomain interface {
	ITransaction
	CategoryRepo
}
