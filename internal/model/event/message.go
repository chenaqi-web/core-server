package event

type Message struct {
	UserID    uint64 `json:"user_id"`
	EventType string `json:"event_type"`
	Body      []byte `json:"body"`
}

type EventUserThumbUp struct {
	Timestamp  int64  `json:"timestamp"`
	UserID     uint64 `json:"user_id"`
	ObjectID   uint64 `json:"object_id"`
	ObjectType string `json:"object_type"`
	Status     string `json:"status"`
}

type EventUserCancelThumbUp struct {
	Timestamp        int64  `json:"timestamp"`
	UserID           uint64 `json:"user_id"`
	ObjectID         uint64 `json:"object_id"`
	ObjectType       string `json:"object_type"`
	IsDeletedInCache int    `json:"is_deleted_in_cache"`
}
