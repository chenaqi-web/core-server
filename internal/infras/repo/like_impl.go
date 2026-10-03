package repo

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"core-server/internal/infras/repo/ent"
	"core-server/internal/infras/repo/ent/interactionlike"
	"core-server/internal/model/entity"
)

type LikeRepo struct{ *EntClient }

func NewLikeRepo(client *EntClient) *LikeRepo { return &LikeRepo{EntClient: client} }

func (r *LikeRepo) Upsert(ctx context.Context, like *entity.InteractionLike) (int, error) {
	const query = `
INSERT INTO interaction_like (user_id, object_type, object_id, status, version, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  status = IF(@skip := (status = ? AND version >= VALUES(version)), status, VALUES(status)),
  version = IF(@skip, version, VALUES(version)),
  updated_at = IF(@skip, updated_at, NOW(3))`
	driver, ok := r.driver.(interface {
		ExecContext(context.Context, string, ...any) (stdsql.Result, error)
	})
	if !ok {
		return 0, fmt.Errorf("ent driver does not support ExecContext")
	}
	result, err := driver.ExecContext(ctx, query, like.UserID, like.ObjectType.String(), like.ObjectID, like.Status.String(), like.Version, entity.LikeStatusTypeThumbUp.String())
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return 0, err
	}
	return 1, nil
}

func (r *LikeRepo) UpdateWithCondition(ctx context.Context, condition string, like *entity.InteractionLike) (int, error) {
	affected, err := r.DB(ctx).InteractionLike.Update().
		Where(
			interactionlike.UserIDEQ(like.UserID),
			interactionlike.ObjectTypeEQ(interactionlike.ObjectType(like.ObjectType.String())),
			interactionlike.ObjectIDEQ(like.ObjectID),
			interactionlike.StatusEQ(interactionlike.Status(condition)),
			interactionlike.VersionLTE(like.Version),
		).
		SetStatus(interactionlike.Status(like.Status.String())).
		SetVersion(like.Version).
		Save(ctx)
	return affected, err
}

func (r *LikeRepo) QueryWithCondition(ctx context.Context, userID uint64, objectType string, objectID uint64, status string) (*entity.InteractionLike, error) {
	node, err := r.DB(ctx).InteractionLike.Query().
		Where(interactionlike.UserIDEQ(userID), interactionlike.ObjectTypeEQ(interactionlike.ObjectType(objectType)), interactionlike.ObjectIDEQ(objectID), interactionlike.StatusEQ(interactionlike.Status(status))).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toEntityInteractionLike(node), nil
}

func (r *LikeRepo) CountUserLiked(ctx context.Context, userID uint64, objectType string) (int64, error) {
	count, err := r.DB(ctx).InteractionLike.Query().
		Where(interactionlike.UserIDEQ(userID), interactionlike.ObjectTypeEQ(interactionlike.ObjectType(objectType)), interactionlike.StatusEQ(interactionlike.StatusThumbUp)).
		Count(ctx)
	return int64(count), err
}

func (r *LikeRepo) PageQueryLikeObjects(ctx context.Context, userID uint64, objectType string, offset, limit int) ([]*entity.InteractionLike, error) {
	nodes, err := r.DB(ctx).InteractionLike.Query().
		Where(
			interactionlike.UserIDEQ(userID),
			interactionlike.ObjectTypeEQ(interactionlike.ObjectType(objectType)),
			interactionlike.StatusEQ(interactionlike.StatusThumbUp),
		).
		Order(ent.Desc(interactionlike.FieldVersion), ent.Desc(interactionlike.FieldUpdatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityInteractionLikes(nodes), nil
}
