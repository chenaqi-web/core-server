package domain

import (
	"context"

	"core-server/internal/model/entity"
)

type CategoryRepo interface {
	Create(ctx context.Context, category *entity.Category) error
	DeleteByID(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*entity.Category, error)
	ListByParentID(ctx context.Context, parentID uint64) ([]*entity.Category, error)
	DeleteByParentID(ctx context.Context, parentID uint64) error
}

type CategoryRepoDomain interface {
	ITransaction
	CategoryRepo
}
