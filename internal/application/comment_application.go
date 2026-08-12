package application

import (
	"context"
	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/infras/repo"
	"core-server/internal/model/dto"
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"
	"errors"

	"go.uber.org/zap"
)

type CommentService struct {
	cfg       *config.Config
	log       *clog.Log
	repo      domain.CommentRepoDomain
	userRepo  domain.UserRepoDomain
	countRepo domain.CountRepoDomain
}

func NewCommentService(
	log *clog.Log,
	repo domain.CommentRepoDomain,
	userRepo domain.UserRepoDomain,
	countRepo domain.CountRepoDomain,
	cfg *config.Config,
) (*CommentService, error) {
	return &CommentService{
		cfg:       cfg,
		log:       log,
		repo:      repo,
		userRepo:  userRepo,
		countRepo: countRepo,
	}, nil
}

func (s *CommentService) CreateComment(ctx context.Context, req *dto.CreateCommentRequest) (bool, error) {
	err := s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		_, err := s.repo.CreateComment(ctx, &entity.Comment{
			ArticleID: req.ArticleID,
			UserID:    req.UserID,
			Content:   req.Content,
			RootID:    0,
		})
		if err != nil {
			return err
		}

		count := &entity.InteractionCount{
			ObjectType:      enum.ObjectTypeArticle,
			ObjectID:        req.ArticleID,
			InteractionType: enum.InteractionTypeComment,
		}

		// 先创建评论关系，如何修改计数表
		return s.countRepo.Upsert(ctx, count, 1)
	})
	if err != nil {
		s.log.Error("CreateComment", zap.Error(err))
		return false, err
	}
	return true, nil
}

func (s *CommentService) CreateReply(ctx context.Context, req *dto.CreateReplyRequest) (bool, error) {
	// 开启事务进行修改
	err := s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		rootComment, err := s.repo.GetByID(ctx, req.ParentID)
		if err != nil {
			return err
		}
		if rootComment == nil || !rootComment.IsTopLevel() || rootComment.ArticleID != req.ArticleID {
			return ErrCommentNotFound
		}

		if req.ReplyToID != 0 && req.ReplyToID != rootComment.ID {
			replyToComment, err := s.repo.GetByID(ctx, req.ReplyToID)
			if err != nil {
				return err
			}
			if replyToComment == nil || replyToComment.ArticleID != req.ArticleID || replyToComment.ParentID != rootComment.ID {
				return ErrCommentNotFound
			}
		}

		_, err = s.repo.CreateReply(ctx, &entity.Comment{
			ArticleID: req.ArticleID,
			UserID:    req.UserID,
			// All replies are second-level comments under the top-level comment.
			ParentID:  rootComment.ID,
			RootID:    rootComment.ID,
			ReplyToID: req.ReplyToID,
			Content:   req.Content,
		})
		if err != nil {
			return err
		}
		if err = s.repo.IncrementChildCount(ctx, rootComment.ID); err != nil {
			return err
		}

		count := &entity.InteractionCount{
			ObjectType:      enum.ObjectTypeArticle,
			ObjectID:        req.ArticleID,
			InteractionType: enum.InteractionTypeComment,
		}
		return s.countRepo.Upsert(ctx, count, 1)
	})
	if err != nil {
		s.log.Error("CreateReply", zap.Error(err))
		return false, err
	}
	return true, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, req *dto.DeleteCommentRequest) error {
	err := s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		// 1.首先获取该评论的内容
		comment, err := s.repo.GetByID(ctx, req.ID)
		if err != nil {
			return err
		}

		// 2.首先删除这个评论
		if err := s.repo.SoftDelete(ctx, req.ID, req.UserID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return ErrCommentNotFound
			}
			return err
		}

		// 3.判断是不是根评论(其实也可以异步删除的)
		deletedCount := int64(1)
		if comment.IsTopLevel() {
			replyCount, err := s.repo.SoftDeleteRepliesByParent(ctx, comment.ID)
			if err != nil {
				return err
			}
			deletedCount += replyCount
		} else if err := s.repo.DecrementChildCount(ctx, comment.ParentID); err != nil {
			return err
		}

		count := &entity.InteractionCount{
			ObjectType:      enum.ObjectTypeArticle,
			ObjectID:        comment.ArticleID,
			InteractionType: enum.InteractionTypeComment,
		}
		return s.countRepo.Upsert(ctx, count, -deletedCount)
	})
	if err != nil {
		s.log.Error("DeleteComment", zap.Error(err))
	}
	return err
}

func (s *CommentService) GetArticleComments(ctx context.Context, req *dto.GetArticleCommentsRequest) (*dto.GetArticleCommentsResponse, error) {
	page := Page(int(req.Page))
	size := Size(int(req.Size))
	offset := (page - 1) * size

	comments, err := s.repo.ListTopByArticle(ctx, req.ArticleID, offset, size)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	authorMap, err := LoadUserMap(ctx, s.userRepo, CollectCommentUserIDs(comments))
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	likeCounts, err := s.loadCommentLikeCounts(ctx, comments)
	if err != nil {
		s.log.Error("load comment like counts error", zap.Error(err))
		return nil, err
	}

	items := make([]*dto.CommentInfoDTO, 0, len(comments))
	for _, c := range comments {
		item := dto.CommentInfoFromEntity(c, authorMap[c.UserID])
		item.LikeCount = likeCounts[c.ID]
		items = append(items, item)
	}

	return &dto.GetArticleCommentsResponse{
		Comments: items,
		Page:     int32(page),
		Size:     int32(size),
	}, nil
}

func (s *CommentService) GetCommentReplies(ctx context.Context, req *dto.GetCommentRepliesRequest) (*dto.GetCommentRepliesResponse, error) {
	page := Page(int(req.Page))
	size := Size(int(req.Size))
	offset := (page - 1) * size

	replies, err := s.repo.ListRepliesByParent(ctx, req.ParentID, offset, size)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	authorMap, err := LoadUserMap(ctx, s.userRepo, CollectCommentUserIDs(replies))
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	likeCounts, err := s.loadCommentLikeCounts(ctx, replies)
	if err != nil {
		s.log.Error("load reply like counts error", zap.Error(err))
		return nil, err
	}

	items := make([]*dto.CommentInfoDTO, 0, len(replies))
	for _, reply := range replies {
		item := dto.CommentInfoFromEntity(reply, authorMap[reply.UserID])
		item.LikeCount = likeCounts[reply.ID]
		items = append(items, item)
	}

	return &dto.GetCommentRepliesResponse{
		Replies: items,
		Page:    int32(page),
		Size:    int32(size),
	}, nil
}

func (s *CommentService) loadCommentLikeCounts(ctx context.Context, comments []*entity.Comment) (map[uint64]uint32, error) {
	commentIDs := make([]uint64, 0, len(comments))
	for _, comment := range comments {
		if comment != nil {
			commentIDs = append(commentIDs, comment.ID)
		}
	}
	counts := make(map[uint64]uint32, len(commentIDs))
	if len(commentIDs) == 0 {
		return counts, nil
	}

	storedCounts, err := s.countRepo.GetByObjects(ctx, enum.ObjectTypeComment, commentIDs)
	if err != nil {
		return nil, err
	}
	for _, count := range storedCounts {
		if count != nil && count.InteractionType == enum.InteractionTypeLike && count.Count > 0 {
			counts[count.ObjectID] = uint32(count.Count)
		}
	}
	return counts, nil
}
