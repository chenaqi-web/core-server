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

type LikeCache interface{}

type LikeDomain interface {
	ITransaction
	LikeRepo
}
