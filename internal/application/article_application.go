package application

import (
	"context"
	"core-server/internal/infras/repo"
	"core-server/internal/rpc/articlepb"
	"database/sql"
	"errors"

	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/entity"

	"go.uber.org/zap"
)

type ArticleService struct {
	cfg       *config.Config
	log       *clog.Log
	ArtRepo   domain.ArticleRepoDomain
	userRepo  domain.UserRepoDomain
	countRepo *CountService
}

func NewArticleService(
	log *clog.Log,
	ArtRepo domain.ArticleRepoDomain,
	userRepo domain.UserRepoDomain,
	countRepo *CountService,
	cfg *config.Config,
) (*ArticleService, error) {
	return &ArticleService{
		cfg:       cfg,
		log:       log,
		ArtRepo:   ArtRepo,
		userRepo:  userRepo,
		countRepo: countRepo,
	}, nil
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
		s.log.Error("CreateArticle error", zap.Error(err))
		return err
	}
	return nil
}

func (s *ArticleService) DeleteArticle(ctx context.Context, id uint64, authorID uint64) error {
	// 普通用户的删除文章
	if err := s.ArtRepo.DeleteByID(ctx, id, authorID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			s.log.Info("DeleteArticle info", zap.Error(err))
			return ErrArticleNotFound
		}
		s.log.Error(err.Error())
		return err
	}
	return nil
}

func (s *ArticleService) GetArticle(ctx context.Context, id uint64) (*aggregate.ArticleAggregate, error) {
	article, err := s.ArtRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Info("GetArticle info", zap.Error(err))
			return nil, nil
		}
		s.log.Error("GetArticle info", zap.Error(err))
		return nil, err
	}
	if article == nil {
		return nil, nil
	}

	author, err := s.userRepo.GetByID(ctx, article.AuthorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Info("GetArticle info", zap.Error(err))
			return nil, nil
		}
		s.log.Error("GetArticle info", zap.Error(err))
		return nil, err
	}

	counts, err := s.countRepo.GetArticleInteractionCount(ctx, article.ID)
	if err != nil {
		s.log.Error("GetArticle interaction counts error", zap.Error(err))
		return nil, err
	}

	return aggregate.NewArticleAggregate(article, author, counts), nil
}

// =====================================================================================================================
// 列表函数

func (s *ArticleService) ListArticles(ctx context.Context, page, pageSize int) ([]*aggregate.ArticleAggregate, error) {
	page = Page(page)
	pageSize = Size(pageSize)
	offset := (page - 1) * pageSize

	articles, err := s.ArtRepo.List(ctx, offset, pageSize)
	if err != nil {
		s.log.Error("ListArticles error", zap.Error(err))
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

	return s.buildArticleAggregates(ctx, articles)
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
	articleIDs := make([]uint64, 0, len(articles))
	for _, article := range articles {
		articleIDs = append(articleIDs, article.ID)
	}
	statsMap, err := s.countRepo.BatchGetArticleInteractionCounts(ctx, articleIDs)
	if err != nil {
		s.log.Error("load article interaction counts error", zap.Error(err))
		return nil, err
	}

	items := make([]*aggregate.ArticleAggregate, 0, len(articles))
	for _, article := range articles {
		items = append(items, aggregate.NewArticleAggregate(article, authorMap[article.AuthorID], statsMap[article.ID]))
	}
	return items, nil
}
