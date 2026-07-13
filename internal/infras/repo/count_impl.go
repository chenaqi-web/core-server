package repo

import (
	"backend/core-server/internal/model/entity"
	"context"
	"errors"

	"gorm.io/gorm"
)

type CountRepo struct {
	client *DBClient
}

func NewCountRepo(client *DBClient) *CountRepo {
	return &CountRepo{client: client}
}

func (r *CountRepo) Upsert(ctx context.Context, count *entity.InteractionCount, delta int64) error {
	db := dbFromContext(ctx, r.client)

	var existing entity.InteractionCount
	err := db.Where(
		"object_type = ? AND object_id = ? AND interaction_type = ?",
		count.ObjectType, count.ObjectID, count.InteractionType,
	).Take(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		count.ID = newID()
		count.Count = delta
		if count.Count < 0 {
			count.Count = 0
		}
		return db.Create(count).Error
	}
	if err != nil {
		return err
	}

	newCount := existing.Count + delta
	if newCount < 0 {
		newCount = 0
	}
	return db.Model(&existing).Update("count", newCount).Error
}
