package application

import (
	"context"
	"core-server/internal/infras/clog"
	"time"

	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/entity"
	"core-server/internal/model/enum"
)

type LikeService struct {
	cfg          *config.Config
	log          *clog.Log
	repo         domain.LikeRepoDomain
	articleRepo  domain.ArticleRepoDomain
	userRepo     domain.UserRepoDomain
	countService *CountService
}

func NewLikeService(
	log *clog.Log,
	repo domain.LikeRepoDomain,
	articleRepo domain.ArticleRepoDomain,
	userRepo domain.UserRepoDomain,
	countService *CountService,
	cfg *config.Config,
) (*LikeService, error) {
	return &LikeService{
		cfg:          cfg,
		repo:         repo,
		articleRepo:  articleRepo,
		userRepo:     userRepo,
		countService: countService,
		log:          log,
	}, nil
}

// =====================================================================================================================
// 点赞状态

func (s *LikeService) HasThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) (bool, error) {
	interaction, err := s.repo.QueryWithCondition(
		ctx,
		userID,
		objectType,
		objectID,
		entity.LikeStatusTypeThumbUp.String(),
	)
	if err != nil {
		return false, err
	}
	return interaction != nil, nil
}

// =====================================================================================================================
// 点赞操作

func (s *LikeService) ThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) error {
	exists, err := s.HasThumbUp(ctx, userID, objectType, objectID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyLiked
	}

	like := &entity.InteractionLike{
		UserID:     userID,
		ObjectType: enum.ParseObjectType(objectType),
		ObjectID:   objectID,
		Status:     entity.LikeStatusTypeThumbUp,
		Version:    time.Now().UnixMicro(),
	}

	err = s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		affected, err := s.repo.Upsert(ctx, like)
		if err != nil {
			return err
		}
		if affected == 0 {
			return nil
		}
		if err := s.countService.repo.Upsert(ctx, &entity.InteractionCount{
			ObjectType:      enum.ParseObjectType(objectType),
			ObjectID:        objectID,
			InteractionType: enum.InteractionTypeLike,
		}, 1); err != nil {
			return err
		}
		if err := s.userRepo.IncrementLikeCount(ctx, userID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("Error in ThumbUp transaction")
		return err
	}

	return nil
}

func (s *LikeService) CancelThumbUp(ctx context.Context, userID uint64, objectType string, objectID uint64) error {
	exists, err := s.HasThumbUp(ctx, userID, objectType, objectID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	like := &entity.InteractionLike{
		UserID:     userID,
		ObjectType: enum.ParseObjectType(objectType),
		ObjectID:   objectID,
		Status:     entity.LikeStatusTypeNothing,
		Version:    time.Now().UnixMicro(),
	}

	return s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		affected, err := s.repo.UpdateWithCondition(ctx, entity.LikeStatusTypeThumbUp.String(), like)
		if err != nil {
			return err
		}
		if affected == 0 {
			return nil
		}
		if err := s.countService.repo.Upsert(ctx, &entity.InteractionCount{
			ObjectType:      enum.ParseObjectType(objectType),
			ObjectID:        objectID,
			InteractionType: enum.InteractionTypeLike,
		}, -1); err != nil {
			return err
		}
		if err := s.userRepo.DecrementLikeCount(ctx, userID); err != nil {
			return err
		}
		return nil
	})
}

// =====================================================================================================================
// 点赞列表方面

func (s *LikeService) UserLikeList(ctx context.Context, userID uint64, objectType string, page, pageSize int) ([]*aggregate.ArticleAggregate, error) {
	offset := (page - 1) * pageSize

	likes, err := s.repo.PageQueryLikeObjects(ctx, userID, objectType, offset, pageSize)
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(likes))
	for _, like := range likes {
		ids = append(ids, like.ObjectID)
	}

	articles, err := s.articleRepo.ListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return articles, nil
}
