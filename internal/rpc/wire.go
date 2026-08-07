package rpc

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAuthRPC,
	NewUserRPC,
	NewLikeRPC,
	NewCategoryRPC,
	NewArticleRPC,
	NewCommentRPC,
	NewServer,
)
