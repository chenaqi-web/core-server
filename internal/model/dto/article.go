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

type DelArticleRequest struct {
	ID     uint64
	UserID uint64
	Role   string
}

type ArticleMsg struct {
	ID          uint64
	Title       string
	Summary     string
	Content     string
	CoverImage  string
	IsTop       bool
	IsPublished bool

	AuthorID     uint64
	AuthorName   string
	AuthorAvatar string

	CategoryID   uint64
	CategoryName string

	ViewCount    uint64
	LikeCount    uint64
	FavorCount   uint64
	CommentCount uint64

	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt time.Time
}

type GetArticleResponse struct {
	*ArticleMsg
}

type ListArticlesResponse struct {
	Articles []*ArticleMsg
}
type ListMyArticlesResponse struct {
	Articles []*ArticleMsg
}
type ListArticlesByCategoryResponse struct {
	Articles []*ArticleMsg
}
type SearchArticlesResponse struct {
	Articles []*ArticleMsg
}

type ListArticlesRequest struct {
	Page     int
	PageSize int
}
type ListMyArticlesRequest struct {
	AuthorID uint64
	Page     int
	PageSize int
}
type ListArticlesByCategoryRequest struct {
	CategoryID uint64
	Page       int
	PageSize   int
}
type SearchArticlesRequest struct {
	Query    string
	Page     int
	PageSize int
}

func ToArticleResponse(agg *aggregate.ArticleAggregate) *ArticleMsg {
	if agg == nil || agg.Article == nil {
		return nil
	}
	a := agg.Article
	r := &ArticleMsg{
		ID:           a.ID,
		Title:        a.Title,
		Summary:      a.Summary,
		Content:      a.Content,
		CoverImage:   a.CoverImage,
		AuthorID:     a.AuthorID,
		CategoryID:   a.CategoryID,
		IsTop:        a.IsTop,
		IsPublished:  a.IsPublished,
		ViewCount:    a.ViewCount,
		LikeCount:    a.LikeCount,
		FavorCount:   a.FavorCount,
		CommentCount: a.CommentCount,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}

	if a.PublishedAt.Valid {
		r.PublishedAt = a.PublishedAt.Time
	}

	if agg.Cate != nil {
		r.CategoryName = agg.Cate.Name
	}

	if agg.Author != nil {
		r.AuthorName = agg.Author.Name
		r.AuthorAvatar = agg.Author.Avatar
	}
	return r
}

func ToArticleResponses(aggs []*aggregate.ArticleAggregate) []*ArticleMsg {
	items := make([]*ArticleMsg, 0, len(aggs))
	for _, agg := range aggs {
		if item := ToArticleResponse(agg); item != nil {
			items = append(items, item)
		}
	}
	return items
}

func ToGetArticleResponse(agg *aggregate.ArticleAggregate) *GetArticleResponse {
	return &GetArticleResponse{ToArticleResponse(agg)}
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
