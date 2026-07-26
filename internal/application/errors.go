package application

import "errors"

var (
	ErrAlreadyLiked  = errors.New("already liked")
	ErrLikeNotExists = errors.New("like not exists")

	ErrCategoryTypeNotFound = errors.New("category type not found")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrArticleNotFound      = errors.New("article not found")
	ErrArticleForbidden     = errors.New("article delete forbidden")

	ErrInvalidAuthInput     = errors.New("invalid auth input")
	ErrInvalidEmail         = errors.New("invalid email")
	ErrInvalidEmailPurpose  = errors.New("invalid email code purpose")
	ErrPasswordRule         = errors.New("invalid password")
	ErrUsernameExists       = errors.New("username already exists")
	ErrEmailExists          = errors.New("email already exists")
	ErrEmailCodeInvalid     = errors.New("email code is invalid or expired")
	ErrEmailCodeCooldown    = errors.New("email code was sent too recently")
	ErrEmailCodeHourlyLimit = errors.New("email code hourly limit reached")
	ErrMailUnavailable      = errors.New("mail service unavailable")
)
