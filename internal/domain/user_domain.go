package domain

import (
	"context"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"
)

type UserRepo interface {
	// 有关用户增删查的操作

	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	GetByName(ctx context.Context, name string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	CreateUser(ctx context.Context, user *entity.User) error
	GetUserMsgByID(ctx context.Context, id uint64) (*aggregate.UserAggregate, error)
	GetUserStat(ctx context.Context, userID uint64) (*entity.UserStat, error)

	Search(ctx context.Context, keyword string, limit, offset int32) ([]*entity.User, uint64, error)
	List(ctx context.Context, limit, offset int32) ([]*entity.User, uint64, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error)

	// 有关用户更新方面的操作

	UpdateProfile(ctx context.Context, user *entity.User) error
	UpdateAvatar(ctx context.Context, userID uint64, avatar string) error
	UpdatePassword(ctx context.Context, userID uint64, password string) error
	UpdateStatus(ctx context.Context, userID uint64, status enum.UserStatus) error

	// 个人主页方面的接口

	IncrementLikeCount(ctx context.Context, userID uint64) error
	DecrementLikeCount(ctx context.Context, userID uint64) error
	IncrementReceiveLikeCount(ctx context.Context, userID uint64) error
	DecrementReceiveLikeCount(ctx context.Context, userID uint64) error
	IncrementArticleCount(ctx context.Context, userID uint64) error
	DecrementArticleCount(ctx context.Context, userID uint64) error
}

type UserRepoDomain interface {
	ITransaction
	UserRepo
}
