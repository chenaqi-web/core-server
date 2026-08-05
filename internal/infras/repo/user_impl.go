package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"core-server/internal/model/entity"
)

type UserRepo struct {
	*DBClient
}

func NewUserRepo(client *DBClient) *UserRepo {
	return &UserRepo{DBClient: client}
}

func (r *UserRepo) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	var u entity.User
	const query = `
SELECT id, created_at, updated_at, deleted_at, name, phone, avatar, email, role, sex, age, like_count
FROM user
WHERE id = ? AND deleted_at IS NULL
LIMIT 1`

	err := r.db(ctx).GetContext(ctx, &u, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	const baseQuery = `
SELECT id, created_at, updated_at, deleted_at, name, phone, avatar, email, role, sex, age, like_count
FROM user
WHERE id IN (?) AND deleted_at IS NULL`

	query, args, err := sqlx.In(baseQuery, ids)
	if err != nil {
		return nil, err
	}
	query = r.DB.Rebind(query)

	var users []*entity.User
	if err := r.db(ctx).SelectContext(ctx, &users, query, args...); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) GetLikeCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	const query = `SELECT like_count FROM user WHERE id = ? AND deleted_at IS NULL`
	if err := r.db(ctx).GetContext(ctx, &count, query, userID); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepo) IncrementLikeCount(ctx context.Context, userID uint64) error {
	const query = `
UPDATE user
SET like_count = like_count + 1, updated_at = NOW(3)
WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db(ctx).ExecContext(ctx, query, userID)
	return err
}

func (r *UserRepo) DecrementLikeCount(ctx context.Context, userID uint64) error {
	const query = `
UPDATE user
SET like_count = CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END,
    updated_at = NOW(3)
WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db(ctx).ExecContext(ctx, query, userID)
	return err
}
