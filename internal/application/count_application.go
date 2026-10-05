package application

import (
	"context"
	"core-server/internal/infras/clog"
	"core-server/internal/model/entity"

	"core-server/internal/domain"
	"core-server/internal/model/enum"

	"go.uber.org/zap"
)

// 这里作为一个计数中心，是可以单独拆分出去的

type CountService struct {
	log      *clog.Log
	repo     domain.CountRepoDomain
	userRepo domain.UserRepoDomain
}

func NewCountService(
	log *clog.Log,
	repo domain.CountRepoDomain,
	userRepo domain.UserRepoDomain,
) *CountService {
	return &CountService{
		repo:     repo,
		userRepo: userRepo,
		log:      log,
	}
}

// 获取计数

func (s *CountService) GetArticleInteractionCount(ctx context.Context, articleID uint64) (*entity.InteractionStats, error) {
	storedCounts, err := s.repo.GetByObject(ctx, enum.ObjectTypeArticle, articleID)
	if err != nil {
		s.log.Error("GetArticleInteractionCount error:", zap.Error(err))
		return nil, err
	}
	counts := &entity.InteractionStats{}
	for _, count := range storedCounts {
		switch count.InteractionType {
		case enum.InteractionTypeLike:
			counts.LikeCount = uint64(count.Count)
		case enum.InteractionTypeComment:
			counts.CommentCount = uint64(count.Count)
		case enum.InteractionTypeView:
			counts.ViewCount = uint64(count.Count)
		case enum.InteractionTypeFavor:
			counts.FavorCount = uint64(count.Count)
		}
	}
	return counts, nil
}

func (s *CountService) BatchGetArticleInteractionCounts(ctx context.Context, articleIDs []uint64) (map[uint64]*entity.InteractionStats, error) {
	statsByArticleID := make(map[uint64]*entity.InteractionStats, len(articleIDs))
	for _, articleID := range articleIDs {
		statsByArticleID[articleID] = &entity.InteractionStats{}
	}

	// 批量获取
	storedCounts, err := s.repo.GetByObjects(ctx, enum.ObjectTypeArticle, articleIDs)
	if err != nil {
		return nil, err
	}

	for _, count := range storedCounts {
		stats, ok := statsByArticleID[count.ObjectID]
		if !ok {
			continue
		}
		switch count.InteractionType {
		case enum.InteractionTypeLike:
			stats.LikeCount = uint64(count.Count)
		case enum.InteractionTypeComment:
			stats.CommentCount = uint64(count.Count)
		case enum.InteractionTypeView:
			stats.ViewCount = uint64(count.Count)
		case enum.InteractionTypeFavor:
			stats.FavorCount = uint64(count.Count)
		}
	}
	return statsByArticleID, nil
}

// =====================================================================================================================
// 计数修改

func (s *CountService) UpdateObjectTypeWithInteractionType(ctx context.Context, object, interaction string, objectID uint64, delta int64) error {
	if err := s.repo.Upsert(ctx, &entity.InteractionCount{
		ObjectType:      enum.ParseObjectType(object),
		ObjectID:        objectID,
		InteractionType: enum.ParseInteractionType(interaction),
	}, delta); err != nil {
		s.log.Error("UpdateObjectTypeWithInteractionType error:", zap.Error(err))
		return err
	}
	return nil
}

func (s *CountService) UpdateLikeCount(ctx context.Context, userID uint64, delta int64) error {
	if err := s.userRepo.UpdateLikeCount(ctx, userID, delta); err != nil {
		s.log.Error("UpdateLikeCount error:", zap.Error(err))
		return err
	}
	return nil
}

func (s *CountService) UpdateReceiveLikeCount(ctx context.Context, userID uint64, delta int64) error {
	if err := s.userRepo.UpdateReceiveLikeCount(ctx, userID, delta); err != nil {
		s.log.Error("UpdateReceiveLikeCount error:", zap.Error(err))
		return err
	}
	return nil
}

func (s *CountService) UpdateArticleCount(ctx context.Context, userID uint64, delta int64) error {
	if err := s.userRepo.UpdateArticleCount(ctx, userID, delta); err != nil {
		s.log.Error("UpdateArticleCount error:", zap.Error(err))
		return err
	}
	return nil
}

func (s *CountService) UpdateViewCount(ctx context.Context, userID uint64, delta int64) error {
	if err := s.userRepo.UpdateViewCount(ctx, userID, delta); err != nil {
		s.log.Error("UpdateViewCount error:", zap.Error(err))
		return err
	}
	return nil
}

func (s *CountService) UpdateCommentCount(ctx context.Context, userID uint64, delta int64) error {
	if err := s.userRepo.UpdateCommentCount(ctx, userID, delta); err != nil {
		s.log.Error("UpdateCommentCount error:", zap.Error(err))
		return err
	}
	return nil
}
