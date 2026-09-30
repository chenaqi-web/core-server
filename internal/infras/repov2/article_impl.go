package repov2

import (
	"context"
	"core-server/internal/infras/repov2/ent/user"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/enum"
	"errors"
	"time"

	"core-server/internal/infras/repov2/ent"
	"core-server/internal/infras/repov2/ent/article"
	"core-server/internal/infras/repov2/ent/predicate"
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
	// 判断作者是否存在用户表里面
	exists, _ := r.db.User.Query().
		Where(user.IDEQ(value.AuthorID)).
		Exist(ctx)
	if !exists {
		return errors.New("user not found")
	}

	create := r.DB(ctx).Article.Create().
		SetTitle(value.Title).
		SetSummary(value.Summary).
		SetContent(value.Content).
		SetCoverImage(value.CoverImage).
		SetAuthorID(value.AuthorID).
		SetCategoryID(value.CategoryID).
		SetIsTop(value.IsTop).
		SetIsPublished(value.IsPublished)
	if value.PublishedAt.Valid {
		create.SetPublishedAt(value.PublishedAt.Time)
	}

	_, err := create.Save(ctx)

	if err != nil {
		return err
	}

	return nil
}

// Edit 编辑文章 需要做权限校验，无法编辑他人的文章
func (r *ArticleRepo) Edit(ctx context.Context, value *entity.Article) error {
	update := r.DB(ctx).Article.Update().
		SetTitle(value.Title).
		SetSummary(value.Summary).
		SetContent(value.Content).
		SetCoverImage(value.CoverImage).
		SetCategoryID(value.CategoryID).
		SetIsPublished(value.IsPublished).
		SetIsTop(value.IsTop).
		Where(article.IDEQ(value.ID), article.AuthorIDEQ(value.AuthorID), article.DeletedAtIsNil())
	if value.IsPublished {
		update.SetPublishedAt(time.Now())
	} else {
		update.ClearPublishedAt()
	}

	affected, err := update.Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return err
}

func (r *ArticleRepo) PublishDraft(ctx context.Context, id, authorID uint64) error {
	affected, err := r.DB(ctx).Article.Update().
		Where(article.IDEQ(id), article.AuthorIDEQ(authorID), article.IsPublishedEQ(false), article.DeletedAtIsNil()).
		SetIsPublished(true).
		SetPublishedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ArticleRepo) DeleteByID(ctx context.Context, id, authorID uint64, role string) (*aggregate.ArticleAggregate, error) {
	now := time.Now()
	predicates := []predicate.Article{
		article.IDEQ(id),
		article.IsPublishedEQ(true),
		article.DeletedAtIsNil(),
	}
	if role != enum.UserRoleAdmin.String() {
		predicates = append(predicates, article.AuthorIDEQ(authorID))
	}

	node, err := r.DB(ctx).Article.Query().
		Where(predicates...).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	affected, err := r.DB(ctx).Article.Update().
		Where(predicates...).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	return toEntityArticleAggregate(node), nil
}

func (r *ArticleRepo) DeleteDraftByID(ctx context.Context, id, authorID uint64) error {
	now := time.Now()
	affected, err := r.DB(ctx).Article.Update().
		Where(article.IDEQ(id), article.AuthorIDEQ(authorID), article.IsPublishedEQ(false), article.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetByID 根据 ID 查询单篇未删除且已发布的文章（前台详情页，过滤草稿）
func (r *ArticleRepo) GetByID(ctx context.Context, id uint64) (*aggregate.ArticleAggregate, error) {
	node, err := r.DB(ctx).Article.Query().
		Where(article.IDEQ(id), article.DeletedAtIsNil(), article.IsPublishedEQ(true)).
		WithCategory().
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toEntityArticleAggregate(node), nil
}

func (r *ArticleRepo) GetByIDForManage(ctx context.Context, id uint64) (*aggregate.ArticleAggregate, error) {
	node, err := r.DB(ctx).Article.Query().
		Where(article.IDEQ(id), article.DeletedAtIsNil()).
		WithCategory().
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toEntityArticleAggregate(node), nil
}

// ListByIDs 根据 ID 批量查询未删除的文章
func (r *ArticleRepo) ListByIDs(ctx context.Context, ids []uint64) ([]*aggregate.ArticleAggregate, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	nodes, err := r.DB(ctx).Article.Query().
		Where(article.IDIn(ids...), article.DeletedAtIsNil(), article.IsPublishedEQ(true)).
		WithCategory().
		WithUser().
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticleAggregates(nodes), nil
}

// List 分页查询全部已发布文章，按 ID 倒序
func (r *ArticleRepo) List(ctx context.Context, page, pageSize int) ([]*aggregate.ArticleAggregate, error) {
	nodes, err := r.DB(ctx).Article.Query().
		Where(article.DeletedAtIsNil(), article.IsPublishedEQ(true)).
		WithCategory().
		WithUser().
		Order(ent.Desc(article.FieldID)).
		Offset(page).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticleAggregates(nodes), nil
}

// ListByAuthor 分页查询某个作者的文章，可按发布状态筛选，按 ID 倒序
func (r *ArticleRepo) ListByAuthor(ctx context.Context, authorID uint64, offset, limit int, isPublished *bool) ([]*aggregate.ArticleAggregate, error) {
	predicates := []predicate.Article{
		article.AuthorIDEQ(authorID),
		article.DeletedAtIsNil(),
	}
	if isPublished != nil {
		predicates = append(predicates, article.IsPublishedEQ(*isPublished))
	}

	nodes, err := r.DB(ctx).Article.Query().
		Where(predicates...).
		WithCategory().
		WithUser().
		Order(ent.Desc(article.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticleAggregates(nodes), nil
}

func (r *ArticleRepo) CountByAuthor(ctx context.Context, authorID uint64, isPublished *bool) (uint64, error) {
	predicates := []predicate.Article{
		article.AuthorIDEQ(authorID),
		article.DeletedAtIsNil(),
	}
	if isPublished != nil {
		predicates = append(predicates, article.IsPublishedEQ(*isPublished))
	}

	total, err := r.DB(ctx).Article.Query().
		Where(predicates...).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return uint64(total), nil
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
func (r *ArticleRepo) ListByCategory(ctx context.Context, categoryID uint64, offset, limit int) ([]*aggregate.ArticleAggregate, error) {
	nodes, err := r.DB(ctx).Article.Query().
		Where(article.CategoryIDEQ(categoryID), article.DeletedAtIsNil(), article.IsPublishedEQ(true)).
		WithCategory().
		WithUser().
		Order(ent.Desc(article.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticleAggregates(nodes), nil
}

// Search 按标题、摘要、正文模糊搜索已发布文章，按 ID 倒序
func (r *ArticleRepo) Search(ctx context.Context, name string, offset, limit int) ([]*aggregate.ArticleAggregate, error) {
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
		WithCategory().
		WithUser().
		Order(ent.Desc(article.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityArticleAggregates(nodes), nil
}
