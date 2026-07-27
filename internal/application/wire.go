package application

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewAuthService,
	NewUserService,
	NewLikeService,
	NewCategoryService,
	NewArticleService,
)
