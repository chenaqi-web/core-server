package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"backend/core-server/internal/model/entity"

	"gorm.io/gorm"
)

type txContextKey struct{}

func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func dbFromContext(ctx context.Context, client *DBClient) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return client.GetDB().WithContext(ctx)
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type LikeRepo struct {
	client *DBClient
}

func NewLikeRepo(client *DBClient) *LikeRepo {
	return &LikeRepo{client: client}
}

func (r *LikeRepo) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.client.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(withTx(ctx, tx))
	})
}

func (r *LikeRepo) Upsert(ctx context.Context, like *entity.InteractionLike) (int, error) {
	db := dbFromContext(ctx, r.client)

	var existing entity.InteractionLike
	err := db.Where(
		"user_id = ? AND object_type = ? AND object_id = ?",
		like.UserID, like.ObjectType, like.ObjectID,
	).Take(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		like.ID = newID()
		if err := db.Create(like).Error; err != nil {
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
	result := db.Model(&existing).Updates(updates)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, nil
	}
	return 1, nil
}

func (r *LikeRepo) UpdateWithCondition(ctx context.Context, condition string, like *entity.InteractionLike) (int, error) {
	db := dbFromContext(ctx, r.client)

	result := db.Model(&entity.InteractionLike{}).
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
	db := dbFromContext(ctx, r.client)

	var like entity.InteractionLike
	err := db.Where(
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
