package application

import "errors"

var (
	ErrAlreadyLiked  = errors.New("already liked")
	ErrLikeNotExists = errors.New("like not exists")

	ErrCategoryTypeNotFound = errors.New("category type not found")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrArticleNotFound      = errors.New("article not found")
	ErrArticleForbidden     = errors.New("article delete forbidden")
)
