package domain

import (
	"backend/core-server/internal/model/entity"
	"context"
)

// 数据库操作的接口

type LikeDomain interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Upsert(ctx context.Context, like *entity.InteractionLike) (int, error)
	UpdateWithCondition(ctx context.Context, condition string, like *entity.InteractionLike) (int, error)
	QueryWithCondition(ctx context.Context, userID, objectType, objectID, status string) (*entity.InteractionLike, error)
}
