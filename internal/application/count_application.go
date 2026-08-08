package application

import (
	"context"

	"core-server/internal/domain"
	"core-server/internal/model/dto"
	"core-server/internal/model/enum"
)

type CountService struct {
	repo domain.CountRepoDomain
}

func NewCountService(repo domain.CountRepoDomain) *CountService {
	return &CountService{repo: repo}
}

func (s *CountService) GetArticleInteractionCount(ctx context.Context, articleID uint64) (*dto.ArticleInteractionCount, error) {
	storedCounts, err := s.repo.GetByObject(ctx, enum.ObjectTypeArticle, articleID)
	if err != nil {
		return nil, err
	}

	counts := &dto.ArticleInteractionCount{}
	for _, count := range storedCounts {
		if count.Count < 0 {
			continue
		}

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
