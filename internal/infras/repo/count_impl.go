package repo

import (
	"context"
	"time"

	"core-server/internal/model/entity"
	"core-server/internal/model/enum"

	"github.com/jmoiron/sqlx"
)

type CountRepo struct {
	*DBClient
}

func NewCountRepo(client *DBClient) *CountRepo {
	return &CountRepo{DBClient: client}
}

func (r *CountRepo) Upsert(ctx context.Context, count *entity.InteractionCount, delta int64) error {
	initialCount := delta
	if initialCount < 0 {
		initialCount = 0
	}

	const query = `
INSERT INTO interaction_count
  (object_type, object_id, interaction_type, count, created_at, updated_at)
VALUES
	(?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  count = GREATEST(count + ?, 0),
	updated_at = ?`

	now := time.Now()
	_, err := r.db(ctx).ExecContext(
		ctx,
		query,
		count.ObjectType,
		count.ObjectID,
		count.InteractionType,
		initialCount,
		now,
		now,
		delta,
		now,
	)
	return err
}

func (r *CountRepo) GetByObject(ctx context.Context, objectType enum.ObjectType, objectID uint64) ([]*entity.InteractionCount, error) {
	var counts []*entity.InteractionCount
	const query = `
SELECT id, created_at, updated_at, object_type, object_id, interaction_type, count
FROM interaction_count
WHERE object_type = ? AND object_id = ?`

	if err := r.db(ctx).SelectContext(ctx, &counts, query, objectType, objectID); err != nil {
		return nil, err
	}
	return counts, nil
}

func (r *CountRepo) GetByObjects(ctx context.Context, objectType enum.ObjectType, objectIDs []uint64) ([]*entity.InteractionCount, error) {
	if len(objectIDs) == 0 {
		return nil, nil
	}

	const baseQuery = `
SELECT id, created_at, updated_at, object_type, object_id, interaction_type, count
FROM interaction_count
WHERE object_type = ? AND object_id IN (?)`

	query, args, err := sqlx.In(baseQuery, objectType, objectIDs)
	if err != nil {
		return nil, err
	}
	query = r.DB.Rebind(query)

	var counts []*entity.InteractionCount
	if err := r.db(ctx).SelectContext(ctx, &counts, query, args...); err != nil {
		return nil, err
	}
	return counts, nil
}
