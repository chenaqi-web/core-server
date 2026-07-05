package entity

import "time"

type InteractionLikeEntity struct {
	ID         string         `json:"id"`
	CreatedAt  time.Time      `json:"-"`
	UpdatedAt  time.Time      `json:"-"`
	UserID     string         `json:"user_id"`
	ObjectType ObjectType     `json:"object_type"`
	ObjectID   string         `json:"object_id"`
	Status     LikeStatusType `json:"status"`
	Version    int64          `json:"version"`
}

//======================================================================================================================
// 点赞状态机

type LikeStatusType string

const (
	LikeStatusTypeUnknown   LikeStatusType = "unknown"
	LikeStatusTypeThumbUp   LikeStatusType = "thumb_up"
	LikeStatusTypeThumbDown LikeStatusType = "thumb_down"
	LikeStatusTypeNothing   LikeStatusType = "nothing" // 设计此状态是为了避免频繁删除数据
)

func (s LikeStatusType) String() string {
	return string(s)
}

func ParseLikeStatusType(s string) LikeStatusType {
	switch s {
	case "thumb_up":
		return LikeStatusTypeThumbUp
	case "thumb_down":
		return LikeStatusTypeThumbDown
	case "nothing":
		return LikeStatusTypeNothing
	default:
		return LikeStatusTypeUnknown
	}
}

// =====================================================================================================================
// 点赞的对象类型

type ObjectType string

const (
	ObjectTypeUnknown ObjectType = "unknown"
	ObjectTypeArticle ObjectType = "article"
	ObjectTypeLife    ObjectType = "life"
)

func (o ObjectType) String() string {
	return string(o)
}

func ParseObjectType(s string) ObjectType {
	switch s {
	case "article":
		return ObjectTypeArticle
	case "life":
		return ObjectTypeLife
	default:
		return ObjectTypeUnknown
	}
}
