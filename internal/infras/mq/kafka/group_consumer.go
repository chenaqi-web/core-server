package kafka

import (
	"context"
	"fmt"
	"time"

	"backend/core-server/internal/config"

	"github.com/IBM/sarama"
)

// BatchMessagesHandler 返回与 msgs 等长的 bool 数组，true 表示可以提交 offset
type BatchMessagesHandler func(ctx context.Context, msgs []*sarama.ConsumerMessage) ([]bool, error)

type BatchConsumer struct {
	ready   chan bool
	handler BatchMessagesHandler
}

func (consumer *BatchConsumer) Setup(_ sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *BatchConsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *BatchConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := make([]*sarama.ConsumerMessage, 0)
	const batchMaxSize = 100
	flushInterval := 10 * time.Second

	flushTicker := time.NewTicker(flushInterval)
	defer flushTicker.Stop()

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			msgs = append(msgs, msg)
			if len(msgs) >= batchMaxSize {
				canCommit, err := consumer.handler(session.Context(), msgs)
				if err == nil {
					consumer.commitOffset(session, msgs, canCommit)
				}
				msgs = nil
			}

		case <-flushTicker.C:
			if len(msgs) > 0 {
				canCommit, err := consumer.handler(session.Context(), msgs)
				if err == nil {
					consumer.commitOffset(session, msgs, canCommit)
				}
				msgs = nil
			}

		case <-session.Context().Done():
			return session.Context().Err()
		}
	}
}

func (consumer *BatchConsumer) commitOffset(session sarama.ConsumerGroupSession, msgs []*sarama.ConsumerMessage, canCommits []bool) {
	if len(msgs) != len(canCommits) {
		return
	}

	for idx, msg := range msgs {
		if !canCommits[idx] {
			break
		}
		session.MarkMessage(msg, "")
	}
	session.Commit()
}

//======================================================================================================================

type GroupConsumer struct {
	topic    string
	groupID  string
	consumer sarama.ConsumerGroup
}

func NewGroupConsumer(cfg *config.Config, groupID, topic string) (*GroupConsumer, error) {
	saramaConfig := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion(cfg.Kafka.Version)
	if err != nil {
		return nil, fmt.Errorf("parse kafka version: %w", err)
	}

	// version must set before producer/consumer init
	saramaConfig.Version = version
	// 当消费者组内成员变化时，Kafka 会触发 Rebalance（重平衡），这会造成短暂的消费停滞, 这里使用sticky策略替代
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategySticky(),
	}
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	// disable auto commit
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = false
	saramaConfig.Consumer.MaxProcessingTime = time.Duration(cfg.Kafka.ConsumerMaxProcessingTime) * time.Second

	client, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers(), groupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("create consumer group client: %w", err)
	}

	return &GroupConsumer{
		topic:    topic,
		groupID:  groupID,
		consumer: client,
	}, nil
}

func (groupConsumer *GroupConsumer) StartBatchConsume(ctx context.Context, handler BatchMessagesHandler) error {
	consumer := BatchConsumer{
		ready:   make(chan bool),
		handler: handler,
	}

	go func() {
		for {
			_ = groupConsumer.consumer.Consume(ctx, []string{groupConsumer.topic}, &consumer)

			if ctx.Err() != nil {
				return
			}

			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	return nil
}

func (groupConsumer *GroupConsumer) Close() error {
	return groupConsumer.consumer.Close()
}
