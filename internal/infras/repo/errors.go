package repo

import "errors"

var ErrNotFound = errors.New("record not found")
var (
	ErrUserNotFound = errors.New("user not found")
)
