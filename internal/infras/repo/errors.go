package repo

import "errors"

var ErrNotFound = errors.New("record not found")

var (
	ErrUserNameExists  = errors.New("user name already exists")
	ErrUserEmailExists = errors.New("user email already exists")
)
