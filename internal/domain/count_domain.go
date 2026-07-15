package domain

import (
	"backend/core-server/internal/model/entity"
	"context"
)

type CountRepoDomain interface {
	Upsert(ctx context.Context, count *entity.InteractionCount, delta int64) error
}
