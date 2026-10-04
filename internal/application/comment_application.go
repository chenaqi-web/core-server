package application

import (
	"context"
	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/infras/repo"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/dto"
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"
	"errors"

	"go.uber.org/zap"
)

type CommentService struct {
	cfg          *config.Config
	log          *clog.Log
	repo         domain.CommentRepoDomain
	articleRepo  domain.ArticleRepoDomain
	userRepo     domain.UserRepoDomain
	countService *CountService
}

func NewCommentService(
	log *clog.Log,
	repo domain.CommentRepoDomain,
	articleRepo domain.ArticleRepoDomain,
	userRepo domain.UserRepoDomain,
	countService *CountService,
	cfg *config.Config,
) (*CommentService, error) {
	return &CommentService{
		cfg:          cfg,
		log:          log,
		repo:         repo,
		articleRepo:  articleRepo,
		userRepo:     userRepo,
		countService: countService,
	}, nil
}

func (s *CommentService) CreateComment(ctx context.Context, req *dto.CreateCommentRequest) (bool, error) {
	err := s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		article, err := s.articleRepo.GetByID(ctx, req.ArticleID)
		if err != nil {
			return err
		}

		_, err = s.repo.CreateComment(ctx, &entity.Comment{
			ArticleID: req.ArticleID,
			UserID:    req.UserID,
			Content:   req.Content,
			RootID:    0,
		})
		if err != nil {
			return err
		}

		if err := s.countService.UpdateObjectTypeWithInteractionType(ctx, enum.ObjectTypeArticle.String(), enum.InteractionTypeComment.String(), req.ArticleID, 1); err != nil {
			return err
		}
		return s.userRepo.UpdateCommentCount(ctx, article.Article.AuthorID, 1)
	})
	if err != nil {
		s.log.Error("CreateComment", zap.Error(err))
		return false, err
	}
	return true, nil
}

func (s *CommentService) CreateReply(ctx context.Context, req *dto.CreateReplyRequest) (bool, error) {
	// 1.首先校验父评论是否存在
	rootComment, err := s.repo.GetByID(ctx, req.RootID)
	if err != nil || rootComment == nil || rootComment.Comment == nil || !rootComment.Comment.IsTopLevel() {
		return false, ErrCommentNotFound
	}

	replyToID := req.ReplyToID
	if replyToID == 0 {
		replyToID = rootComment.Comment.ID
	}

	// 开启事务进行修改
	err = s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		article, err := s.articleRepo.GetByID(ctx, rootComment.Comment.ArticleID)
		if err != nil {
			return err
		}

		_, err = s.repo.CreateReply(ctx, &entity.Comment{
			ArticleID: rootComment.Comment.ArticleID,
			UserID:    req.UserID,
			RootID:    rootComment.Comment.ID,
			ReplyToID: replyToID,
			Content:   req.Content,
		})
		if err != nil {
			return err
		}
		if err = s.repo.IncrementChildCount(ctx, rootComment.Comment.ID); err != nil {
			return err
		}

		if err := s.countService.UpdateObjectTypeWithInteractionType(ctx, enum.ObjectTypeArticle.String(), enum.InteractionTypeComment.String(), rootComment.Comment.ArticleID, 1); err != nil {
			return err
		}
		return s.userRepo.UpdateCommentCount(ctx, article.Article.AuthorID, 1)
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
		commentAggregate, err := s.repo.GetByID(ctx, req.ID)
		if err != nil {
			return err
		}
		if commentAggregate == nil || commentAggregate.Comment == nil {
			return ErrCommentNotFound
		}
		comment := commentAggregate.Comment
		articleID := comment.ArticleID

		article, err := s.articleRepo.GetByID(ctx, articleID)
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
			replyCount, err := s.repo.SoftDeleteRepliesByRoot(ctx, comment.ID)
			if err != nil {
				return err
			}
			deletedCount += replyCount
		} else if err := s.repo.DecrementChildCount(ctx, comment.RootID); err != nil {
			return err
		}

		if err := s.countService.UpdateObjectTypeWithInteractionType(ctx, enum.ObjectTypeArticle.String(), enum.InteractionTypeComment.String(), articleID, -deletedCount); err != nil {
			return err
		}
		return s.userRepo.UpdateCommentCount(ctx, article.Article.AuthorID, -deletedCount)
	})
	if err != nil {
		s.log.Error("DeleteComment", zap.Error(err))
		return err
	}
	return nil
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

	likeCounts, err := s.loadCommentLikeCounts(ctx, comments)
	if err != nil {
		s.log.Error("load comment like counts error", zap.Error(err))
		return nil, err
	}

	items := make([]*dto.CommentInfoDTO, 0, len(comments))
	for _, itemAggregate := range comments {
		if itemAggregate == nil || itemAggregate.Comment == nil {
			continue
		}
		item := dto.CommentInfoFromEntity(itemAggregate.Comment, itemAggregate.Author)
		item.LikeCount = likeCounts[itemAggregate.Comment.ID]
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

	replies, err := s.repo.ListRepliesByRoot(ctx, req.RootID, offset, size)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	replyToIDs := make([]uint64, 0, len(replies))
	for _, reply := range replies {
		if reply != nil && reply.Comment != nil && reply.Comment.ReplyToID != 0 {
			replyToIDs = append(replyToIDs, reply.Comment.ReplyToID)
		}
	}
	replyToComments, err := s.repo.ListByIDs(ctx, replyToIDs)
	if err != nil {
		s.log.Error("load reply targets error", zap.Error(err))
		return nil, err
	}
	replyToCommentMap := make(map[uint64]*aggregate.CommentAggregate, len(replyToComments))
	for _, replyToComment := range replyToComments {
		if replyToComment != nil && replyToComment.Comment != nil {
			replyToCommentMap[replyToComment.Comment.ID] = replyToComment
		}
	}
	likeCounts, err := s.loadCommentLikeCounts(ctx, replies)
	if err != nil {
		s.log.Error("load reply like counts error", zap.Error(err))
		return nil, err
	}

	items := make([]*dto.CommentInfoDTO, 0, len(replies))
	for _, reply := range replies {
		if reply == nil || reply.Comment == nil {
			continue
		}
		item := dto.CommentInfoFromEntity(reply.Comment, reply.Author)
		item.LikeCount = likeCounts[reply.Comment.ID]
		if replyToComment := replyToCommentMap[reply.Comment.ReplyToID]; replyToComment != nil && replyToComment.Author != nil {
			item.ReplyToUserName = replyToComment.Author.Name
		}
		items = append(items, item)
	}

	return &dto.GetCommentRepliesResponse{
		Replies: items,
		Page:    int32(page),
		Size:    int32(size),
	}, nil
}

// =====================================================================================================================

func (s *CommentService) loadCommentLikeCounts(ctx context.Context, comments []*aggregate.CommentAggregate) (map[uint64]uint32, error) {
	commentIDs := make([]uint64, 0, len(comments))
	for _, item := range comments {
		if item != nil && item.Comment != nil {
			commentIDs = append(commentIDs, item.Comment.ID)
		}
	}
	counts := make(map[uint64]uint32, len(commentIDs))
	if len(commentIDs) == 0 {
		return counts, nil
	}

	storedCounts, err := s.countService.repo.GetByObjects(ctx, enum.ObjectTypeComment, commentIDs)
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
