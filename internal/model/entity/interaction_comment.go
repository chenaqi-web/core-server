package entity

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID          uint64       `json:"id"`
	ArticleID   uint64       `json:"articleId"`
	UserID      uint64       `json:"userId"`
	ParentID    uint64       `json:"parentId"`
	RootID      uint64       `json:"rootId"`
	ReplyToID   uint64       `json:"replyToId"`
	ReplyToName string       `json:"replyToName"`
	Content     string       `json:"content"`
	LikeCount   uint32       `json:"likeCount"`
	ChildCount  uint32       `json:"childCount"`
	CreatedAt   time.Time    `json:"created_at"`
	DeletedAt   sql.NullTime `json:"deleted_at"`
}

func (Comment) TableName() string {
	return "comment"
}
