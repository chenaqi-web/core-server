package rpc

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewUserRPC,
	NewLikeRPC,
	NewCategoryRPC,
	NewArticleRPC,
	NewServer,
)
