package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"backend/core-server/internal/infras/cache"

	"github.com/IBM/sarama"
	"github.com/avast/retry-go"
	"github.com/cespare/xxhash/v2"
)

type MessageHandler func(ctx context.Context, msg *sarama.ConsumerMessage) error

// BatchHandlerWrapper 将单条消息处理器包装为批量处理器。
// redisClient 不为 nil 时，通过消息 hash 做短时去重。
// dlqTopic 不为空且 producer 不为 nil 时，处理失败的消息会写入死信队列。
func BatchHandlerWrapper(
	redisClient *cache.CacheClient,
	producer *SyncProducer,
	dlqTopic string,
	handler MessageHandler,
) BatchMessagesHandler {
	logger := slog.Default().With("component", "kafka_batch_handler")

	return func(ctx context.Context, msgs []*sarama.ConsumerMessage) ([]bool, error) {
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

			hashKey := fmt.Sprintf("kafka:msg:%s", strconv.FormatUint(hashMessage(msg), 10))

			if redisClient != nil {
				success, err := redisClient.Cache.SetNX(ctx, hashKey, "1", 5*time.Minute).Result()
				if err != nil {
					logger.Warn("redis SetNX failed, proceeding without dedup", "error", err, "key", hashKey)
				} else if !success {
					continue
				}
			}

			if err := handleMessageWithDLQ(ctx, producer, dlqTopic, msg, handler); err != nil {
				if redisClient != nil {
					if _, delErr := redisClient.Cache.Del(ctx, hashKey).Result(); delErr != nil {
						logger.Error("failed to delete dedup key", "key", hashKey, "error", delErr)
					}
				}

				canCommit[idx] = false
				if redisClient != nil {
					continue
				}
				return canCommit, nil
			}
		}

		return canCommit, nil
	}
}

func handleMessageWithDLQ(
	ctx context.Context,
	producer *SyncProducer,
	dlqTopic string,
	msg *sarama.ConsumerMessage,
	handler MessageHandler,
) (err error) {
	defer func() {
		if err == nil || producer == nil || dlqTopic == "" {
			return
		}

		err = retry.Do(
			func() error {
				return producer.SendMessage(dlqTopic, string(msg.Key), msg.Value)
			},
			retry.Attempts(3),
			retry.MaxDelay(3*time.Second),
			retry.DelayType(retry.BackOffDelay),
		)
	}()

	err = handler(ctx, msg)
	return err
}

func hashMessage(msg *sarama.ConsumerMessage) uint64 {
	hash := xxhash.New()
	hash.Write(msg.Key)
	hash.Write(msg.Value)
	return hash.Sum64()
}
