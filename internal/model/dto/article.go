package dto

import (
	"time"
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

type ArticleInteractionCount struct {
	LikeCount    uint64 `json:"like_count"`
	CommentCount uint64 `json:"comment_count"`
	ViewCount    uint64 `json:"view_count"`
}
