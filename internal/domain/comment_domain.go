package domain

import (
	"backend/core-server/internal/model/entity"
	"context"
)

type CommentRepo interface {
	CreateComment(ctx context.Context, comment *entity.Comment) (uint64, error)
	CreateReply(ctx context.Context, comment *entity.Comment) (uint64, error)
	SoftDelete(ctx context.Context, id, userID uint64) error

	GetByID(ctx context.Context, id uint64) (*entity.Comment, error)
	IncrementChildCount(ctx context.Context, rootID uint64) error
	DecrementChildCount(ctx context.Context, rootID uint64) error

	ListTopByArticle(ctx context.Context, articleID uint64, offset, limit int) ([]*entity.Comment, error)
	ListRepliesByParent(ctx context.Context, parentID uint64, offset, limit int) ([]*entity.Comment, error)
}

type CommentRepoDomain interface {
	ITransaction
	CommentRepo
}
