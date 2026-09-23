package application

import (
	"context"
	"core-server/internal/config"
	"core-server/internal/domain"
	"core-server/internal/infras/clog"
	"core-server/internal/infras/repov2"
	"core-server/internal/model/aggregate"
	"core-server/internal/model/entity"
	"core-server/internal/rpc/articlepb"
	"database/sql"
	"errors"
	"time"

	"go.uber.org/zap"
)

type ArticleService struct {
	log       *clog.Log
	cfg       *config.Config
	ArtRepo   domain.ArticleRepoDomain
	userRepo  domain.UserRepoDomain
	countRepo *CountService
}

func NewArticleService(
	log *clog.Log,
	cfg *config.Config,
	ArtRepo domain.ArticleRepoDomain,
	userRepo domain.UserRepoDomain,
	countRepo *CountService,
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
		AuthorID:    req.GetAuthorID(),
		Title:       req.GetTitle(),
		Content:     req.GetContent(),
		Summary:     req.GetSummary(),
		CategoryID:  req.GetCategoryID(),
		IsTop:       req.GetIsTop(),
		IsPublished: req.GetIsPublished(),
		Visibility:  req.GetVisibility(),
		CoverImage:  req.GetCoverImage(),
	}

	if a.IsPublished {
		a.PublishedAt = sql.NullTime{
			Valid: true, // 表示当前这个字段有数据
			Time:  time.Now(),
		}
	}

	if err := s.ArtRepo.Create(ctx, a); err != nil {
		s.log.Error("CreateArticle error", zap.Error(err))
		return err
	}
	return nil
}

func (s *ArticleService) EditorArticle(ctx context.Context, req *articlepb.EditorArticleRequest, authorID uint64) error {
	if authorID != req.GetAuthorID() {
		// 说明当前用户不是文章的作者
		return errors.New("have no right to modify others' articles")
	}
	a := &entity.Article{
		ID:         req.GetId(),
		AuthorID:   req.GetAuthorID(),
		Title:      req.GetTitle(),
		Content:    req.GetContent(),
		Summary:    req.GetSummary(),
		CategoryID: req.GetCategoryID(),
		IsTop:      req.GetIsTop(),
		Visibility: req.GetVisibility(),
		CoverImage: req.GetCoverImage(),
	}
	if a.ID == 0 {
		return errors.New("have no right to modify others' articles")
	}
	if err := s.ArtRepo.Edit(ctx, a); err != nil {
		return err
	}
	return nil
}

func (s *ArticleService) DeleteArticle(ctx context.Context, id uint64, authorID uint64) error {
	// 这个删除是包含管理员删除的
	// 当其是管理员的时候，传入的authorID是0,在repo下面是不会加上这个authorId = ? 的判断条件
	if err := s.ArtRepo.DeleteByID(ctx, id, authorID); err != nil {
		if errors.Is(err, repov2.ErrNotFound) {
			return ErrArticleNotFound
		}
		s.log.Error("DeleteArticle error", zap.Error(err))
		return err
	}
	return nil
}

func (s *ArticleService) GetArticle(ctx context.Context, id uint64) (*aggregate.ArticleAggregate, error) {
	article, err := s.ArtRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repov2.ErrNotFound) {
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
		if errors.Is(err, repov2.ErrNotFound) {
			return nil, nil
		}
		s.log.Error("GetArticle info", zap.Error(err))
		return nil, err
	}

	counts, err := s.countRepo.GetArticleInteractionCount(ctx, article.ID)
	if err != nil {
		s.log.Error("GetArticleInteractionCount error", zap.Error(err))
		return nil, err
	}

	// todo 暂时先将这个浏览量增加函数放到这个博客获取的位置
	if err := s.countRepo.IncrementArticleView(ctx, article.ID); err != nil {
		s.log.Error("increment article view count error", zap.Error(err))
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

	return s.BuildArticleAggregates(ctx, articles)
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

	return s.BuildArticleAggregates(ctx, articles)
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
	return s.BuildArticleAggregates(ctx, articles)
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
	return s.BuildArticleAggregates(ctx, articles)
}

// =====================================================================================================================

// BuildArticleAggregates  对聚合根的构建
func (s *ArticleService) BuildArticleAggregates(ctx context.Context, arts []*entity.Article) ([]*aggregate.ArticleAggregate, error) {
	// 一次遍历同时收集作者ID和文章ID
	authorIDs := make([]uint64, 0, len(arts))
	articleIDs := make([]uint64, 0, len(arts))
	seenAuthors := make(map[uint64]struct{})

	for _, article := range arts {
		// 收集文章ID
		articleIDs = append(articleIDs, article.ID)

		// 收集作者ID（去重）
		if _, ok := seenAuthors[article.AuthorID]; !ok {
			seenAuthors[article.AuthorID] = struct{}{}
			authorIDs = append(authorIDs, article.AuthorID)
		}
	}

	// 批量加载用户信息
	users, err := s.userRepo.ListByIDs(ctx, authorIDs)
	if err != nil {
		return nil, err
	}
	authorMap := make(map[uint64]*entity.User, len(users))
	for _, user := range users {
		authorMap[user.ID] = user
	}

	// 批量获取文章互动统计数据
	statsMap, err := s.countRepo.BatchGetArticleInteractionCounts(ctx, articleIDs)
	if err != nil {
		return nil, err
	}

	// 构建聚合结果
	items := make([]*aggregate.ArticleAggregate, 0, len(arts))
	for _, article := range arts {
		items = append(items, aggregate.NewArticleAggregate(
			article,
			authorMap[article.AuthorID],
			statsMap[article.ID],
		))
	}

	return items, nil
}
