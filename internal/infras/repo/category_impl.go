package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"backend/core-server/internal/model/entity"
)

type CategoryRepo struct {
	*DBClient
}

func NewCategoryRepo(client *DBClient) *CategoryRepo {
	return &CategoryRepo{DBClient: client}
}

func (r *CategoryRepo) Create(ctx context.Context, category *entity.Category) error {
	now := time.Now()
	const query = `
INSERT INTO category (parent_id, name, created_at, updated_at)
VALUES (?, ?, ?, ?)`

	result, err := r.db(ctx).ExecContext(ctx, query, category.ParentID, category.Name, now, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	category.ID = uint64(id)
	category.CreatedAt = now
	category.UpdatedAt = now
	return nil
}

func (r *CategoryRepo) DeleteByID(ctx context.Context, id uint64) error {
	const query = `
UPDATE category
SET deleted_at = NOW(3), updated_at = NOW(3)
WHERE id = ? AND deleted_at IS NULL`

	result, err := r.db(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, id uint64) (*entity.Category, error) {
	var category entity.Category
	const query = `
SELECT id, created_at, updated_at, deleted_at, parent_id, name
FROM category
WHERE id = ? AND deleted_at IS NULL
LIMIT 1`

	err := r.db(ctx).GetContext(ctx, &category, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepo) ListByParentID(ctx context.Context, parentID uint64) ([]*entity.Category, error) {
	var categories []*entity.Category
	const query = `
SELECT id, created_at, updated_at, deleted_at, parent_id, name
FROM category
WHERE parent_id = ? AND deleted_at IS NULL
ORDER BY id ASC`

	if err := r.db(ctx).SelectContext(ctx, &categories, query, parentID); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepo) DeleteByParentID(ctx context.Context, parentID uint64) error {
	const query = `
UPDATE category
SET deleted_at = NOW(3), updated_at = NOW(3)
WHERE parent_id = ? AND deleted_at IS NULL`

	_, err := r.db(ctx).ExecContext(ctx, query, parentID)
	return err
}
