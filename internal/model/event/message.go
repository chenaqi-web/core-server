package event

import "backend/core-server/internal/model/entity"

// Message 消息的实体
type Message struct {
	UserID    string `json:"user_id"` // 用作kafka中的partition key
	EventType string `json:"event_type"`
	Body      []byte `json:"body"`
}

type EventUserThumbUp struct {
	Timestamp     int64  `json:"timestamp"`
	UserID        string `json:"user_id"`
	ObjectID      string `json:"object_id"`
	ObjectType    string `json:"object_type"`
	ObjectOwnerID string `json:"object_owner_id"`
	Status        string `json:"status"`
}

type EventUserCancelThumbUp struct {
	Timestamp        int64  `json:"timestamp"`
	UserID           string `json:"user_id"`
	ObjectID         string `json:"object_id"`
	ObjectType       string `json:"object_type"`
	ObjectOwnerID    string `json:"object_owner_id"`
	IsDeletedInCache int    `json:"is_deleted_in_cache"`
}

type EventFavorCountsDataSync struct {
	ObjectType string           `json:"object_type"`
	ID2Count   map[string]int64 `json:"id_to_count"`
}

type EventFollowCountsDataSync struct {
	ID2Count map[string]int64 `json:"id_to_count"`
}

type EventLikeInteractionsDataSync struct {
	UserID       string                    `json:"user_id"`
	ObjectType   string                    `json:"object_type"`
	Interactions []*entity.InteractionLike `json:"interactions"`
}

type EventLikeCountsDataSync struct {
	ObjectType string           `json:"object_type"`
	ID2Count   map[string]int64 `json:"id_to_count"`
}

type EventUserComment struct {
	Timestamp     int64  `json:"timestamp"`
	UserID        string `json:"user_id"`
	ObjectID      string `json:"object_id"`
	ObjectType    string `json:"object_type"`
	ObjectOwnerID string `json:"object_owner_id"`
	Status        string `json:"status"`
}

type EventUserCancelComment struct {
	Timestamp        int64  `json:"timestamp"`
	UserID           string `json:"user_id"`
	ObjectID         string `json:"object_id"`
	ObjectType       string `json:"object_type"`
	ObjectOwnerID    string `json:"object_owner_id"`
	IsDeletedInCache int    `json:"is_deleted_in_cache"`
}

type EventCommentCountsDataSync struct {
	ObjectType string           `json:"object_type"`
	ID2Count   map[string]int64 `json:"id_to_count"`
}

type EventCommentInteractionsDataSync struct {
	UserID     string `json:"user_id"`
	ObjectType string `json:"object_type"`
	//Interactions []*entity.InteractionComment `json:"interactions"`
}
