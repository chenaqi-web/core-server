package enum

// 互动类型
type InteractionType string

const (
	InteractionTypeUnknown InteractionType = "unknown"
	InteractionTypeLike    InteractionType = "like"    // 点赞
	InteractionTypeComment InteractionType = "comment" // 评论
)

func (t InteractionType) String() string {
	return string(t)
}

func ParseInteractionType(s string) InteractionType {
	switch s {
	case "like":
		return InteractionTypeLike
	case "comment":
		return InteractionTypeComment
	default:
		return InteractionTypeUnknown
	}
}
