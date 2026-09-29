package domain

import (
	"context"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/entity"
)

type CommentRepo interface {
	CreateComment(ctx context.Context, comment *entity.Comment) (uint64, error)
	CreateReply(ctx context.Context, comment *entity.Comment) (uint64, error)
	SoftDelete(ctx context.Context, id, userID uint64) error
	SoftDeleteRepliesByRoot(ctx context.Context, rootID uint64) (int64, error)

	GetByID(ctx context.Context, id uint64) (*aggregate.CommentAggregate, error)
	ListByIDs(ctx context.Context, ids []uint64) ([]*aggregate.CommentAggregate, error)
	IncrementChildCount(ctx context.Context, rootID uint64) error
	DecrementChildCount(ctx context.Context, rootID uint64) error

	ListTopByArticle(ctx context.Context, articleID uint64, offset, limit int) ([]*aggregate.CommentAggregate, error)
	ListRepliesByRoot(ctx context.Context, rootID uint64, offset, limit int) ([]*aggregate.CommentAggregate, error)
}

type CommentRepoDomain interface {
	ITransaction
	CommentRepo
}
