package cache

import "errors"

var (
	ErrKeyNotFound          = errors.New("cache key not found")
	ErrLuaScriptExecFailure = errors.New("lua script execution failure")
)
