package application

import (
	"backend/core-server/internal/infras/repo"
	"backend/core-server/internal/rpc/articlepb"
	"context"
	"errors"

	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/clog"
	"backend/core-server/internal/model/aggregate"
	"backend/core-server/internal/model/entity"
)

type ArticleService struct {
	cfg      *config.Config
	log      *clog.Log
	ArtRepo  domain.ArticleRepoDomain
	userRepo domain.UserRepoDomain
}

func NewArticleService(
	log *clog.Log,
	ArtRepo domain.ArticleRepoDomain,
	userRepo domain.UserRepoDomain,
	cfg *config.Config,
) (*ArticleService, error) {
	return &ArticleService{cfg: cfg, log: log, ArtRepo: ArtRepo, userRepo: userRepo}, nil
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
	if err := s.ArtRepo.Create(ctx, a); err != nil {
		s.log.Error(err.Error())
		return err
	}
	return nil
}

func (s *ArticleService) DeleteArticle(ctx context.Context, id uint64, authorID uint64) error {
	if err := s.ArtRepo.DeleteByID(ctx, id, authorID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			s.log.Error(err.Error())
			return ErrArticleNotFound
		}
		return err
	}
	return nil
}

func (s *ArticleService) GetArticle(ctx context.Context, id uint64) (*aggregate.ArticleAggregate, error) {
	article, err := s.ArtRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	if article == nil {
		return nil, nil
	}

	author, err := s.userRepo.GetByID(ctx, article.AuthorID)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	return aggregate.NewArticleAggregate(article, author), nil
}

func (s *ArticleService) ListArticles(ctx context.Context, page, pageSize int) ([]*aggregate.ArticleAggregate, error) {
	page = Page(page)
	pageSize = Size(pageSize)
	offset := (page - 1) * pageSize

	articles, err := s.ArtRepo.List(ctx, offset, pageSize)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	return s.buildArticleAggregates(ctx, articles)
}

func (s *ArticleService) ListMyArticles(ctx context.Context, authorID uint64, page, pageSize int) ([]*aggregate.ArticleAggregate, error) {
	page = Page(page)
	size := Size(pageSize)
	offset := (page - 1) * pageSize

	articles, err := s.ArtRepo.ListByAuthor(ctx, authorID, offset, size)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	if len(articles) == 0 {
		return nil, nil
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	items := make([]*aggregate.ArticleAggregate, 0, len(articles))
	for _, article := range articles {
		items = append(items, aggregate.NewArticleAggregate(article, author))
	}
	return items, nil
}

func (s *ArticleService) ListArticlesByCategory(ctx context.Context, categoryID uint64, page, pageSize int) ([]*aggregate.ArticleAggregate, error) {
	page = Page(page)
	size := Size(pageSize)
	offset := (page - 1) * size

	articles, err := s.ArtRepo.ListByCategory(ctx, categoryID, offset, size)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	return s.buildArticleAggregates(ctx, articles)
}

func (s *ArticleService) SearchArticles(ctx context.Context, q string, page, pageSize int) ([]*aggregate.ArticleAggregate, error) {
	page = Page(page)
	pageSize = Size(pageSize)
	offset := (page - 1) * pageSize

	articles, err := s.ArtRepo.Search(ctx, q, offset, pageSize)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	return s.buildArticleAggregates(ctx, articles)
}

// =====================================================================================================================
// 下面是辅助函数

func (s *ArticleService) buildArticleAggregates(ctx context.Context, articles []*entity.Article) ([]*aggregate.ArticleAggregate, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	authorMap, err := LoadUserMap(ctx, s.userRepo, CollectArticleAuthorIDs(articles))
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	items := make([]*aggregate.ArticleAggregate, 0, len(articles))
	for _, article := range articles {
		items = append(items, aggregate.NewArticleAggregate(article, authorMap[article.AuthorID]))
	}
	return items, nil
}
