package domain

import (
	"context"
	"core-server/internal/model/entity"
)

type ArticleRepo interface {
	Create(ctx context.Context, article *entity.Article) error
	DeleteByID(ctx context.Context, id, authorID uint64) error
	GetByID(ctx context.Context, id uint64) (*entity.Article, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*entity.Article, error)

	List(ctx context.Context, offset, limit int) ([]*entity.Article, error)
	ListByAuthor(ctx context.Context, authorID uint64, offset, limit int) ([]*entity.Article, error)
	ListByCategory(ctx context.Context, categoryID uint64, offset, limit int) ([]*entity.Article, error)

	Search(ctx context.Context, name string, offset, limit int) ([]*entity.Article, error)
}

type ArticleRepoDomain interface {
	ITransaction
	ArticleRepo
}
