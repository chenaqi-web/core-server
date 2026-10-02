package repo

import (
	"context"

	"core-server/internal/infras/repo/ent/interactioncount"
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"
)

type CountRepo struct{ *EntClient }

func NewCountRepo(client *EntClient) *CountRepo { return &CountRepo{EntClient: client} }

func (r *CountRepo) Upsert(ctx context.Context, count *entity.InteractionCount, delta int64) error {
	initialCount := delta
	if initialCount < 0 {
		initialCount = 0
	}
	const query = `
INSERT INTO interaction_count (object_type, object_id, interaction_type, count, created_at, updated_at)
VALUES (?, ?, ?, ?, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE count = GREATEST(count + ?, 0), updated_at = NOW(3)`
	return r.driver.Exec(ctx, query, []any{count.ObjectType.String(), count.ObjectID, count.InteractionType.String(), initialCount, delta}, nil)
}

func (r *CountRepo) GetByObject(ctx context.Context, objectType enum.ObjectType, objectID uint64) ([]*entity.InteractionCount, error) {
	nodes, err := r.DB(ctx).InteractionCount.Query().
		Where(interactioncount.ObjectTypeEQ(interactioncount.ObjectType(objectType.String())), interactioncount.ObjectIDEQ(objectID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityInteractionCounts(nodes), nil
}

func (r *CountRepo) GetByObjects(ctx context.Context, objectType enum.ObjectType, objectIDs []uint64) ([]*entity.InteractionCount, error) {
	if len(objectIDs) == 0 {
		return nil, nil
	}
	nodes, err := r.DB(ctx).InteractionCount.Query().
		Where(interactioncount.ObjectTypeEQ(interactioncount.ObjectType(objectType.String())), interactioncount.ObjectIDIn(objectIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityInteractionCounts(nodes), nil
}
