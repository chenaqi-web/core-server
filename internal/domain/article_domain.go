package domain

import (
	"backend/core-server/internal/model/entity"
	"context"
)

type ArticleRepo interface {
	Create(ctx context.Context, article *entity.Article) error
	GetByID(ctx context.Context, id uint64) (*entity.Article, error)

	List(ctx context.Context, offset, limit int) ([]*entity.Article, uint64, error)
}

type ArticleRepoDomain interface {
	ITransaction
	ArticleRepo
}
