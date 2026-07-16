package repo

import (
	"context"
	"errors"
	"fmt"

	"backend/core-server/internal/model/entity"

	"gorm.io/gorm"
)

type LikeRepo struct {
	*DBClient
}

func NewLikeRepo(client *DBClient) *LikeRepo {
	return &LikeRepo{DBClient: client}
}

// Upsert 原子写入点赞记录。
// 语义与原先一致：若已是 thumb_up 且 version >= 入参 version，则跳过；否则插入或更新。
// 依赖 uk_like_user_object(user_id, object_type, object_id)。
func (r *LikeRepo) Upsert(ctx context.Context, like *entity.InteractionLike) (int, error) {
	if like.ID == "" {
		like.ID = fmt.Sprintf("%s:%s:%s", like.UserID, like.ObjectType, like.ObjectID)
	}

	const sql = `
INSERT INTO interaction_like
  (id, user_id, object_type, object_id, object_owner_id, status, version, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  status = IF(@skip := (status = ? AND version >= VALUES(version)), status, VALUES(status)),
  version = IF(@skip, version, VALUES(version)),
  object_owner_id = IF(@skip, object_owner_id,
    IF(VALUES(object_owner_id) <> '', VALUES(object_owner_id), object_owner_id)),
  updated_at = IF(@skip, updated_at, NOW(3))
`

	result := r.db(ctx).Exec(
		sql,
		like.ID,
		like.UserID,
		like.ObjectType,
		like.ObjectID,
		like.ObjectOwnerID,
		like.Status,
		like.Version,
		entity.LikeStatusTypeThumbUp,
	)
	if result.Error != nil {
		return 0, result.Error
	}

	// MySQL: insert=1, update(有变更)=2, 无变更=0
	if result.RowsAffected == 0 {
		return 0, nil
	}
	return 1, nil
}

func (r *LikeRepo) UpdateWithCondition(ctx context.Context, condition string, like *entity.InteractionLike) (int, error) {
	result := r.db(ctx).Model(&entity.InteractionLike{}).
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
	err := r.db(ctx).Where(
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
