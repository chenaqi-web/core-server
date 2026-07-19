package rpc

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewUserRPC,
	NewLikeRPC,
	NewServer,
)
