package entity

import (
	"backend/core-server/internal/model/enum"
	"time"
)

// 点赞关系表

type InteractionLike struct {
	ID         string          `json:"id"`
	CreatedAt  time.Time       `json:"-"`
	UpdatedAt  time.Time       `json:"-"`
	UserID     string          `json:"user_id"`
	ObjectType enum.ObjectType `json:"object_type"`
	ObjectID   string          `json:"object_id"`
	Status     LikeStatusType  `json:"status"`
	Version    int64           `json:"version"`
}

//======================================================================================================================
// 点赞状态机

type LikeStatusType string

const (
	LikeStatusTypeUnknown LikeStatusType = "unknown"
	LikeStatusTypeThumbUp LikeStatusType = "thumb_up"
	LikeStatusTypeNothing LikeStatusType = "nothing" // 设计此状态是为了避免频繁删除数据
)

func (s LikeStatusType) String() string {
	return string(s)
}

func ParseLikeStatusType(s string) LikeStatusType {
	switch s {
	case "thumb_up":
		return LikeStatusTypeThumbUp
	case "nothing":
		return LikeStatusTypeNothing
	default:
		return LikeStatusTypeUnknown
	}
}
