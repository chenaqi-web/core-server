package domain

import (
	"backend/core-server/internal/model/entity"
	"context"
)

type ITransaction interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type LikeRepo interface {
	Upsert(ctx context.Context, like *entity.InteractionLike) (int, error)
	UpdateWithCondition(ctx context.Context, condition string, like *entity.InteractionLike) (int, error)
	QueryWithCondition(ctx context.Context, userID, objectType, objectID, status string) (*entity.InteractionLike, error)
}

type LikeRepoDomain interface {
	ITransaction
	LikeRepo
}

// =====================================================================================================================

type LikeCacheDomain interface {
	CompensationCountDecr(ctx context.Context, objectID, objectType string) error
	CompensationCountIncr(ctx context.Context, objectID, objectType string) error

	ThumbUp(ctx context.Context, userID, objectType, objectID string, score int64) error
	CancelThumbUp(ctx context.Context, userID, objectType, objectID string) (int, int64, error)
	ExistZSetMember(ctx context.Context, userID, objectType, objectID string) (bool, error)
}
