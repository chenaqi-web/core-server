package domain

import (
	"context"
	"core-server/internal/model/entity"
)

type LikeRepo interface {
	Upsert(ctx context.Context, like *entity.InteractionLike) (int, error)
	UpdateWithCondition(ctx context.Context, condition string, like *entity.InteractionLike) (int, error)
	QueryWithCondition(ctx context.Context, userID, objectType, objectID, status string) (*entity.InteractionLike, error)
	CountUserLiked(ctx context.Context, userID, objectType string) (int64, error)
	PageQueryLikeObjects(ctx context.Context, userID, objectType string, offset, limit int) ([]*entity.InteractionLike, error)
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
	PageQueryObjects(ctx context.Context, userID, objectType string, page, size int) ([]string, error)
	SetLikeList(ctx context.Context, userID, objectType string, objectIDs []string, scores []float64) error
	SetUserThumbUpTotalCount(ctx context.Context, userID, objectType string, count int64) error
	QueryUserLikeTotalCount(ctx context.Context, userID, objectType string) (int64, error)
}
