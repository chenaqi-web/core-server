package rpc

import (
	"context"

	"backend/core-server/internal/application"
	"backend/core-server/internal/model/aggregate"
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
	agg, err := a.ArticleService.GetArticle(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if agg == nil {
		return &articlepb.GetArticleResponse{}, nil
	}
	return &articlepb.GetArticleResponse{Article: toArticlePB(agg)}, nil
}

func (a *ArticleRPC) ListArticles(ctx context.Context, req *articlepb.ListArticlesRequest) (*articlepb.ListArticlesResponse, error) {
	articles, err := a.ArticleService.ListArticles(ctx, int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	return &articlepb.ListArticlesResponse{Articles: toArticlePBList(articles)}, nil
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
	return &articlepb.ListMyArticlesResponse{Articles: toArticlePBList(articles)}, nil
}

func (a *ArticleRPC) ListByCategory(ctx context.Context, req *articlepb.ListByCategoryRequest) (*articlepb.ListByCategoryResponse, error) {
	articles, err := a.ArticleService.ListArticlesByCategory(ctx, req.GetCategoryID(), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	return &articlepb.ListByCategoryResponse{Articles: toArticlePBList(articles)}, nil
}

func (a *ArticleRPC) SearchArticles(ctx context.Context, req *articlepb.SearchArticlesRequest) (*articlepb.SearchArticlesResponse, error) {
	articles, err := a.ArticleService.SearchArticles(ctx, req.GetQ(), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	return &articlepb.SearchArticlesResponse{Articles: toArticlePBList(articles)}, nil
}

func toArticlePBList(aggregates []*aggregate.ArticleAggregate) []*articlepb.Article {
	if len(aggregates) == 0 {
		return nil
	}
	items := make([]*articlepb.Article, 0, len(aggregates))
	for _, agg := range aggregates {
		items = append(items, toArticlePB(agg))
	}
	return items
}

func toArticlePB(agg *aggregate.ArticleAggregate) *articlepb.Article {
	if agg == nil || agg.Article == nil {
		return nil
	}
	a := agg.Article
	pb := &articlepb.Article{
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
		UpdatedAt:    uint64(a.UpdatedAt.Unix()),
	}
	if agg.Author != nil {
		pb.AuthorName = agg.Author.Name
		pb.AuthorAvatar = agg.Author.Avatar
	}
	return pb
}
