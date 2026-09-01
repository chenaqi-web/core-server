package dto

// QueryObjectCommentedCountInBatchRequest 批量查询实体被评论数量。
type QueryObjectCommentedCountInBatchRequest struct {
	ObjectType string
	ObjectIDs  []uint64
}

// QueryObjectCommentedCountInBatchResponse 批量查询实体被评论数量结果。
type QueryObjectCommentedCountInBatchResponse struct {
	Counts map[uint64]uint64
}
