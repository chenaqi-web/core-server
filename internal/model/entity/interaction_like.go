package entity

import (
	"backend/core-server/internal/model/enum"
	"time"
)

type InteractionLike struct {
	ID            string          `gorm:"primaryKey;size:64" json:"id"`
	CreatedAt     time.Time       `gorm:"comment:创建时间" json:"-"`
	UpdatedAt     time.Time       `gorm:"comment:更新时间" json:"-"`
	UserID        string          `gorm:"size:64;not null;uniqueIndex:uk_like_user_object,priority:1" json:"user_id"`
	ObjectType    enum.ObjectType `gorm:"size:32;not null;uniqueIndex:uk_like_user_object,priority:2" json:"object_type"`
	ObjectID      string          `gorm:"size:64;not null;uniqueIndex:uk_like_user_object,priority:3" json:"object_id"`
	ObjectOwnerID string          `gorm:"size:64;not null;default:'';index:idx_object_owner_id" json:"object_owner_id"`
	Status        LikeStatusType  `gorm:"size:32;not null" json:"status"`
	Version       int64           `gorm:"not null;default:0" json:"version"`
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
