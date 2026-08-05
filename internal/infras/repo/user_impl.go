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
SELECT id, created_at, updated_at, deleted_at, name, phone, avatar, email, role, status, auth_version, sex, age
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

// 根据用户名查询用户
func (r *UserRepo) GetByName(ctx context.Context, name string) (*entity.User, error) {
	var u entity.User
	const query = `
SELECT id, created_at, updated_at, deleted_at, name, password, phone, avatar, email, role, status, auth_version, sex, age
FROM user
WHERE name = ? AND deleted_at IS NULL
LIMIT 1`

	err := r.db(ctx).GetContext(ctx, &u, query, name)
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
SELECT id, created_at, updated_at, deleted_at, name, phone, avatar, email, role, status, auth_version, sex, age
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
