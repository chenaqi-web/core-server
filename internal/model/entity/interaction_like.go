package entity

import (
	"backend/core-server/internal/model/enum"
	"time"
)

type InteractionLike struct {
	ID            string          `db:"id" json:"id"`
	CreatedAt     time.Time       `db:"created_at" json:"-"`
	UpdatedAt     time.Time       `db:"updated_at" json:"-"`
	UserID        string          `db:"user_id" json:"user_id"`
	ObjectType    enum.ObjectType `db:"object_type" json:"object_type"`
	ObjectID      string          `db:"object_id" json:"object_id"`
	ObjectOwnerID string          `db:"object_owner_id" json:"object_owner_id"`
	Status        LikeStatusType  `db:"status" json:"status"`
	Version       int64           `db:"version" json:"version"`
}

func (InteractionLike) TableName() string {
	return "interaction_like"
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
