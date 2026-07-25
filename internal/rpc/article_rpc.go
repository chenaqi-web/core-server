package rpc

import (
	"context"

	"backend/core-server/internal/application"
	"backend/core-server/internal/model/entity"
	"backend/core-server/internal/rpc/articlepb"
)

type ArticleRPC struct {
	articlepb.UnimplementedArticleServiceServer
	ArticleService *application.ArticleService
}

func NewArticleRPC(articleService *application.ArticleService) *ArticleRPC {
	return &ArticleRPC{ArticleService: articleService}
}

func (a *ArticleRPC) CreateArticle(ctx context.Context, req *articlepb.CreateArticleRequest) (*articlepb.CreateArticleResponse, error) {
	err := a.ArticleService.CreateArticle(ctx, req)
	if err != nil {
		return nil, err
	}
	return &articlepb.CreateArticleResponse{Success: true}, nil
}

func (a *ArticleRPC) GetArticle(ctx context.Context, req *articlepb.GetArticleRequest) (*articlepb.GetArticleResponse, error) {
	art, err := a.ArticleService.GetArticle(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if art == nil {
		return &articlepb.GetArticleResponse{}, nil
	}
	return &articlepb.GetArticleResponse{Article: toArticlePB(art)}, nil
}

func (a *ArticleRPC) ListArticles(ctx context.Context, req *articlepb.ListArticlesRequest) (*articlepb.ListArticlesResponse, error) {
	articles, err := a.ArticleService.ListArticles(ctx, int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	items := make([]*articlepb.Article, 0, len(articles))
	for _, art := range articles {
		items = append(items, toArticlePB(art))
	}
	total := uint64(len(items))
	return &articlepb.ListArticlesResponse{Articles: items, Total: total}, nil
}

func (a *ArticleRPC) DeleteArticle(ctx context.Context, req *articlepb.DeleteArticleRequest) (*articlepb.DeleteArticleResponse, error) {
	if err := a.ArticleService.DeleteArticle(ctx, req.GetId(), req.GetAuthorID()); err != nil {
		return nil, err
	}
	return &articlepb.DeleteArticleResponse{Success: true}, nil
}

func (a *ArticleRPC) ListMyArticles(ctx context.Context, req *articlepb.ListMyArticlesRequest) (*articlepb.ListMyArticlesResponse, error) {
	articles, err := a.ArticleService.ListMyArticles(ctx, req.GetAuthorID(), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	items := make([]*articlepb.Article, 0, len(articles))
	for _, art := range articles {
		items = append(items, toArticlePB(art))
	}
	total := uint64(len(items))
	return &articlepb.ListMyArticlesResponse{Articles: items, Total: total}, nil
}

func (a *ArticleRPC) ListByCategory(ctx context.Context, req *articlepb.ListByCategoryRequest) (*articlepb.ListByCategoryResponse, error) {
	articles, err := a.ArticleService.ListArticlesByCategory(ctx, req.GetCategoryID(), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	items := make([]*articlepb.Article, 0, len(articles))
	for _, art := range articles {
		items = append(items, toArticlePB(art))
	}
	total := uint64(len(items))
	return &articlepb.ListByCategoryResponse{Articles: items, Total: total}, nil
}

func (a *ArticleRPC) SearchArticles(ctx context.Context, req *articlepb.SearchArticlesRequest) (*articlepb.SearchArticlesResponse, error) {
	articles, err := a.ArticleService.SearchArticles(ctx, req.GetQ(), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	items := make([]*articlepb.Article, 0, len(articles))
	for _, art := range articles {
		items = append(items, toArticlePB(art))
	}
	total := uint64(len(items))
	return &articlepb.SearchArticlesResponse{Articles: items, Total: total}, nil
}

func toArticlePB(a *entity.Article) *articlepb.Article {
	return &articlepb.Article{
		Id:           a.ID,
		AuthorID:     a.AuthorID,
		Title:        a.Title,
		Content:      a.Content,
		Summary:      a.Summary,
		CategoryID:   a.CategoryID,
		IsTop:        a.IsTop,
		CoverImage:   a.CoverImage,
		ViewCount:    a.ViewCount,
		LikeCount:    a.LikeCount,
		CommentCount: a.CommentCount,
		CreatedAt:    uint64(a.CreatedAt.Unix()),
	}
}
