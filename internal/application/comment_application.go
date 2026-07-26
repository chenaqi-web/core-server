package application

import (
	"context"
	"errors"
	"strings"

	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/clog"
	"backend/core-server/internal/infras/repo"
	"backend/core-server/internal/model/dto"
	"backend/core-server/internal/model/entity"
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
		_, err := s.repo.Create(ctx, &entity.Comment{
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

func (s *CommentService) CreateReply(ctx context.Context, req *dto.CreateReplyRequest) (*dto.CreateReplyResponse, error) {
	if req == nil || strings.TrimSpace(req.Content) == "" {
		return nil, ErrCommentInvalid
	}

	content := strings.TrimSpace(req.Content)
	var replyID uint64
	err := s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		id, err := s.repo.Create(ctx, &entity.Comment{
			ArticleID:   root.ArticleID,
			UserID:      req.UserID,
			ParentID:    root.ID,
			RootID:      root.ID,
			ReplyToID:   req.ReplyToID,
			ReplyToName: req.ReplyToName,
			Content:     content,
		})
		if err != nil {
			return err
		}
		if err = s.repo.IncrementChildCount(ctx, root.ID); err != nil {
			return err
		}
		if err = s.artRepo.UpdateCommentCount(ctx, root.ArticleID, 1); err != nil {
			return err
		}
		replyID = id
		return nil
	})
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	return &dto.CreateReplyResponse{ReplyID: replyID}, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, req *dto.DeleteCommentRequest) error {
	if req == nil {
		return ErrCommentInvalid
	}

	comment, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		s.log.Error(err.Error())
		return err
	}
	if comment == nil {
		return ErrCommentNotFound
	}
	if comment.UserID != req.UserID {
		return ErrCommentForbidden
	}

	return s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.SoftDelete(ctx, req.ID, req.UserID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return ErrCommentNotFound
			}
			return err
		}
		if comment.IsTopLevel() {
			return s.artRepo.UpdateCommentCount(ctx, comment.ArticleID, -1)
		}
		if err := s.repo.DecrementChildCount(ctx, comment.RootID); err != nil {
			return err
		}
		return s.artRepo.UpdateCommentCount(ctx, comment.ArticleID, -1)
	})
}

func (s *CommentService) GetArticleComments(ctx context.Context, req *dto.GetArticleCommentsRequest) (*dto.GetArticleCommentsResponse, error) {
	if req == nil {
		return nil, ErrCommentInvalid
	}

	page := Page(int(req.Page))
	size := Size(int(req.Size))
	offset := (page - 1) * size

	comments, err := s.repo.ListTopByArticle(ctx, req.ArticleID, offset, size, req.OrderBy)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	total, err := s.repo.CountTopByArticle(ctx, req.ArticleID)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	rootIDs := make([]uint64, 0, len(comments))
	for _, c := range comments {
		rootIDs = append(rootIDs, c.ID)
	}

	allReplies, err := s.repo.ListRepliesByRoots(ctx, rootIDs)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	previewMap := GroupPreviewReplies(allReplies, previewReplyLimit)
	allComments := append([]*entity.Comment{}, comments...)
	for _, replies := range previewMap {
		allComments = append(allComments, replies...)
	}

	authorMap, err := LoadUserMap(ctx, s.userRepo, CollectCommentUserIDs(allComments))
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	items := make([]*dto.CommentInfoDTO, 0, len(comments))
	for _, c := range comments {
		info := dto.CommentInfoFromEntity(c, authorMap[c.UserID])
		for _, reply := range previewMap[c.ID] {
			info.Replies = append(info.Replies, dto.CommentInfoFromEntity(reply, authorMap[reply.UserID]))
		}
		items = append(items, info)
	}

	return &dto.GetArticleCommentsResponse{
		Comments: items,
		Page:     int32(page),
		Size:     int32(size),
		Total:    int32(total),
	}, nil
}

func (s *CommentService) GetCommentReplies(ctx context.Context, req *dto.GetCommentRepliesRequest) (*dto.GetCommentRepliesResponse, error) {
	if req == nil {
		return nil, ErrCommentInvalid
	}

	page := Page(int(req.Page))
	size := Size(int(req.Size))
	offset := (page - 1) * size

	replies, err := s.repo.ListRepliesByRoot(ctx, req.RootID, offset, size)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	total, err := s.repo.CountRepliesByRoot(ctx, req.RootID)
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
		Total:   int32(total),
	}, nil
}

func (s *CommentService) GetUserComments(ctx context.Context, req *dto.GetUserCommentsRequest) (*dto.GetUserCommentsResponse, error) {
	if req == nil {
		return nil, ErrCommentInvalid
	}

	page := Page(int(req.Page))
	size := Size(int(req.Size))
	offset := (page - 1) * size

	comments, err := s.repo.ListByUser(ctx, req.UserID, offset, size)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	total, err := s.repo.CountByUser(ctx, req.UserID)
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

	return &dto.GetUserCommentsResponse{
		Comments: items,
		Page:     int32(page),
		Size:     int32(size),
		Total:    int32(total),
	}, nil
}
