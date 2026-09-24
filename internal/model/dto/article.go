package dto

import (
	"core-server/internal/model/aggregate"
	"time"
)

type CreateArticleRequest struct {
	AuthorID    uint64
	Title       string
	Summary     string
	Content     string
	CoverImage  string
	CategoryID  uint64
	IsTop       bool
	IsPublished bool
}

type EditorArticleRequest struct {
	ID          uint64
	AuthorID    uint64
	Title       string
	Summary     string
	Content     string
	CoverImage  string
	CategoryID  uint64
	IsTop       bool
	IsPublished bool
}

type ArticleResponse struct {
	ID           uint64
	Title        string
	Summary      string
	Content      string
	CoverImage   string
	AuthorID     uint64
	CategoryID   uint64
	IsTop        bool
	IsPublished  bool
	ViewCount    uint64
	LikeCount    uint64
	FavorCount   uint64
	CommentCount uint64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PublishedAt  time.Time
	AuthorName   string
	AuthorAvatar string
}

type CreateArticleResponse struct{ Success bool }
type EditorArticleResponse struct {
	Success   bool
	ArticleID uint64
}
type GetArticleResponse struct{ Article *ArticleResponse }
type ListArticlesResponse struct{ Articles []*ArticleResponse }
type DeleteArticleResponse struct{ Success bool }
type ListMyArticlesResponse struct{ Articles []*ArticleResponse }
type ListArticlesByCategoryResponse struct{ Articles []*ArticleResponse }
type SearchArticlesResponse struct{ Articles []*ArticleResponse }

type GetArticleRequest struct{ ID uint64 }
type ListArticlesRequest struct{ Page, PageSize int }
type ListMyArticlesRequest struct {
	AuthorID       uint64
	Page, PageSize int
}
type ListArticlesByCategoryRequest struct {
	CategoryID     uint64
	Page, PageSize int
}
type SearchArticlesRequest struct {
	Query          string
	Page, PageSize int
}

func ToArticleResponse(agg *aggregate.ArticleAggregate) *ArticleResponse {
	if agg == nil || agg.Article == nil {
		return nil
	}
	a := agg.Article
	r := &ArticleResponse{ID: a.ID, Title: a.Title, Summary: a.Summary, Content: a.Content, CoverImage: a.CoverImage,
		AuthorID: a.AuthorID, CategoryID: a.CategoryID, IsTop: a.IsTop, IsPublished: a.IsPublished,
		ViewCount: a.ViewCount, LikeCount: a.LikeCount, FavorCount: a.FavorCount,
		CommentCount: a.CommentCount, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt}
	if a.PublishedAt.Valid {
		r.PublishedAt = a.PublishedAt.Time
	}
	if agg.Stats != nil {
		r.ViewCount = agg.Stats.ViewCount
		r.LikeCount = agg.Stats.LikeCount
		r.CommentCount = agg.Stats.CommentCount
	}
	if agg.Author != nil {
		r.AuthorName = agg.Author.Name
		r.AuthorAvatar = agg.Author.Avatar
	}
	return r
}

func ToArticleResponses(aggs []*aggregate.ArticleAggregate) []*ArticleResponse {
	items := make([]*ArticleResponse, 0, len(aggs))
	for _, agg := range aggs {
		if item := ToArticleResponse(agg); item != nil {
			items = append(items, item)
		}
	}
	return items
}

func ToGetArticleResponse(agg *aggregate.ArticleAggregate) *GetArticleResponse {
	return &GetArticleResponse{Article: ToArticleResponse(agg)}
}

func ToListArticlesResponse(aggs []*aggregate.ArticleAggregate) *ListArticlesResponse {
	return &ListArticlesResponse{Articles: ToArticleResponses(aggs)}
}

func ToListMyArticlesResponse(aggs []*aggregate.ArticleAggregate) *ListMyArticlesResponse {
	return &ListMyArticlesResponse{Articles: ToArticleResponses(aggs)}
}

func ToListArticlesByCategoryResponse(aggs []*aggregate.ArticleAggregate) *ListArticlesByCategoryResponse {
	return &ListArticlesByCategoryResponse{Articles: ToArticleResponses(aggs)}
}

func ToSearchArticlesResponse(aggs []*aggregate.ArticleAggregate) *SearchArticlesResponse {
	return &SearchArticlesResponse{Articles: ToArticleResponses(aggs)}
}
