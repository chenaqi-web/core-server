package application

import (
	"context"
	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/infras/repo"
	"core-server/internal/model/dto"
	"core-server/internal/model/entity"
	"errors"
)

type CommentService struct {
	cfg      *config.Config
	log      *clog.Log
	repo     domain.CommentRepoDomain
	userRepo domain.UserRepoDomain
	artRepo  domain.ArticleRepoDomain
}

func NewCommentService(
	log *clog.Log,
	repo domain.CommentRepoDomain,
	userRepo domain.UserRepoDomain,
	artRepo domain.ArticleRepoDomain,
	cfg *config.Config,
) (*CommentService, error) {
	return &CommentService{
		cfg:      cfg,
		log:      log,
		repo:     repo,
		userRepo: userRepo,
		artRepo:  artRepo,
	}, nil
}

func (s *CommentService) CreateComment(ctx context.Context, req *dto.CreateCommentRequest) (bool, error) {
	err := s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		_, err := s.repo.CreateComment(ctx, &entity.Comment{
			ArticleID: req.ArticleID,
			UserID:    req.UserID,
			Content:   req.Content,
			// 一级评论，rootID是0
			RootID: 0,
		})
		if err != nil {
			return err
		}
		if err = s.artRepo.UpdateCommentCount(ctx, req.ArticleID, 1); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error(err.Error())
		return false, err
	}
	return true, nil
}

func (s *CommentService) CreateReply(ctx context.Context, req *dto.CreateReplyRequest) (bool, error) {
	// 开启事务进行修改
	err := s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		_, err := s.repo.CreateReply(ctx, &entity.Comment{
			ArticleID: req.ArticleID,
			UserID:    req.UserID,
			ParentID:  req.ParentID,
			RootID:    1, // 回复通常都不是根评论，用1来划分
			ReplyToID: req.ReplyToID,
			Content:   req.Content,
		})
		if err != nil {
			return err
		}
		// 增加根评论的下面的回复评论数
		if err = s.repo.IncrementChildCount(ctx, req.ParentID); err != nil {
			return err
		}
		// 修改主表的评论数
		if err = s.artRepo.UpdateCommentCount(ctx, req.ArticleID, 1); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error(err.Error())
		return false, err
	}
	return true, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, req *dto.DeleteCommentRequest) error {
	comment, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		s.log.Error(err.Error())
		return err
	}

	err = s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.SoftDelete(ctx, req.ID, req.UserID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return ErrCommentNotFound
			}
			return err
		}

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

		return s.artRepo.UpdateCommentCount(ctx, comment.ArticleID, -deletedCount)
	})
	if err != nil {
		s.log.Error(err.Error())
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

	items := make([]*dto.CommentInfoDTO, 0, len(comments))
	for _, c := range comments {
		items = append(items, dto.CommentInfoFromEntity(c, authorMap[c.UserID]))
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

	items := make([]*dto.CommentInfoDTO, 0, len(replies))
	for _, reply := range replies {
		items = append(items, dto.CommentInfoFromEntity(reply, authorMap[reply.UserID]))
	}

	return &dto.GetCommentRepliesResponse{
		Replies: items,
		Page:    int32(page),
		Size:    int32(size),
	}, nil
}
