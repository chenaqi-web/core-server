package application

import (
	"backend/core-server/internal/application/user_svc"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	user_svc.NewUserService,
)
