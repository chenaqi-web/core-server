package enum

// =====================================================================================================================
// 点赞的对象类型

type ObjectType string

const (
	ObjectTypeUnknown ObjectType = "unknown"
	ObjectTypeArticle ObjectType = "article"
	ObjectTypeLife    ObjectType = "life"
	ObjectTypeComment ObjectType = "comment"
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
	case "comment":
		return ObjectTypeComment
	default:
		return ObjectTypeUnknown
	}
}
