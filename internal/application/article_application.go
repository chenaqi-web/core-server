package application

import (
	"backend/core-server/internal/rpc/articlepb"
	"context"

	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/clog"
	"backend/core-server/internal/model/entity"
)

type ArticleService struct {
	cfg  *config.Config
	log  *clog.Log
	repo domain.ArticleRepoDomain
}

func NewArticleService(
	log *clog.Log,
	repo domain.ArticleRepoDomain,
	cfg *config.Config,
) (*ArticleService, error) {
	return &ArticleService{cfg: cfg, log: log, repo: repo}, nil
}

func (s *ArticleService) CreateArticle(ctx context.Context, req *articlepb.CreateArticleRequest) (uint64, error) {
	a := &entity.Article{
		AuthorID:   req.GetAuthorID(),
		Title:      req.GetTitle(),
		Content:    req.GetContent(),
		Summary:    req.GetSummary(),
		CategoryID: req.GetCategoryID(),
		IsTop:      uint8(req.GetIsTop()),
		CoverImage: req.GetCoverImage(),
	}
	if err := s.repo.Create(ctx, a); err != nil {
		s.log.Error(err.Error())
		return 0, err
	}
	return a.ID, nil
}

func (s *ArticleService) GetArticle(ctx context.Context, id uint64) (*entity.Article, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ArticleService) ListArticles(ctx context.Context, page, pageSize int) ([]*entity.Article, uint64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}
