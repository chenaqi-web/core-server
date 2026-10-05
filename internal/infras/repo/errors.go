package repo

import "errors"

var ErrNotFound = errors.New("record not found")
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrCategoryNotFound = errors.New("category not found")
	ErrArticleNotFound  = errors.New("article not found")
)
