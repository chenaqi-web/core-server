package entity

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID          uint64       `db:"id" json:"id"`
	ArticleID   uint64       `db:"article_id" json:"articleId"`
	UserID      uint64       `db:"user_id" json:"userId"`
	ParentID    uint64       `db:"parent_id" json:"parentId"`
	RootID      uint64       `db:"root_id" json:"rootId"`
	ReplyToID   uint64       `db:"reply_to_id" json:"replyToId"`
	ReplyToName string       `db:"reply_to_name" json:"replyToName"`
	Content     string       `db:"content" json:"content"`
	LikeCount   uint32       `db:"like_count" json:"likeCount"`
	ChildCount  uint32       `db:"child_count" json:"childCount"`
	CreatedAt   time.Time    `db:"created_at" json:"createdAt"`
	DeletedAt   sql.NullTime `db:"deleted_at" json:"deletedAt"`
}

func (Comment) TableName() string {
	return "comment"
}

func (c *Comment) IsTopLevel() bool {
	return c.ParentID == 0
}
