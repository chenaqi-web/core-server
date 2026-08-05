package domain

import (
	"context"
	"core-server/internal/model/entity"
)

type UserRepo interface {
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error)
	GetLikeCount(ctx context.Context, userID uint64) (int64, error)
}

type UserRepoDomain interface {
	ITransaction
	UserRepo
}

type UserCacheDomain interface {
}
