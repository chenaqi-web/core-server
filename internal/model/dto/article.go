package dto

import (
	"time"

	"backend/core-server/internal/model/entity"
)

// AuthorDTO 文章作者展示信息，来源于 user 表，不含 password 等敏感字段。
type AuthorDTO struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
}

// ArticleDTO 文章展示数据传输对象，聚合文章本体与作者信息。
type ArticleDTO struct {
	ID           uint64     `json:"id"`
	Title        string     `json:"title"`
	Summary      string     `json:"summary"`
	Content      string     `json:"content"`
	CoverImage   string     `json:"cover_image"`
	CategoryID   uint64     `json:"category_id"`
	IsTop        bool       `json:"is_top"`
	ViewCount    uint64     `json:"view_count"`
	LikeCount    uint64     `json:"like_count"`
	CommentCount uint64     `json:"comment_count"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Author       *AuthorDTO `json:"author"`
}

func AuthorFromEntity(u *entity.User) *AuthorDTO {
	if u == nil {
		return nil
	}
	return &AuthorDTO{
		ID:     u.ID,
		Name:   u.Name,
		Avatar: u.Avatar,
		Email:  u.Email,
		Role:   u.Role,
	}
}

func ArticleFromEntity(a *entity.Article, author *entity.User) *ArticleDTO {
	if a == nil {
		return nil
	}
	return &ArticleDTO{
		ID:           a.ID,
		Title:        a.Title,
		Summary:      a.Summary,
		Content:      a.Content,
		CoverImage:   a.CoverImage,
		CategoryID:   a.CategoryID,
		IsTop:        a.IsTop,
		ViewCount:    a.ViewCount,
		LikeCount:    a.LikeCount,
		CommentCount: a.CommentCount,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
		Author:       AuthorFromEntity(author),
	}
}

func ArticlesFromEntities(articles []*entity.Article, authors map[uint64]*entity.User) []*ArticleDTO {
	if len(articles) == 0 {
		return nil
	}
	items := make([]*ArticleDTO, 0, len(articles))
	for _, a := range articles {
		var author *entity.User
		if authors != nil {
			author = authors[a.AuthorID]
		}
		items = append(items, ArticleFromEntity(a, author))
	}
	return items
}
