package application

import (
	"context"
	"core-server/internal/infras/clog"
	"core-server/internal/model/entity"

	"core-server/internal/domain"
	"core-server/internal/model/enum"

	"go.uber.org/zap"
)

type CountService struct {
	log       *clog.Log
	repo      domain.CountRepoDomain
	likeCache domain.LikeCacheDomain
}

func NewCountService(
	log *clog.Log,
	repo domain.CountRepoDomain,
) *CountService {
	return &CountService{
		repo: repo,
		log:  log,
	}
}

// IncrementArticleView 增加博客的浏览量
func (s *CountService) IncrementArticleView(ctx context.Context, articleID uint64) error {
	if err := s.repo.Upsert(ctx, &entity.InteractionCount{
		ObjectType:      enum.ObjectTypeArticle,
		ObjectID:        articleID,
		InteractionType: enum.InteractionTypeView,
	}, 1); err != nil {
		s.log.Error("IncrementArticleView error:", zap.Error(err))
		return err
	}
	return nil
}

// GetArticleInteractionCount 拿到博客的3维(浏览，点赞，评论数量)
// todo 后续点赞应该优先从缓存中提取,大量点赞的情况下发送到kafka异步消费，写入数据库比较慢
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
		}
	}
	return counts, nil
}

// BatchGetArticleInteractionCounts 批量获取文章的三维(主要是服务列表操作的)
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
		}
	}
	return statsByArticleID, nil
}
