package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"backend/core-server/internal/model/entity"
)

type CommentRepo struct {
	*DBClient
}

func NewCommentRepo(client *DBClient) *CommentRepo {
	return &CommentRepo{DBClient: client}
}

const commentSelectColumns = `
id, article_id, user_id, parent_id, root_id, reply_to_id, reply_to_name,
content, like_count, child_count, created_at, deleted_at`

func (r *CommentRepo) Create(ctx context.Context, comment *entity.Comment) (uint64, error) {
	now := time.Now()
	const query = `
INSERT INTO comment
    (article_id, user_id, parent_id, root_id, reply_to_id, reply_to_name,
     content, like_count, child_count, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db(ctx).ExecContext(ctx, query,
		comment.ArticleID, comment.UserID, comment.ParentID, comment.RootID,
		comment.ReplyToID, comment.ReplyToName, comment.Content,
		comment.LikeCount, comment.ChildCount, now,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *CommentRepo) SetRootID(ctx context.Context, id, rootID uint64) error {
	const query = `UPDATE comment SET root_id = ? WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db(ctx).ExecContext(ctx, query, rootID, id)
	return err
}

func (r *CommentRepo) GetByID(ctx context.Context, id uint64) (*entity.Comment, error) {
	var c entity.Comment
	query := fmt.Sprintf(`SELECT %s FROM comment WHERE id = ? AND deleted_at IS NULL LIMIT 1`, commentSelectColumns)
	err := r.db(ctx).GetContext(ctx, &c, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CommentRepo) SoftDelete(ctx context.Context, id, userID uint64) error {
	now := time.Now()
	const query = `
UPDATE comment
SET deleted_at = ?
WHERE id = ? AND user_id = ? AND deleted_at IS NULL`
	result, err := r.db(ctx).ExecContext(ctx, query, now, id, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CommentRepo) IncrementChildCount(ctx context.Context, rootID uint64) error {
	const query = `
UPDATE comment
SET child_count = child_count + 1
WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db(ctx).ExecContext(ctx, query, rootID)
	return err
}

func (r *CommentRepo) DecrementChildCount(ctx context.Context, rootID uint64) error {
	const query = `
UPDATE comment
SET child_count = CASE WHEN child_count > 0 THEN child_count - 1 ELSE 0 END
WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db(ctx).ExecContext(ctx, query, rootID)
	return err
}

func (r *CommentRepo) ListTopByArticle(ctx context.Context, articleID uint64, offset, limit int, orderBy string) ([]*entity.Comment, error) {
	orderClause := "ORDER BY created_at DESC"
	if strings.EqualFold(orderBy, "hot") {
		orderClause = "ORDER BY like_count DESC, created_at DESC"
	}

	var items []*entity.Comment
	query := fmt.Sprintf(`
SELECT %s FROM comment
WHERE article_id = ? AND parent_id = 0 AND deleted_at IS NULL
%s
LIMIT ?, ?`, commentSelectColumns, orderClause)

	if err := r.db(ctx).SelectContext(ctx, &items, query, articleID, offset, limit); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CommentRepo) CountTopByArticle(ctx context.Context, articleID uint64) (int, error) {
	var count int
	const query = `
SELECT COUNT(1) FROM comment
WHERE article_id = ? AND parent_id = 0 AND deleted_at IS NULL`
	if err := r.db(ctx).GetContext(ctx, &count, query, articleID); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *CommentRepo) ListRepliesByRoot(ctx context.Context, rootID uint64, offset, limit int) ([]*entity.Comment, error) {
	var items []*entity.Comment
	query := fmt.Sprintf(`
SELECT %s FROM comment
WHERE parent_id = ? AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT ?, ?`, commentSelectColumns)

	if err := r.db(ctx).SelectContext(ctx, &items, query, rootID, offset, limit); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CommentRepo) CountRepliesByRoot(ctx context.Context, rootID uint64) (int, error) {
	var count int
	const query = `
SELECT COUNT(1) FROM comment
WHERE parent_id = ? AND deleted_at IS NULL`
	if err := r.db(ctx).GetContext(ctx, &count, query, rootID); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *CommentRepo) ListRepliesByRoots(ctx context.Context, rootIDs []uint64) ([]*entity.Comment, error) {
	if len(rootIDs) == 0 {
		return nil, nil
	}

	baseQuery := fmt.Sprintf(`
SELECT %s FROM comment
WHERE parent_id IN (?) AND deleted_at IS NULL
ORDER BY parent_id, created_at ASC`, commentSelectColumns)

	query, args, err := sqlx.In(baseQuery, rootIDs)
	if err != nil {
		return nil, err
	}
	query = r.DB.Rebind(query)

	var items []*entity.Comment
	if err := r.db(ctx).SelectContext(ctx, &items, query, args...); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CommentRepo) ListByUser(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Comment, error) {
	var items []*entity.Comment
	query := fmt.Sprintf(`
SELECT %s FROM comment
WHERE user_id = ? AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT ?, ?`, commentSelectColumns)

	if err := r.db(ctx).SelectContext(ctx, &items, query, userID, offset, limit); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CommentRepo) CountByUser(ctx context.Context, userID uint64) (int, error) {
	var count int
	const query = `SELECT COUNT(1) FROM comment WHERE user_id = ? AND deleted_at IS NULL`
	if err := r.db(ctx).GetContext(ctx, &count, query, userID); err != nil {
		return 0, err
	}
	return count, nil
}
