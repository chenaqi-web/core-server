package domain

import (
	"context"
	"core-server/internal/model/entity"
)

type CountRepoDomain interface {
	Upsert(ctx context.Context, count *entity.InteractionCount, delta int64) error
}
