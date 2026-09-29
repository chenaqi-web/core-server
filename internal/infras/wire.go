package infras

import (
	"core-server/internal/domain"
	"core-server/internal/infras/cache"
	"core-server/internal/infras/clog"
	"core-server/internal/infras/mq/kafka"
	"core-server/internal/infras/repo"
	"core-server/internal/infras/repov2"

	"github.com/google/wire"
)

var RepoProviderSet = wire.NewSet(
	repo.NewDBClient,

	// v2 ent 版本
	repov2.NewEntClient,
	repov2.NewLikeRepo,
	repov2.NewCountRepo,
	repov2.NewUserRepo,
	repov2.NewCategoryRepo,
	repov2.NewArticleRepo,
	repov2.NewCommentRepo,
	// todo 新操作

	wire.Bind(new(domain.LikeRepoDomain), new(*repov2.LikeRepo)),
	wire.Bind(new(domain.CountRepoDomain), new(*repov2.CountRepo)),
	wire.Bind(new(domain.UserRepoDomain), new(*repov2.UserRepo)),
	wire.Bind(new(domain.CategoryRepoDomain), new(*repov2.CategoryRepo)),
	wire.Bind(new(domain.ArticleRepoDomain), new(*repov2.ArticleRepo)),
	wire.Bind(new(domain.CommentRepoDomain), new(*repov2.CommentRepo)),
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
)
