//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"core-server/internal/application"
	"core-server/internal/config"
	"core-server/internal/infras"
	"core-server/internal/jobs"
	"core-server/internal/rpc"
)

//go:generate go run github.com/google/wire/cmd/wire

func InitializeServer(cfg *config.Config) (*rpc.Server, error) {
	wire.Build(
		infras.JobProviderSet,
		jobs.ProviderSet,
		application.ProviderSet,
		rpc.ProviderSet,
	)
	return nil, nil
}
