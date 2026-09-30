package application

import (
	"context"
	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/model/dto"
	"core-server/internal/model/entity"
	"database/sql"
	"time"

	"go.uber.org/zap"
)

type ArticleService struct {
	log      *clog.Log
	cfg      *config.Config
	ArtRepo  domain.ArticleRepoDomain
	userRepo domain.UserRepoDomain
}

func NewArticleService(
	log *clog.Log,
	cfg *config.Config,
	ArtRepo domain.ArticleRepoDomain,
	userRepo domain.UserRepoDomain,
) (*ArticleService, error) {
	return &ArticleService{
		cfg:      cfg,
		log:      log,
		ArtRepo:  ArtRepo,
		userRepo: userRepo,
	}, nil
}

func (s *ArticleService) CreateArticle(ctx context.Context, req *dto.CreateArticleRequest) error {
	a := &entity.Article{
		AuthorID:    req.AuthorID,
		Title:       req.Title,
		Summary:     req.Summary,
		Content:     req.Content,
		CoverImage:  req.CoverImage,
		CategoryID:  req.CategoryID,
		IsTop:       req.IsTop,
		IsPublished: req.IsPublished,
	}

	// 用来记录这篇文章的状态
	if a.IsPublished {
		a.PublishedAt = sql.NullTime{
			Valid: true, // 表示当前这个字段有数据
			Time:  time.Now(),
		}
	}

	if err := s.ArtRepo.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.ArtRepo.Create(ctx, a); err != nil {
			return err
		}
		if a.IsPublished {
			return s.userRepo.IncrementArticleCount(ctx, a.AuthorID)
		}
		return nil
	}); err != nil {
		s.log.Error("CreateArticle error", zap.Error(err))
		return err
	}
	return nil
}

func (s *ArticleService) EditorArticle(ctx context.Context, req *dto.EditorArticleRequest) error {
	a := &entity.Article{
		ID:          req.ID,
		AuthorID:    req.AuthorID,
		Title:       req.Title,
		Content:     req.Content,
		Summary:     req.Summary,
		CategoryID:  req.CategoryID,
		IsTop:       req.IsTop,
		IsPublished: req.IsPublished,
		CoverImage:  req.CoverImage,
	}
	if err := s.ArtRepo.Edit(ctx, a); err != nil {
		s.log.Error("EditorArticle error", zap.Error(err))
		return err
	}
	return nil
}

func (s *ArticleService) DeleteArticle(ctx context.Context, req *dto.DelArticleRequest) error {
	// 这个删除是包含管理员删除的
	if err := s.ArtRepo.WithTransaction(ctx, func(ctx context.Context) error {
		res, err := s.ArtRepo.DeleteByID(ctx, req.ID, req.UserID, req.Role)
		if err != nil {
			return err
		}
		return s.userRepo.DecrementArticleCount(ctx, res.Article.AuthorID)
	}); err != nil {
		s.log.Error("DeleteArticle error", zap.Error(err))
		return err
	}
	return nil
}

// =====================================================================================================================
// 草稿箱

func (s *ArticleService) PublishDraft(ctx context.Context, req *dto.PublishDraftRequest) error {
	if err := s.ArtRepo.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.ArtRepo.PublishDraft(ctx, req.ID, req.AuthorID); err != nil {
			return err
		}
		return s.userRepo.IncrementArticleCount(ctx, req.AuthorID)
	}); err != nil {
		s.log.Error("PublishDraft error", zap.Error(err))
		return err
	}
	return nil
}

func (s *ArticleService) DeleteDraft(ctx context.Context, req *dto.DeleteDraftRequest) error {
	if err := s.ArtRepo.DeleteDraftByID(ctx, req.ID, req.AuthorID); err != nil {
		s.log.Error("DeleteDraft error", zap.Error(err))
		return err
	}
	return nil
}

// =====================================================================================================================
// 列表函数

func (s *ArticleService) GetArticle(ctx context.Context, id uint64) (*dto.GetArticleResponse, error) {
	res, err := s.ArtRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("GetArticle info", zap.Error(err))
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	return dto.ToGetArticleResponse(res), nil
}

func (s *ArticleService) ListArticles(ctx context.Context, req *dto.ListArticlesRequest) (*dto.ListArticlesResponse, error) {
	page := Page(req.Page)
	pageSize := Size(req.PageSize)
	offset := (page - 1) * pageSize

	articles, err := s.ArtRepo.List(ctx, offset, pageSize)
	if err != nil {
		s.log.Error("ListArticles error", zap.Error(err))
		return nil, err
	}

	return dto.ToListArticlesResponse(articles), nil
}

func (s *ArticleService) ListMyArticles(ctx context.Context, req *dto.ListMyArticlesRequest) (*dto.ListMyArticlesResponse, error) {
	page := Page(req.Page)
	size := Size(req.PageSize)
	offset := (page - 1) * size

	articles, err := s.ArtRepo.ListByAuthor(ctx, req.AuthorID, offset, size, req.IsPublished)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	total, err := s.ArtRepo.CountByAuthor(ctx, req.AuthorID, req.IsPublished)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}

	return dto.ToListMyArticlesResponse(articles, total), nil
}

func (s *ArticleService) ListArticlesByCategory(ctx context.Context, req *dto.ListArticlesByCategoryRequest) (*dto.ListArticlesByCategoryResponse, error) {
	page := Page(req.Page)
	size := Size(req.PageSize)
	offset := (page - 1) * size

	articles, err := s.ArtRepo.ListByCategory(ctx, req.CategoryID, offset, size)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	return dto.ToListArticlesByCategoryResponse(articles), nil
}

func (s *ArticleService) SearchArticles(ctx context.Context, req *dto.SearchArticlesRequest) (*dto.SearchArticlesResponse, error) {
	page := Page(req.Page)
	pageSize := Size(req.PageSize)
	offset := (page - 1) * pageSize

	articles, err := s.ArtRepo.Search(ctx, req.Query, offset, pageSize)
	if err != nil {
		s.log.Error(err.Error())
		return nil, err
	}
	return dto.ToSearchArticlesResponse(articles), nil
}
