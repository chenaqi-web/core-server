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

func (r *CategoryRepo) CreateType(ctx context.Context, categoryType *entity.CategoryType) error {
	now := time.Now()
	const query = `
INSERT INTO category_type (name, created_at, updated_at)
VALUES (?, ?, ?)`

	result, err := r.db(ctx).ExecContext(ctx, query, categoryType.Name, now, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	categoryType.ID = uint64(id)
	categoryType.CreatedAt = now
	categoryType.UpdatedAt = now
	return nil
}

func (r *CategoryRepo) DeleteTypeByID(ctx context.Context, id uint64) error {
	const query = `
UPDATE category_type
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

func (r *CategoryRepo) QueryTypeByID(ctx context.Context, id uint64) (*entity.CategoryType, error) {
	var categoryType entity.CategoryType
	const query = `
SELECT id, created_at, updated_at, deleted_at, name
FROM category_type
WHERE id = ? AND deleted_at IS NULL
LIMIT 1`

	err := r.db(ctx).GetContext(ctx, &categoryType, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &categoryType, nil
}

func (r *CategoryRepo) ListTypes(ctx context.Context) ([]*entity.CategoryType, error) {
	var types []*entity.CategoryType
	const query = `
SELECT id, created_at, updated_at, deleted_at, name
FROM category_type
WHERE deleted_at IS NULL
ORDER BY id ASC`

	if err := r.db(ctx).SelectContext(ctx, &types, query); err != nil {
		return nil, err
	}
	return types, nil
}

// =====================================================================================================================

func (r *CategoryRepo) Create(ctx context.Context, category *entity.Category) error {
	now := time.Now()
	const query = `
INSERT INTO category (type_id, name, created_at, updated_at)
VALUES (?, ?, ?, ?)`

	result, err := r.db(ctx).ExecContext(ctx, query, category.TypeID, category.Name, now, now)
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

func (r *CategoryRepo) DeleteByTypeID(ctx context.Context, typeID uint64) error {
	const query = `
UPDATE category
SET deleted_at = NOW(3), updated_at = NOW(3)
WHERE type_id = ? AND deleted_at IS NULL`

	_, err := r.db(ctx).ExecContext(ctx, query, typeID)
	return err
}

func (r *CategoryRepo) ListByTypeID(ctx context.Context, typeID uint64) ([]*entity.Category, error) {
	var categories []*entity.Category
	const query = `
SELECT id, created_at, updated_at, deleted_at, type_id, name
FROM category
WHERE type_id = ? AND deleted_at IS NULL
ORDER BY id ASC`

	if err := r.db(ctx).SelectContext(ctx, &categories, query, typeID); err != nil {
		return nil, err
	}
	return categories, nil
}
