package application

import (
	"backend/core-server/internal/infras/repo"
	"backend/core-server/internal/rpc/articlepb"
	"context"
	"errors"

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

func (s *ArticleService) CreateArticle(ctx context.Context, req *articlepb.CreateArticleRequest) error {
	a := &entity.Article{
		AuthorID:   req.GetAuthorID(),
		Title:      req.GetTitle(),
		Content:    req.GetContent(),
		Summary:    req.GetSummary(),
		CategoryID: req.GetCategoryID(),
		IsTop:      req.GetIsTop(),
		CoverImage: req.GetCoverImage(),
	}
	if err := s.repo.Create(ctx, a); err != nil {
		s.log.Error(err.Error())
		return err
	}
	return nil
}

func (s *ArticleService) GetArticle(ctx context.Context, id uint64) (*entity.Article, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ArticleService) ListArticles(ctx context.Context, page, pageSize int) ([]*entity.Article, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

func (s *ArticleService) DeleteArticle(ctx context.Context, id uint64, authorID uint64) error {
	if err := s.repo.DeleteByID(ctx, id, authorID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrArticleNotFound
		}
		return err
	}
	return nil
}
