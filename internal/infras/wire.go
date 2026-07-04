package infras

import (
	"backend/core-server/internal/infras/repo"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	repo.NewSQLClient,
	repo.NewUserPepo,
)
