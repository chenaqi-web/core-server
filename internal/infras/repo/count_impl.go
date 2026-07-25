package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"backend/core-server/internal/model/entity"
)

type CountRepo struct {
	*DBClient
}

func NewCountRepo(client *DBClient) *CountRepo {
	return &CountRepo{DBClient: client}
}

func (r *CountRepo) Upsert(ctx context.Context, count *entity.InteractionCount, delta int64) error {
	const selectQuery = `
SELECT id, created_at, updated_at, object_type, object_id, interaction_type, count
FROM interaction_count
WHERE object_type = ? AND object_id = ? AND interaction_type = ?
LIMIT 1`

	var existing entity.InteractionCount
	err := r.db(ctx).GetContext(
		ctx,
		&existing,
		selectQuery,
		count.ObjectType,
		count.ObjectID,
		count.InteractionType,
	)
	if errors.Is(err, sql.ErrNoRows) {
		now := time.Now()
		if count.ID == "" {
			count.ID = fmt.Sprintf("%s:%s:%s", count.ObjectType, count.ObjectID, count.InteractionType)
		}
		count.Count = delta
		if count.Count < 0 {
			count.Count = 0
		}

		const insertQuery = `
INSERT INTO interaction_count
  (id, object_type, object_id, interaction_type, count, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?)`

		_, err := r.db(ctx).ExecContext(
			ctx,
			insertQuery,
			count.ID,
			count.ObjectType,
			count.ObjectID,
			count.InteractionType,
			count.Count,
			now,
			now,
		)
		return err
	}
	if err != nil {
		return err
	}

	newCount := existing.Count + delta
	if newCount < 0 {
		newCount = 0
	}

	const updateQuery = `
UPDATE interaction_count
SET count = ?, updated_at = NOW(3)
WHERE id = ?`

	_, err = r.db(ctx).ExecContext(ctx, updateQuery, newCount, existing.ID)
	return err
}
