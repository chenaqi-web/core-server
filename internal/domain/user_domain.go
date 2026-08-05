package domain

import (
	"context"
	"core-server/internal/model/entity"
)

type UserRepo interface {
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	GetByName(ctx context.Context, name string) (*entity.User, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error)
}

type UserRepoDomain interface {
	ITransaction
	UserRepo
}
