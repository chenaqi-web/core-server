package infras

import (
	"backend/core-server/internal/infras/cache"
	"backend/core-server/internal/infras/repo"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	// database
	repo.NewDBClient,
	repo.NewUserPepo,

	// cache
	cache.NewClient,
	cache.NewILikeCache,

	// mq

)
