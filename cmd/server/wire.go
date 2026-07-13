//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"backend/core-server/internal/application"
	"backend/core-server/internal/config"
	"backend/core-server/internal/infras"
	"backend/core-server/internal/jobs"
	"backend/core-server/internal/rpc"
)

//go:generate go run github.com/google/wire/cmd/wire

func InitializeApp(cfg *config.Config) (*App, error) {
	wire.Build(
		infras.JobProviderSet,
		jobs.ProviderSet,
		application.ProviderSet,
		rpc.ProviderSet,
		NewApp,
	)
	return nil, nil
}
