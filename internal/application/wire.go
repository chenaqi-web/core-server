package application

import (
	"backend/core-server/internal/application/user_svc"

	"github.com/google/wire"

	"backend/core-server/internal/application/health_svc"
)

var ProviderSet = wire.NewSet(
	health_svc.New,
	user_svc.NewUserService,
)
