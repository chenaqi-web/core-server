package domain

import (
	"context"
	"core-server/internal/model/entity"
)

type UserRepo interface {
	// 有关用户增删查的操作

	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	GetByName(ctx context.Context, name string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetStat(ctx context.Context, userID uint64) (*entity.UserStat, error)
	GetStats(ctx context.Context, userIDs []uint64) (map[uint64]*entity.UserStat, error)

	Create(ctx context.Context, user *entity.User) error
	CreateStat(ctx context.Context, user *entity.UserStat) error
	Search(ctx context.Context, keyword string, limit, offset uint32) ([]*entity.User, uint64, error)
	List(ctx context.Context, limit, offset uint32) ([]*entity.User, uint64, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error)

	// 有关用户互动方面的操作

	GetLikeCount(ctx context.Context, userID uint64) (int64, error)
	GetReceiveLikeCount(ctx context.Context, userID uint64) (int64, error)
	SetReceiveLikeCount(ctx context.Context, userID uint64, count int64) error
	IncrementLikeCount(ctx context.Context, userID uint64) error
	DecrementLikeCount(ctx context.Context, userID uint64) error

	// 有关用户更新方面的操作

	UpdateProfile(ctx context.Context, user *entity.User) error
	UpdateAvatar(ctx context.Context, userID uint64, avatar string) error
	UpdatePassword(ctx context.Context, userID uint64, password string) error
	UpdateStatus(ctx context.Context, userID uint64, status string) error
}

type UserRepoDomain interface {
	ITransaction
	UserRepo
}
