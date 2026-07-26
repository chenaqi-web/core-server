package domain

import (
	"backend/core-server/internal/model/entity"
	"context"
)

type CommentRepo interface {
	CreateComment(ctx context.Context, comment *entity.Comment) (uint64, error)
	CreateReply(ctx context.Context, comment *entity.Comment) (uint64, error)
	GetByID(ctx context.Context, id uint64) (*entity.Comment, error)
	SoftDelete(ctx context.Context, id, userID uint64) error
	IncrementChildCount(ctx context.Context, rootID uint64) error
	DecrementChildCount(ctx context.Context, rootID uint64) error

	ListTopByArticle(ctx context.Context, articleID uint64, offset, limit int, orderBy string) ([]*entity.Comment, error)
	CountTopByArticle(ctx context.Context, articleID uint64) (int, error)
	ListRepliesByRoot(ctx context.Context, rootID uint64, offset, limit int) ([]*entity.Comment, error)
	CountRepliesByRoot(ctx context.Context, rootID uint64) (int, error)
	ListRepliesByRoots(ctx context.Context, rootIDs []uint64) ([]*entity.Comment, error)
	ListByUser(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Comment, error)
	CountByUser(ctx context.Context, userID uint64) (int, error)
}

type CommentRepoDomain interface {
	ITransaction
	CommentRepo
}
