//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"backend/core-server/internal/application"
	"backend/core-server/internal/config"
	"backend/core-server/internal/infras"
	"backend/core-server/internal/rpc"
)

//go:generate go run github.com/google/wire/cmd/wire

func InitializeServer(cfg *config.Config) (*rpc.Server, error) {
	wire.Build(
		infras.ProviderSet,
		application.ProviderSet,
		rpc.ProviderSet,
	)
	return nil, nil
}
