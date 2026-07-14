package repo

import (
	"context"
	"errors"

	"backend/core-server/internal/model/entity"

	"gorm.io/gorm"
)

type LikeRepo struct {
	*DBClient
}

func NewLikeRepo(client *DBClient) *LikeRepo {
	return &LikeRepo{DBClient: client}
}

func (r *LikeRepo) Upsert(ctx context.Context, like *entity.InteractionLike) (int, error) {
	var existing entity.InteractionLike
	err := r.DB.Where(
		"user_id = ? AND object_type = ? AND object_id = ?",
		like.UserID, like.ObjectType, like.ObjectID,
	).Take(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := r.DB.Create(like).Error; err != nil {
			return 0, err
		}
		return 1, nil
	}
	if err != nil {
		return 0, err
	}

	if existing.Status == entity.LikeStatusTypeThumbUp && existing.Version >= like.Version {
		return 0, nil
	}

	updates := map[string]interface{}{
		"status":  like.Status,
		"version": like.Version,
	}
	if like.ObjectOwnerID != "" {
		updates["object_owner_id"] = like.ObjectOwnerID
	}
	result := r.DB.Model(&existing).Updates(updates)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, nil
	}
	return 1, nil
}

func (r *LikeRepo) UpdateWithCondition(ctx context.Context, condition string, like *entity.InteractionLike) (int, error) {
	result := r.DB.Model(&entity.InteractionLike{}).
		Where(
			"user_id = ? AND object_type = ? AND object_id = ? AND status = ? AND version <= ?",
			like.UserID, like.ObjectType, like.ObjectID, condition, like.Version,
		).
		Updates(map[string]interface{}{
			"status":  like.Status,
			"version": like.Version,
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

func (r *LikeRepo) QueryWithCondition(ctx context.Context, userID, objectType, objectID, status string) (*entity.InteractionLike, error) {
	var like entity.InteractionLike
	err := r.DB.Where(
		"user_id = ? AND object_type = ? AND object_id = ? AND status = ?",
		userID, objectType, objectID, status,
	).Take(&like).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &like, nil
}
