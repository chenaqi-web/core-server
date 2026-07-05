package infras

import (
	"backend/core-server/internal/infras/cache"
	"backend/core-server/internal/infras/mq/kafka"
	"backend/core-server/internal/infras/repo"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	// database
	repo.NewDBClient,
	repo.NewUserPepo,

	// cache
	cache.NewClient,
	cache.NewILikeCache,

	// mq
	kafka.NewTopicManager, // 先注册Topic管理器
	kafka.NewKafkaManager, // Kafka管理对象(负责管理消费)
	kafka.NewSyncProducer, // 全局共用一个生产者
)
