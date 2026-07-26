package domain

import (
	"backend/core-server/internal/model/entity"
	"context"
)

type UserRepo interface {
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error)
}

type UserRepoDomain interface {
	ITransaction
	UserRepo
}

type UserCacheDomain interface {
}
