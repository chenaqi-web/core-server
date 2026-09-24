package entity

import (
	"database/sql"
	"time"
)

// Article maps to `blog_posts` table in SQL.
type Article struct {
	ID          uint64       `db:"id"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`
	PublishedAt sql.NullTime `db:"published_at"` // 发布时间 如果为null的话就表示还未发布处于草稿状态

	Title      string `db:"title"`
	Summary    string `db:"summary"`
	Content    string `db:"content"`
	CoverImage string `db:"cover_image"`

	AuthorID   uint64   `db:"author_id"`
	Author     User     `db:"author"`
	CategoryID uint64   `db:"category_id"`
	Category   Category `db:"category"`

	IsTop       bool `db:"is_top"`
	IsPublished bool `db:"is_published"` // 是否发布 0为草稿箱 1为已发布

	ViewCount    uint64 `db:"view_count"`
	LikeCount    uint64 `db:"like_count"`
	FavorCount   uint64 `db:"favor_count"`
	CommentCount uint64 `db:"comment_count"`
}

func (Article) TableName() string { return "blog_article" }
