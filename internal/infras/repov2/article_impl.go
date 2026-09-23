package repov2

import (
	"context"
	"core-server/internal/infras/repov2/ent/user"
	"errors"
	"time"

	"core-server/internal/infras/repov2/ent"
	"core-server/internal/infras/repov2/ent/article"
	"core-server/internal/model/entity"
)

type ArticleRepo struct {
	*EntClient
}

func NewArticleRepo(client *EntClient) *ArticleRepo {
	return &ArticleRepo{
		EntClient: client,
	}
}

// Create 创建一篇文章
func (r *ArticleRepo) Create(ctx context.Context, value *entity.Article) error {
	create := r.DB(ctx).Article.Create().
		SetTitle(value.Title).
		SetSummary(value.Summary).
		SetContent(value.Content).
		SetCoverImage(value.CoverImage).
		SetAuthorID(value.AuthorID).
		SetCategoryID(value.CategoryID).
		SetIsTop(value.IsTop).
		SetIsPublished(value.IsPublished).
		SetVisibility(int(value.Visibility))
	// 判断作者是否存在用户表里面
	exists, _ := r.db.User.Query().
		Where(user.IDEQ(value.AuthorID)).
		Exist(ctx)
	if !exists {
		return errors.New("user not found")
	}

	// 如果用户选择发表文章 设置发布时间为当前时间
	if value.PublishedAt.Valid {
		create.SetPublishedAt(time.Now())
	}

	_, err := create.Save(ctx)
	return err
}

// Edit 编辑文章 需要做权限校验，无法编辑他人的文章
func (r *ArticleRepo) Edit(ctx context.Context, value *entity.Article) error {
	update := r.DB(ctx).Article.Update().
		SetTitle(value.Title).
		SetSummary(value.Summary).
		SetContent(value.Content).
		SetCoverImage(value.CoverImage).
		SetAuthorID(value.AuthorID).
		SetCategoryID(value.CategoryID).
		SetIsTop(value.IsTop).
		SetVisibility(int(value.Visibility)).
		Where(article.IDEQ(value.ID), article.DeletedAtIsNil())

	_, err := update.Save(ctx)
	return err
}

// DeleteByID 软删除文章；authorID 为 0 时表示管理员删除，不再校验作者
func (r *ArticleRepo) DeleteByID(ctx context.Context, id, authorID uint64) error {
	now := time.Now()
	update := r.DB(ctx).Article.Update().
		Where(article.IDEQ(id), article.DeletedAtIsNil()).
		SetDeletedAt(now)

	// 非管理员删除时，额外校验文章归属
	if authorID != 0 {
		update = update.Where(article.AuthorIDEQ(authorID))
	}

	affected, err := update.Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetByID 根据 ID 查询单篇未删除且已发布的文章（前台详情页，过滤草稿）
func (r *ArticleRepo) GetByID(ctx context.Context, id uint64) (*entity.Article, error) {
	node, err := r.DB(ctx).Article.Query().
		Where(article.IDEQ(id), article.DeletedAtIsNil(), article.IsPublishedEQ(true)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toEntityArticle(node), nil
}

// ListByIDs 根据 ID 批量查询未删除的文章
func (r *ArticleRepo) ListByIDs(ctx context.Context, ids []uint64) ([]*entity.Article, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	nodes, err := r.DB(ctx).Article.Query().
		Where(article.IDIn(ids...), article.DeletedAtIsNil(), article.IsPublishedEQ(true)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticles(nodes), nil
}

// List 分页查询全部已发布文章，按 ID 倒序
func (r *ArticleRepo) List(ctx context.Context, page, pagesize int) ([]*entity.Article, error) {
	nodes, err := r.DB(ctx).Article.Query().
		Where(article.DeletedAtIsNil(), article.IsPublishedEQ(true), article.VisibilityEQ(1)).
		Order(ent.Desc(article.FieldID)).
		Offset(page).
		Limit(pagesize).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticles(nodes), nil
}

// ListByAuthor 分页查询某个作者的文章（无法查看草稿和私密文章） 按 ID 倒序
func (r *ArticleRepo) ListByAuthor(ctx context.Context, authorID uint64, offset, limit int) ([]*entity.Article, error) {
	nodes, err := r.DB(ctx).Article.Query().
		Where(article.AuthorIDEQ(authorID), article.DeletedAtIsNil(), article.IsPublishedEQ(true), article.VisibilityEQ(1)).
		Order(ent.Desc(article.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticles(nodes), nil
}

// ListMe 查看自己的全部文章 (含草稿和私密文章)
func (r *ArticleRepo) ListMe(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Article, error) {
	nodes, err := r.DB(ctx).Article.Query().
		Where(article.AuthorIDEQ(userID), article.DeletedAtIsNil()).
		Order(ent.Desc(article.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticles(nodes), nil
}

// ListByCategory 分页查询某个分类下的已发布文章，按 ID 倒序
func (r *ArticleRepo) ListByCategory(ctx context.Context, categoryID uint64, offset, limit int) ([]*entity.Article, error) {
	nodes, err := r.DB(ctx).Article.Query().
		Where(article.CategoryIDEQ(categoryID), article.DeletedAtIsNil(), article.IsPublishedEQ(true)).
		Order(ent.Desc(article.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticles(nodes), nil
}

// Search 按标题、摘要、正文模糊搜索已发布文章，按 ID 倒序
func (r *ArticleRepo) Search(ctx context.Context, name string, offset, limit int) ([]*entity.Article, error) {
	nodes, err := r.DB(ctx).Article.Query().
		Where(
			article.DeletedAtIsNil(),
			article.IsPublishedEQ(true),
			article.Or(
				article.TitleContains(name),
				article.SummaryContains(name),
				article.ContentContains(name),
			),
		).
		Order(ent.Desc(article.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticles(nodes), nil
}
