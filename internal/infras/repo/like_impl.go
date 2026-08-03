package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"core-server/internal/model/entity"
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

	const sqlQuery = `
INSERT INTO interaction_like
  (id, user_id, object_type, object_id, status, version, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  status = IF(@skip := (status = ? AND version >= VALUES(version)), status, VALUES(status)),
  version = IF(@skip, version, VALUES(version)),
  updated_at = IF(@skip, updated_at, NOW(3))
`

	result, err := r.db(ctx).ExecContext(
		ctx,
		sqlQuery,
		like.ID,
		like.UserID,
		like.ObjectType,
		like.ObjectID,
		like.Status,
		like.Version,
		entity.LikeStatusTypeThumbUp,
	)
	if err != nil {
		return 0, err
	}

	// MySQL: insert=1, update(有变更)=2, 无变更=0
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		return 0, nil
	}
	return 1, nil
}

func (r *LikeRepo) UpdateWithCondition(ctx context.Context, condition string, like *entity.InteractionLike) (int, error) {
	const sqlQuery = `
UPDATE interaction_like
SET status = ?, version = ?, updated_at = NOW(3)
WHERE user_id = ? AND object_type = ? AND object_id = ? AND status = ? AND version <= ?`

	result, err := r.db(ctx).ExecContext(
		ctx,
		sqlQuery,
		like.Status,
		like.Version,
		like.UserID,
		like.ObjectType,
		like.ObjectID,
		condition,
		like.Version,
	)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rowsAffected), nil
}

func (r *LikeRepo) QueryWithCondition(ctx context.Context, userID, objectType, objectID, status string) (*entity.InteractionLike, error) {
	var like entity.InteractionLike
	const sqlQuery = `
SELECT id, created_at, updated_at, user_id, object_type, object_id, status, version
FROM interaction_like
WHERE user_id = ? AND object_type = ? AND object_id = ? AND status = ?
LIMIT 1`

	err := r.db(ctx).GetContext(ctx, &like, sqlQuery, userID, objectType, objectID, status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &like, nil
}
