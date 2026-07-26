package repo

import (
	"context"
	"database/sql"
	"errors"

	"backend/core-server/internal/model/entity"
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
SELECT id, created_at, updated_at, deleted_at, name, phone, avatar, email, role, sex, age
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
