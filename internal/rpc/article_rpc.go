package rpc

import (
	"context"

	"core-server/internal/application"
	"core-server/internal/model/dto"
	"core-server/internal/rpc/articlepb"
)

type ArticleRPC struct {
	articlepb.UnimplementedArticleServiceServer
	ArticleService *application.ArticleService
}

func NewArticleRPC(articleService *application.ArticleService) *ArticleRPC {
	return &ArticleRPC{ArticleService: articleService}
}

func (a *ArticleRPC) CreateArticle(ctx context.Context, req *articlepb.CreateArticleRequest) (*articlepb.CreateArticleResponse, error) {
	err := a.ArticleService.CreateArticle(ctx, &dto.CreateArticleRequest{
		AuthorID:    req.GetAuthorID(),
		Title:       req.GetTitle(),
		Summary:     req.GetSummary(),
		Content:     req.GetContent(),
		CoverImage:  req.GetCoverImage(),
		CategoryID:  req.GetCategoryID(),
		IsTop:       req.GetIsTop(),
		IsPublished: req.GetIsPublished(),
	})
	if err != nil {
		return nil, err
	}
	return &articlepb.CreateArticleResponse{Success: true}, nil
}

func (a *ArticleRPC) EditorArticle(ctx context.Context, req *articlepb.EditorArticleRequest) (*articlepb.EditorArticleResponse, error) {
	err := a.ArticleService.EditorArticle(ctx, &dto.EditorArticleRequest{
		ID:          req.GetId(),
		AuthorID:    req.GetAuthorID(),
		Title:       req.GetTitle(),
		Summary:     req.GetSummary(),
		Content:     req.GetContent(),
		CoverImage:  req.GetCoverImage(),
		CategoryID:  req.GetCategoryID(),
		IsTop:       req.GetIsTop(),
		IsPublished: req.GetIsPublished(),
	})
	if err != nil {
		return nil, err
	}
	return &articlepb.EditorArticleResponse{
		Success:   true,
		ArticleID: req.GetId(),
	}, nil
}

func (a *ArticleRPC) PublishDraft(ctx context.Context, req *articlepb.PublishDraftRequest) (*articlepb.PublishDraftResponse, error) {
	if err := a.ArticleService.PublishDraft(ctx, &dto.PublishDraftRequest{
		ID:       req.GetId(),
		AuthorID: req.GetAuthorID(),
	}); err != nil {
		return nil, err
	}
	return &articlepb.PublishDraftResponse{Success: true}, nil
}

