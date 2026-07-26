package domain

import (
	"backend/core-server/internal/model/entity"
	"context"
)

type UserRepo interface {
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error)
	FindByID(ctx context.Context, id uint64) (*entity.User, error)
	FindByName(ctx context.Context, name string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, user *entity.User) error
	UpdatePasswordAndIncrementAuthVersion(ctx context.Context, id uint64, passwordHash string) error
}

type UserRepoDomain interface {
	ITransaction
	UserRepo
}

type UserCacheDomain interface {
}
