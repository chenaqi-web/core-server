package rpc

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAuthRPC,
	NewLikeRPC,
	NewCategoryRPC,
	NewArticleRPC,
	NewCommentRPC,
	NewServer,
)
