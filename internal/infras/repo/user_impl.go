package repo

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"

	"core-server/internal/model/entity"
)

const userListColumns = `id, created_at, updated_at, deleted_at, name, phone, avatar, email, role, sex, age, like_count, receive_like_count, status, auth_version`

type UserRepo struct {
	*DBClient
}

func NewUserRepo(client *DBClient) *UserRepo {
	return &UserRepo{DBClient: client}
}

func (r *UserRepo) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	var u entity.User
	const query = `
SELECT id, created_at, updated_at, deleted_at, name, phone, avatar, email, role, status, auth_version, sex, age, like_count, receive_like_count
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

func (r *UserRepo) GetByName(ctx context.Context, name string) (*entity.User, error) {
	var u entity.User
	const query = `
SELECT id, created_at, updated_at, deleted_at, name, password, phone, avatar, email, role, status, auth_version, sex, age
FROM user
WHERE name = ? AND deleted_at IS NULL
LIMIT 1`

	err := r.db(ctx).GetContext(ctx, &u, query, name)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var u entity.User
	const query = `
SELECT id, created_at, updated_at, deleted_at, name, password, phone, avatar, email, role, status, auth_version, sex, age
FROM user
WHERE email = ? AND deleted_at IS NULL
LIMIT 1`

	err := r.db(ctx).GetContext(ctx, &u, query, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	const query = `
INSERT INTO user (name, password, email, role, status)
VALUES (?, ?, ?, ?, ?)`
	result, err := r.db(ctx).ExecContext(ctx, query, user.Name, user.Password, user.Email, user.Role, user.Status)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = uint64(id)
	return nil
}

func (r *UserRepo) ListByIDs(ctx context.Context, ids []uint64) ([]*entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	const baseQuery = `
SELECT id, created_at, updated_at, deleted_at, name, phone, avatar, email, role, sex, age, like_count, receive_like_count, status, auth_version
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

func (r *UserRepo) List(ctx context.Context, keyword string, limit, offset uint32) ([]*entity.User, uint64, error) {
	keyword = strings.TrimSpace(keyword)
	where := "WHERE deleted_at IS NULL"
	args := make([]any, 0, 3)
	if keyword != "" {
		where += " AND (name LIKE ? OR email LIKE ?)"
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}

	var total uint64
	if err := r.db(ctx).GetContext(ctx, &total, "SELECT COUNT(*) FROM user "+where, args...); err != nil {
		return nil, 0, err
	}

	query := "SELECT " + userListColumns + " FROM user " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	var users []*entity.User
	if err := r.db(ctx).SelectContext(ctx, &users, query, args...); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepo) GetLikeCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	const query = `SELECT like_count FROM user WHERE id = ? AND deleted_at IS NULL`
	if err := r.db(ctx).GetContext(ctx, &count, query, userID); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepo) GetReceiveLikeCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	const query = `SELECT receive_like_count FROM user WHERE id = ? AND deleted_at IS NULL`
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

func (r *UserRepo) SetReceiveLikeCount(ctx context.Context, userID uint64, count int64) error {
	const query = `
UPDATE user
SET receive_like_count = ?, updated_at = NOW(3)
WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db(ctx).ExecContext(ctx, query, count, userID)
	return err
}

func (r *UserRepo) UpdateProfile(ctx context.Context, user *entity.User) error {
	const query = `
UPDATE user
SET name = ?, phone = ?, sex = ?, age = ?, updated_at = NOW(3)
WHERE id = ? AND deleted_at IS NULL`

	_, err := r.db(ctx).ExecContext(ctx, query, user.Name, user.Phone, user.Sex, user.Age, user.ID)
	return err
}

func (r *UserRepo) UpdateAvatar(ctx context.Context, userID uint64, avatar string) error {
	const query = `
UPDATE user
SET avatar = ?, updated_at = NOW(3)
WHERE id = ? AND deleted_at IS NULL`

	_, err := r.db(ctx).ExecContext(ctx, query, avatar, userID)
	return err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID uint64, hashedPassword string) error {
	const query = `
UPDATE user 
SET password = ?, updated_at = NOW() 
WHERE id = ?`

	result, err := r.db(ctx).ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepo) UpdateStatus(ctx context.Context, userID uint64, status string) error {
	result, err := r.db(ctx).ExecContext(ctx, `
UPDATE user
SET status = ?, auth_version = auth_version + 1, updated_at = NOW(3)
WHERE id = ? AND deleted_at IS NULL`, status, userID)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed == 0 {
		return sql.ErrNoRows
	}
	return nil
}