func (a *ArticleRPC) GetArticle(ctx context.Context, req *articlepb.GetArticleRequest) (*articlepb.GetArticleResponse, error) {
	res, err := a.ArticleService.GetArticle(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return &articlepb.GetArticleResponse{
		Id:           res.ID,
		Title:        res.Title,
		Summary:      res.Summary,
		Content:      res.Content,
		CoverImage:   res.CoverImage,
		AuthorID:     res.AuthorID,
		CategoryID:   res.CategoryID,
		IsTop:        res.IsTop,
		IsPublished:  res.IsPublished,
		ViewCount:    res.ViewCount,
		LikeCount:    res.LikeCount,
		CommentCount: res.CommentCount,
		CreatedAt:    uint64(res.CreatedAt.Unix()),
		UpdatedAt:    uint64(res.UpdatedAt.Unix()),
		PublishedAt:  uint64(res.PublishedAt.Unix()),
		AuthorName:   res.AuthorName,
		AuthorAvatar: res.AuthorAvatar,
		FavorCount:   res.FavorCount,
	}, nil
}

func (a *ArticleRPC) ListArticles(ctx context.Context, req *articlepb.ListArticlesRequest) (*articlepb.ListArticlesResponse, error) {
	res, err := a.ArticleService.ListArticles(ctx, &dto.ListArticlesRequest{Page: int(req.GetPage()), PageSize: int(req.GetPageSize())})
	if err != nil {
		return nil, err
	}
	return &articlepb.ListArticlesResponse{Articles: toArticlePBList(res.Articles)}, nil
}

func (a *ArticleRPC) DeleteArticle(ctx context.Context, req *articlepb.DeleteArticleRequest) (*articlepb.DeleteArticleResponse, error) {
	if err := a.ArticleService.DeleteArticle(ctx, &dto.DelArticleRequest{
		ID:     req.GetId(),
		UserID: req.GetUserID(),
		Role:   req.GetRole(),
	}); err != nil {
		return nil, err
	}
	return &articlepb.DeleteArticleResponse{Success: true}, nil
}

func (a *ArticleRPC) DeleteDraft(ctx context.Context, req *articlepb.DeleteDraftRequest) (*articlepb.DeleteDraftResponse, error) {
	if err := a.ArticleService.DeleteDraft(ctx, &dto.DeleteDraftRequest{
		ID:       req.GetId(),
		AuthorID: req.GetAuthorID(),
	}); err != nil {
		return nil, err
	}
	return &articlepb.DeleteDraftResponse{Success: true}, nil
}

func (a *ArticleRPC) ListMyArticles(ctx context.Context, req *articlepb.ListMyArticlesRequest) (*articlepb.ListMyArticlesResponse, error) {
	res, err := a.ArticleService.ListMyArticles(ctx, &dto.ListMyArticlesRequest{
		AuthorID:    req.GetAuthorID(),
		Page:        int(req.GetPage()),
		PageSize:    int(req.GetPageSize()),
		IsPublished: req.IsPublished,
	})
	if err != nil {
		return nil, err
	}
	return &articlepb.ListMyArticlesResponse{Articles: toArticlePBList(res.Articles), Total: res.Total}, nil
}

func (a *ArticleRPC) ListByCategory(ctx context.Context, req *articlepb.ListByCategoryRequest) (*articlepb.ListByCategoryResponse, error) {
	res, err := a.ArticleService.ListArticlesByCategory(ctx, &dto.ListArticlesByCategoryRequest{CategoryID: req.GetCategoryID(), Page: int(req.GetPage()), PageSize: int(req.GetPageSize())})
	if err != nil {
		return nil, err
	}
	return &articlepb.ListByCategoryResponse{Articles: toArticlePBList(res.Articles)}, nil
}

func (a *ArticleRPC) SearchArticles(ctx context.Context, req *articlepb.SearchArticlesRequest) (*articlepb.SearchArticlesResponse, error) {
	res, err := a.ArticleService.SearchArticles(ctx, &dto.SearchArticlesRequest{Query: req.GetQ(), Page: int(req.GetPage()), PageSize: int(req.GetPageSize())})
	if err != nil {
		return nil, err
	}
	return &articlepb.SearchArticlesResponse{Articles: toArticlePBList(res.Articles)}, nil
}

func toArticlePBList(aggregates []*dto.ArticleMsg) []*articlepb.Article {
	if len(aggregates) == 0 {
		return nil
	}
	items := make([]*articlepb.Article, 0, len(aggregates))
	for _, item := range aggregates {
		items = append(items, toArticlePB(item))
	}
	return items
}

func toArticlePB(a *dto.ArticleMsg) *articlepb.Article {
	if a == nil {
		return nil
	}
	pb := &articlepb.Article{
		Id:           a.ID,
		AuthorID:     a.AuthorID,
		Title:        a.Title,
		Content:      a.Content,
		Summary:      a.Summary,
		CategoryID:   a.CategoryID,
		IsTop:        a.IsTop,
		CoverImage:   a.CoverImage,
		IsPublished:  a.IsPublished,
		ViewCount:    a.ViewCount,
		LikeCount:    a.LikeCount,
		FavorCount:   a.FavorCount,
		CommentCount: a.CommentCount,
		CreatedAt:    uint64(a.CreatedAt.Unix()),
		UpdatedAt:    uint64(a.UpdatedAt.Unix()),
		PublishedAt:  uint64(a.PublishedAt.Unix()),
	}
	pb.AuthorName = a.AuthorName
	pb.AuthorAvatar = a.AuthorAvatar
	return pb
}
