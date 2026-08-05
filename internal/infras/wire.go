package infras

import (
	"core-server/internal/domain"
	"core-server/internal/infras/cache"
	"core-server/internal/infras/clog"
	"core-server/internal/infras/mail"
	"core-server/internal/infras/mq/kafka"
	"core-server/internal/infras/repo"

	"github.com/google/wire"
)

var RepoProviderSet = wire.NewSet(
	repo.NewDBClient,
	repo.NewLikeRepo,
	repo.NewCountRepo,
	repo.NewUserRepo,
	repo.NewCategoryRepo,
	repo.NewArticleRepo,
	repo.NewCommentRepo,
	// todo 新操作

	wire.Bind(new(domain.LikeRepoDomain), new(*repo.LikeRepo)),
	wire.Bind(new(domain.CountRepoDomain), new(*repo.CountRepo)),
	wire.Bind(new(domain.UserRepoDomain), new(*repo.UserRepo)),
	wire.Bind(new(domain.CategoryRepoDomain), new(*repo.CategoryRepo)),
	wire.Bind(new(domain.ArticleRepoDomain), new(*repo.ArticleRepo)),
	wire.Bind(new(domain.CommentRepoDomain), new(*repo.CommentRepo)),
)

var CacheProviderSet = wire.NewSet(
	cache.NewClient,
	cache.NewILikeCache,
	// 新接口

	wire.Bind(new(domain.LikeCacheDomain), new(*cache.ILikeCache)),
)

var MQProviderSet = wire.NewSet(
	kafka.NewSyncProducer,
	kafka.NewTopicManager,
	kafka.NewKafkaManager,
)

var LogProviderSet = wire.NewSet(
	clog.NewLog,
)

var JobProviderSet = wire.NewSet(
	RepoProviderSet,
	CacheProviderSet,
	MQProviderSet,
	LogProviderSet,
	mail.NewSender,
)
