package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"backend/core-server/internal/model/entity"
)

type ArticleRepo struct {
	*DBClient
}

func NewArticleRepo(client *DBClient) *ArticleRepo {
	return &ArticleRepo{DBClient: client}
}

func (r *ArticleRepo) Create(ctx context.Context, article *entity.Article) error {
	now := time.Now()
	const query = `
                INSERT INTO blog_article 
                    (title, summary, content, cover_image, author_id,
                    category_id, is_top, view_count, like_count, comment_count,
                    created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db(ctx).ExecContext(ctx, query,
		article.Title, article.Summary, article.Content, article.CoverImage,
		article.AuthorID, article.CategoryID, article.IsTop,
		article.ViewCount, article.LikeCount, article.CommentCount,
		now, now,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *ArticleRepo) DeleteByID(ctx context.Context, id, authorID uint64) error {
	now := time.Now()
	query := `
UPDATE blog_article 
SET deleted_at = ?, updated_at = ? 
WHERE id = ? 
  AND deleted_at IS NULL
  AND author_id = ?`
	result, err := r.db(ctx).ExecContext(ctx, query, now, now, id, authorID)
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

func (r *ArticleRepo) GetByID(ctx context.Context, id uint64) (*entity.Article, error) {
	var a entity.Article
	const query = `
SELECT id, created_at, updated_at, deleted_at, title, summary, content, cover_image, author_id, category_id, is_top, view_count, like_count, comment_count
FROM blog_article
WHERE id = ? AND deleted_at IS NULL
LIMIT 1`

	err := r.db(ctx).GetContext(ctx, &a, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ArticleRepo) List(ctx context.Context, offset, limit int) ([]*entity.Article, error) {
	var items []*entity.Article
	const query = `
SELECT id, created_at, updated_at, deleted_at, title, summary, content, cover_image, author_id, category_id, is_top, view_count, like_count, comment_count
FROM blog_article
WHERE deleted_at IS NULL
ORDER BY id DESC
LIMIT ?, ?`

	if err := r.db(ctx).SelectContext(ctx, &items, query, offset, limit); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ArticleRepo) ListByAuthor(ctx context.Context, authorID uint64, offset, limit int) ([]*entity.Article, error) {
	var items []*entity.Article
	const query = `
SELECT id, created_at, updated_at, deleted_at, title, summary, content, cover_image, author_id, category_id, is_top, view_count, like_count, comment_count
FROM blog_article
WHERE author_id = ? AND deleted_at IS NULL
ORDER BY id DESC
LIMIT ?, ?`

	if err := r.db(ctx).SelectContext(ctx, &items, query, authorID, offset, limit); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ArticleRepo) ListByCategory(ctx context.Context, categoryID uint64, offset, limit int) ([]*entity.Article, error) {
	var items []*entity.Article
	const query = `
SELECT id, created_at, updated_at, deleted_at, title, summary, content, cover_image, author_id, category_id, is_top, view_count, like_count, comment_count
FROM blog_article
WHERE category_id = ? AND deleted_at IS NULL
ORDER BY id DESC
LIMIT ?, ?`

	if err := r.db(ctx).SelectContext(ctx, &items, query, categoryID, offset, limit); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ArticleRepo) Search(ctx context.Context, name string, offset, limit int) ([]*entity.Article, error) {
	var items []*entity.Article
	const query = `
SELECT id, created_at, updated_at, deleted_at, title, summary, content, cover_image, author_id, category_id, is_top, view_count, like_count, comment_count
FROM blog_article
WHERE (title LIKE ? OR summary LIKE ? OR content LIKE ?) AND deleted_at IS NULL
ORDER BY id DESC
LIMIT ?, ?`

	like := "%" + name + "%"
	if err := r.db(ctx).SelectContext(ctx, &items, query, like, like, like, offset, limit); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ArticleRepo) UpdateCommentCount(ctx context.Context, articleID uint64, delta int64) error {
	const query = `
UPDATE blog_article
SET comment_count = comment_count + ?, updated_at = ?
WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db(ctx).ExecContext(ctx, query, delta, time.Now(), articleID)
	return err
}
