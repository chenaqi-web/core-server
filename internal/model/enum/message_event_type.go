package enum

type MessageEventType string

const (
	MessageEventTypeUnknown           MessageEventType = "unknown"
	MessageEventTypeUserThumbUp       MessageEventType = "user_thumb_up"
	MessageEventTypeUserCancelThumbUp MessageEventType = "user_cancel_thumb_up"
)

func (t MessageEventType) String() string {
	return string(t)
}

func ParseMessageEventType(s string) MessageEventType {
	switch s {
	case "user_thumb_up":
		return MessageEventTypeUserThumbUp
	case "user_cancel_thumb_up":
		return MessageEventTypeUserCancelThumbUp
	default:
		return MessageEventTypeUnknown
	}
}
