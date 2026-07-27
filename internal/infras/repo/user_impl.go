package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"backend/core-server/internal/model/entity"
)

type UserRepo struct {
	*DBClient
}

const authUserColumns = `id, created_at, updated_at, deleted_at, name, password, phone, avatar, email, role, status, auth_version, sex, age`

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

func (r *UserRepo) ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	const baseQuery = `
SELECT id, created_at, updated_at, deleted_at, name, phone, avatar, email, role, sex, age
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

func (r *UserRepo) FindByID(ctx context.Context, id uint64) (*entity.User, error) {
	const query = `
SELECT ` + authUserColumns + `
FROM user
WHERE id = ? AND deleted_at IS NULL
LIMIT 1`
	return r.findOne(ctx, query, id)
}

func (r *UserRepo) FindByName(ctx context.Context, name string) (*entity.User, error) {
	const query = `
SELECT ` + authUserColumns + `
FROM user
WHERE name = ? AND deleted_at IS NULL
LIMIT 1`
	return r.findOne(ctx, query, name)
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	const query = `
SELECT ` + authUserColumns + `
FROM user
WHERE email = ? AND deleted_at IS NULL
LIMIT 1`
	return r.findOne(ctx, query, email)
}

func (r *UserRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	// Unique indexes also cover soft-deleted rows, so registration checks must do the same.
	const query = `SELECT EXISTS(SELECT 1 FROM user WHERE name = ? LIMIT 1)`
	return r.exists(ctx, query, name)
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	// Unique indexes also cover soft-deleted rows, so registration checks must do the same.
	const query = `SELECT EXISTS(SELECT 1 FROM user WHERE email = ? LIMIT 1)`
	return r.exists(ctx, query, email)
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	const query = `
INSERT INTO user
    (created_at, updated_at, name, password, phone, avatar, email, role, status, auth_version, sex, age)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db(ctx).ExecContext(
		ctx,
		query,
		now,
		now,
		user.Name,
		user.Password,
		user.Phone,
		user.Avatar,
		user.Email,
		user.Role,
		user.Status,
		user.AuthVersion,
		user.Sex,
		user.Age,
	)
	if err != nil {
		return mapUserWriteError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get created user ID: %w", err)
	}
	if id < 0 {
		return fmt.Errorf("created user ID is invalid")
	}
	user.ID = uint64(id)
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *UserRepo) UpdatePasswordAndIncrementAuthVersion(ctx context.Context, id uint64, passwordHash string) error {
	const query = `
UPDATE user
SET password = ?, auth_version = auth_version + 1, updated_at = ?
WHERE id = ? AND deleted_at IS NULL`
	result, err := r.db(ctx).ExecContext(ctx, query, passwordHash, time.Now(), id)
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

func (r *UserRepo) findOne(ctx context.Context, query string, argument any) (*entity.User, error) {
	var user entity.User
	if err := r.db(ctx).GetContext(ctx, &user, query, argument); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) exists(ctx context.Context, query string, argument any) (bool, error) {
	var exists bool
	if err := r.db(ctx).GetContext(ctx, &exists, query, argument); err != nil {
		return false, err
	}
	return exists, nil
}

func mapUserWriteError(err error) error {
	var mysqlError *mysql.MySQLError
	if !errors.As(err, &mysqlError) || mysqlError.Number != 1062 {
		return err
	}
	message := strings.ToLower(mysqlError.Message)
	switch {
	case strings.Contains(message, "uk_user_name"), strings.Contains(message, "uk_name"):
		return ErrUserNameExists
	case strings.Contains(message, "uk_user_email"), strings.Contains(message, "email"):
		return ErrUserEmailExists
	default:
		return err
	}
}
