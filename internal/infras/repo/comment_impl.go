package repo

import (
	"context"
	"time"

	"core-server/internal/infras/repo/ent"
	"core-server/internal/infras/repo/ent/comment"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/entity"
)

type CommentRepo struct{ *EntClient }

func NewCommentRepo(client *EntClient) *CommentRepo {
	return &CommentRepo{EntClient: client}
}

func (r *CommentRepo) CreateComment(ctx context.Context, value *entity.Comment) (uint64, error) {
	node, err := r.DB(ctx).Comment.Create().
		SetArticleID(value.ArticleID).
		SetUserID(value.UserID).
		SetRootID(value.RootID).
		SetReplyToID(value.ReplyToID).
		SetContent(value.Content).
		SetLikeCount(value.LikeCount).
		SetChildCount(value.ChildCount).
		Save(ctx)
	if err != nil {
		return 0, err
	}
	return node.ID, nil
}

func (r *CommentRepo) CreateReply(ctx context.Context, value *entity.Comment) (uint64, error) {
	return r.CreateComment(ctx, value)
}

func (r *CommentRepo) GetByID(ctx context.Context, id uint64) (*aggregate.CommentAggregate, error) {
	node, err := r.DB(ctx).Comment.Query().
		Where(comment.IDEQ(id), comment.DeletedAtIsNil()).
		WithUser().
		WithArticle().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return aggregate.NewCommentAggregate(toEntityComment(node), toEntityUser(node.Edges.User)), nil
}

func (r *CommentRepo) ListByIDs(ctx context.Context, ids []uint64) ([]*aggregate.CommentAggregate, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	nodes, err := r.DB(ctx).Comment.Query().
		Where(comment.IDIn(ids...), comment.DeletedAtIsNil()).
		WithUser().
		WithArticle().
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityCommentAggregates(nodes), nil
}

func (r *CommentRepo) SoftDelete(ctx context.Context, id, userID uint64) error {
	affected, err := r.DB(ctx).Comment.Update().
		Where(comment.IDEQ(id), comment.UserIDEQ(userID), comment.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CommentRepo) SoftDeleteRepliesByRoot(ctx context.Context, rootID uint64) (int64, error) {
	affected, err := r.DB(ctx).Comment.Update().
		Where(comment.RootIDEQ(rootID), comment.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	return int64(affected), err
}

func (r *CommentRepo) IncrementChildCount(ctx context.Context, rootID uint64) error {
	return r.DB(ctx).Comment.Update().
		Where(comment.IDEQ(rootID), comment.DeletedAtIsNil()).
		AddChildCount(1).
		Exec(ctx)
}

func (r *CommentRepo) DecrementChildCount(ctx context.Context, rootID uint64) error {
	return r.DB(ctx).Comment.Update().
		Where(comment.IDEQ(rootID), comment.DeletedAtIsNil(), comment.ChildCountGT(0)).
		AddChildCount(-1).
		Exec(ctx)
}

func (r *CommentRepo) ListTopByArticle(ctx context.Context, articleID uint64, offset, limit int) ([]*aggregate.CommentAggregate, error) {
	nodes, err := r.DB(ctx).Comment.Query().
		Where(comment.ArticleIDEQ(articleID), comment.RootIDEQ(0), comment.DeletedAtIsNil()).
		WithUser().
		WithArticle().
		Order(ent.Desc(comment.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityCommentAggregates(nodes), nil
}

func (r *CommentRepo) ListRepliesByRoot(ctx context.Context, rootID uint64, offset, limit int) ([]*aggregate.CommentAggregate, error) {
	nodes, err := r.DB(ctx).Comment.Query().
		Where(comment.RootIDEQ(rootID), comment.DeletedAtIsNil()).
		WithUser().
		WithArticle().
		Order(ent.Asc(comment.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toEntityCommentAggregates(nodes), nil
}
