package jobdbsync

import (
	"backend/core-server/internal/infras/clog"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
	"backend/core-server/internal/infras/cache"
	"backend/core-server/internal/infras/mq/kafka"
	jobaggregator "backend/core-server/internal/jobs/job-aggregator"
	"backend/core-server/internal/model/enum"
	"backend/core-server/internal/model/event"

	"github.com/IBM/sarama"
	"github.com/avast/retry-go"
	"github.com/cespare/xxhash/v2"
	"go.uber.org/zap"
)

type MessageQueueConsumer struct {
	logger *clog.Log

	// kafka 消费协调器
	kafkaManager *kafka.KafkaManager

	// redis client
	redisClient *cache.CacheClient

	// database
	likeRepo  domain.LikeRepoDomain
	countRepo domain.CountRepoDomain

	// cache
	likeCache *cache.ILikeCache

	// 聚合器
	likeCountAggregator *jobaggregator.ObjectCountAggregator

	// 死信topic 和 死信生产者
	dlqTopic string
	producer *kafka.SyncProducer
}

// kafka消费者构建阶段

func NewMessageQueueConsumer(
	cfg *config.Config,
	logger *clog.Log,
	producer *kafka.SyncProducer,
	kafkaManager *kafka.KafkaManager,
	redisClient *cache.CacheClient,
	likeRepo domain.LikeRepoDomain,
	countRepo domain.CountRepoDomain,
	likeCache *cache.ILikeCache,
) *MessageQueueConsumer {

	consumer := &MessageQueueConsumer{
		logger:       logger,
		kafkaManager: kafkaManager,
		redisClient:  redisClient,
		likeRepo:     likeRepo,
		countRepo:    countRepo,
		likeCache:    likeCache,
		producer:     producer,
		dlqTopic:     cfg.Kafka.DlqTopicName(),
	}

	consumer.likeCountAggregator = jobaggregator.NewObjectCountAggregator(
		logger,
		countRepo,
		cfg.CountAggregator.FlushDuration(),
		cfg.CountAggregator.BufferCapacity(),
		cfg.CountAggregator.DBDuration(),
	)

	return consumer
}

func (c *MessageQueueConsumer) Start() error {
	// 1. 加载消息处理handler
	c.kafkaManager.SetBatchHandler(c.batchHandleMessages)
	c.likeCountAggregator.Start()

	// 2.启动所有的消费者
	if err := c.kafkaManager.StartGroupConsumers(); err != nil {
		return err
	}

	log.Println("message queue consumer started")
	return nil
}

func (c *MessageQueueConsumer) Stop() error {
	c.likeCountAggregator.Stop()
	return c.kafkaManager.StopGroupConsumers()
}

// =====================================================================================================================

// batchHandleMessages Kafka任务管理器的消费函数
func (c *MessageQueueConsumer) batchHandleMessages(ctx context.Context, msgs []*sarama.ConsumerMessage) ([]bool, error) {
	canCommit := make([]bool, len(msgs))
	for i := range canCommit {
		canCommit[i] = true
	}
	if len(msgs) == 0 {
		return canCommit, nil
	}

	for idx, msg := range msgs {
		if len(msg.Value) == 0 {
			continue
		}

		hash := hashSaramaMessage(msg)
		hashKey := fmt.Sprintf("msg_%s", strconv.FormatUint(hash, 10))
		success, err := c.redisClient.Cache.SetNX(ctx, hashKey, "true", 10*time.Minute).Result()
		if err != nil {
			c.logger.Warn("redis dedup failed", zap.String("key", hashKey), zap.Error(err))
		} else if !success {
			continue
		}

		if err := c.handleMessageWrapper(ctx, msg, c.handleMessage); err != nil {
			_, _ = c.redisClient.Cache.Del(ctx, hashKey).Result()
			canCommit[idx] = false
			return canCommit, nil
		}
	}
	return canCommit, nil
}

func (c *MessageQueueConsumer) handleMessageWrapper(
	ctx context.Context,
	msg *sarama.ConsumerMessage,
	handler func(ctx context.Context, msg *event.Message) error,
) (err error) {
	defer func() {
		if err != nil && c.dlqTopic != "" {
			err = retry.Do(func() error {
				return c.producer.SendMessage(c.dlqTopic, string(msg.Key), msg.Value)
			},
				retry.Attempts(3),
				retry.MaxDelay(5*time.Second),
				retry.DelayType(retry.BackOffDelay),
			)
		}
	}()

	var entityMessage event.Message
	if err = json.Unmarshal(msg.Value, &entityMessage); err != nil {
		return err
	}
	return handler(ctx, &entityMessage)
}

func hashSaramaMessage(msg *sarama.ConsumerMessage) uint64 {
	hash := xxhash.New()
	hash.Write(msg.Key)
	hash.Write(msg.Value)
	return hash.Sum64()
}

func (c *MessageQueueConsumer) handleMessage(ctx context.Context, msg *event.Message) error {
	if msg.Body == nil {
		return nil
	}

	// 处理不同的消息类型
	switch enum.ParseMessageEventType(msg.EventType) {
	case enum.MessageEventTypeUserThumbUp:
		var message event.EventUserThumbUp
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}
		return c.handleUserLike(ctx, &message)
	case enum.MessageEventTypeUserCancelThumbUp:
		var message event.EventUserCancelThumbUp
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}
		return c.handleUserCancelThumbUp(ctx, &message)
	default:
		c.logger.Warn("unknown message event type", zap.String("event_type", msg.EventType))
	}
	return nil
}
