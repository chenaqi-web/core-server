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
	id, err := a.ArticleService.CreateArticle(ctx, req)
	if err != nil {
		return nil, err
	}
	return &articlepb.CreateArticleResponse{Id: id}, nil
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
	articles, total, err := a.ArticleService.ListArticles(ctx, int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	items := make([]*articlepb.Article, 0, len(articles))
	for _, art := range articles {
		items = append(items, toArticlePB(art))
	}
	return &articlepb.ListArticlesResponse{Articles: items, Total: total}, nil
}

func toArticlePB(a *entity.Article) *articlepb.Article {
	return &articlepb.Article{
		Id:        a.ID,
		AuthorID:  a.AuthorID,
		Title:     a.Title,
		Content:   a.Content,
		CreatedAt: uint64(a.CreatedAt.Unix()),
	}
}
