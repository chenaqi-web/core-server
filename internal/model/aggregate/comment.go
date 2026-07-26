package aggregate

import "backend/core-server/internal/model/entity"

type CommentAggregate struct {
	Comment *entity.Comment
	Author  *entity.User
}

func NewCommentAggregate(comment *entity.Comment, author *entity.User) *CommentAggregate {
	return &CommentAggregate{
		Comment: comment,
		Author:  author,
	}
}
