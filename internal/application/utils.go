package application

import (
	"errors"
)

const (
	defaultPageSize   = 10
	previewReplyLimit = 3
)

var (
	ErrAlreadyLiked = errors.New("already liked")

	ErrNotFound             = errors.New("not found")
	ErrCategoryTypeNotFound = errors.New("category type not found")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrArticleNotFound      = errors.New("article not found")

	ErrCommentNotFound = errors.New("comment not found")
)

func Page(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func Size(size int) int {
	if size <= 0 {
		return defaultPageSize
	}
	return size
}
