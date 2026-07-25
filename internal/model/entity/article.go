package entity

import (
	"database/sql"
	"time"
)

// Article maps to `blog_posts` table in SQL.
type Article struct {
	ID        uint64       `db:"id"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`

	Title      string `db:"title"`
	Summary    string `db:"summary"`
	Content    string `db:"content"`
	CoverImage string `db:"cover_image"`

	AuthorID     uint64 `db:"author_id"`
	CategoryID   uint64 `db:"category_id"`
	IsTop        uint8  `db:"is_top"`
	ViewCount    uint64 `db:"view_count"`
	LikeCount    uint64 `db:"like_count"`
	CommentCount uint64 `db:"comment_count"`
}

func (Article) TableName() string { return "blog_article" }
