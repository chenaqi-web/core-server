package domain

import (
	"context"
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"
)

type CountRepoDomain interface {
	Upsert(ctx context.Context, count *entity.InteractionCount, delta int64) error
	GetByObject(ctx context.Context, objectType enum.ObjectType, objectID uint64) ([]*entity.InteractionCount, error)
	GetByObjects(ctx context.Context, objectType enum.ObjectType, objectIDs []uint64) ([]*entity.InteractionCount, error)
}
