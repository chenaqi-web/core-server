package domain

import (
	"context"
	"core-server/internal/model/entity"
)

type LikeRepo interface {
	Upsert(ctx context.Context, like *entity.InteractionLike) (int, error)
	UpdateWithCondition(ctx context.Context, condition string, like *entity.InteractionLike) (int, error)
	QueryWithCondition(ctx context.Context, userID uint64, objectType string, objectID uint64, status string) (*entity.InteractionLike, error)
	CountUserLiked(ctx context.Context, userID uint64, objectType string) (int64, error)
	PageQueryLikeObjects(ctx context.Context, userID uint64, objectType string, offset, limit int) ([]*entity.InteractionLike, error)
}

type LikeRepoDomain interface {
	ITransaction
	LikeRepo
}

// =====================================================================================================================

type LikeCacheDomain interface {
	CompensationCountDecr(ctx context.Context, objectID uint64, objectType string) error
	CompensationCountIncr(ctx context.Context, objectID uint64, objectType string) error
	GetObjectLikeCount(ctx context.Context, objectID uint64, objectType string) (uint64, error)

	ThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64, score int64) error
	CancelThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) (int, int64, error)
	ExistZSetMember(ctx context.Context, userID uint64, objectType string, objectID uint64) (bool, error)
	PageQueryObjects(ctx context.Context, userID uint64, objectType string, page, size int) ([]uint64, error)
	SetLikeList(ctx context.Context, userID uint64, objectType string, objectIDs []uint64, scores []float64) error
}
