package event

type Message struct {
	UserID    string `json:"user_id"`
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
