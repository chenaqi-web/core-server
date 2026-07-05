package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"backend/core-server/internal/config"

	"golang.org/x/sync/errgroup"
)

type KafkaManager struct {
	// config
	cfg *config.Config

	// Topic 管理器
	topicManager *TopicManager

	mu sync.Mutex

	// 消费者组map
	groupConsumersMap map[string]*GroupConsumer

	// lifecycle controller
	ctx    context.Context
	cancel context.CancelFunc

	batchHandler BatchMessagesHandler
}

func NewKafkaManager(cfg *config.Config, topicManager *TopicManager) *KafkaManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaManager{
		cfg:               cfg,
		topicManager:      topicManager,
		groupConsumersMap: make(map[string]*GroupConsumer),
		ctx:               ctx,
		cancel:            cancel,
	}
}

func (km *KafkaManager) Close() error {
	if err := km.StopGroupConsumers(); err != nil {
		return err
	}
	return km.topicManager.Close()
}

func (km *KafkaManager) SetBatchHandler(handler BatchMessagesHandler) {
	km.batchHandler = handler
}

func (km *KafkaManager) InitTopics() error {
	if err := km.topicManager.CreateTopics(); err != nil {
		return err
	}
	return km.topicManager.CreateDlqTopic()
}

func (km *KafkaManager) StartGroupConsumers() error {
	if km.batchHandler == nil {
		return errors.New("batch handler is not set")
	}

	km.mu.Lock()
	if len(km.groupConsumersMap) > 0 {
		km.mu.Unlock()
		return errors.New("consumers already started; call StopGroupConsumers first")
	}
	km.groupConsumersMap = make(map[string]*GroupConsumer)
	km.mu.Unlock()

	g, _ := errgroup.WithContext(km.ctx)
	g.SetLimit(10)

	for _, topic := range km.topicManager.topics {
		topic := topic
		g.Go(func() error {
			consumer, err := NewGroupConsumer(km.cfg, topic.GroupID, topic.Name)
			if err != nil {
				return err
			}

			km.mu.Lock()
			km.groupConsumersMap[topic.Name] = consumer
			km.mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("create group consumers: %w", err)
	}

	var startErr error
	for _, consumer := range km.groupConsumersMap {
		if err := consumer.StartBatchConsume(km.ctx, km.batchHandler); err != nil {
			startErr = errors.Join(startErr, err)
		}
	}

	return startErr
}

func (km *KafkaManager) StopGroupConsumers() error {
	km.cancel()

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(10)

	for topicName, consumer := range km.groupConsumersMap {
		g.Go(func() error {
			if err := consumer.Close(); err != nil {
				return fmt.Errorf("stop consumer %s: %w", topicName, err)
			}
			return nil
		})
	}

	err := g.Wait()

	km.mu.Lock()
	km.groupConsumersMap = make(map[string]*GroupConsumer)
	km.mu.Unlock()

	km.ctx, km.cancel = context.WithCancel(context.Background())
	return err
}
