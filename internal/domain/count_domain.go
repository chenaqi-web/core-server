package domain

import (
	"backend/core-server/internal/model/entity"
	"context"
)

type CountDomain interface {
	Upsert(ctx context.Context, count *entity.InteractionCount, delta int64) error
}
