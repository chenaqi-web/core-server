package infras

import (
	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/cache"
	"backend/core-server/internal/infras/clog"
	"backend/core-server/internal/infras/mq/kafka"
	"backend/core-server/internal/infras/repo"

	"github.com/google/wire"
)

var RepoProviderSet = wire.NewSet(
	repo.NewDBClient,
	repo.NewLikeRepo,
	repo.NewCountRepo,
	// todo 新操作

	wire.Bind(new(domain.LikeRepoDomain), new(*repo.LikeRepo)),
	wire.Bind(new(domain.CountRepoDomain), new(*repo.CountRepo)),
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
