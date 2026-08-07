package domain

import (
	"context"
	"core-server/internal/model/entity"
)

type UserRepo interface {
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	GetByName(ctx context.Context, name string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error)

	GetLikeCount(ctx context.Context, userID uint64) (int64, error)
	GetReceiveLikeCount(ctx context.Context, userID uint64) (int64, error)
	SetReceiveLikeCount(ctx context.Context, userID uint64, count int64) error
	IncrementLikeCount(ctx context.Context, userID uint64) error
	DecrementLikeCount(ctx context.Context, userID uint64) error
}

type UserRepoDomain interface {
	ITransaction
	UserRepo
}
