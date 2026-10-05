package domain

import (
	"context"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/entity"
)

type ArticleRepo interface {
	Create(ctx context.Context, article *entity.Article) error
	Edit(ctx context.Context, article *entity.Article) error
	PublishDraft(ctx context.Context, id, authorID uint64) error
	DeleteByID(ctx context.Context, id, authorID uint64, role string) error
	DeleteDraftByID(ctx context.Context, id, authorID uint64) error

	GetByID(ctx context.Context, id uint64) (*aggregate.ArticleAggregate, error)
	GetByIDForManage(ctx context.Context, id uint64) (*aggregate.ArticleAggregate, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*aggregate.ArticleAggregate, error)
	List(ctx context.Context, offset, limit int) ([]*aggregate.ArticleAggregate, error)
	ListByAuthor(ctx context.Context, authorID uint64, offset, limit int, isPublished *bool) ([]*aggregate.ArticleAggregate, uint64, error)
	ListByCategory(ctx context.Context, categoryID uint64, offset, limit int) ([]*aggregate.ArticleAggregate, error)

	Search(ctx context.Context, name string, offset, limit int) ([]*aggregate.ArticleAggregate, error)
}

type ArticleRepoDomain interface {
	ITransaction
	ArticleRepo
}
