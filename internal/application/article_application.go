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

func (s *ArticleService) DeleteArticle(ctx context.Context, id uint64, authorID uint64) error {
	if err := s.repo.DeleteByID(ctx, id, authorID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrArticleNotFound
		}
		return err
	}
	return nil
}

func (s *ArticleService) GetArticle(ctx context.Context, id uint64) (*entity.Article, error) {
	res, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	return res, nil
}

func (s *ArticleService) ListArticles(ctx context.Context, page, pageSize int) ([]*entity.Article, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	list, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	return list, nil
}

func (s *ArticleService) ListMyArticles(ctx context.Context, authorID uint64, page, pageSize int) ([]*entity.Article, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListByAuthor(ctx, authorID, offset, pageSize)
}

func (s *ArticleService) ListArticlesByCategory(ctx context.Context, categoryID uint64, page, pageSize int) ([]*entity.Article, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListByCategory(ctx, categoryID, offset, pageSize)
}

func (s *ArticleService) SearchArticles(ctx context.Context, q string, page, pageSize int) ([]*entity.Article, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.Search(ctx, q, offset, pageSize)
}
