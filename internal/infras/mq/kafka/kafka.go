package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"backend/core-server/internal/config"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
)

// KafkaManager 本质是一个消费协调器
// 创建Topic和消费者

type KafkaManager struct {
	// config
	cfg *config.Config

	// Topic 管理器
	topicManager *TopicManager

	// lifecycle controller
	ctx    context.Context
	cancel context.CancelFunc

	// 消费者组map
	mu                *sync.Mutex
	groupConsumersMap map[string][]*GroupConsumer

	// 用于消费的处理器
	batchHandler BatchMessagesHandler
}

// NewKafkaManager 新建一个消费者管理对象
func NewKafkaManager(cfg *config.Config, topicManager *TopicManager) *KafkaManager {
	ctx, cancel := context.WithCancel(context.Background())

	// 1.初始化Topic
	_ = topicManager.CreateTopics()
	_ = topicManager.CreateDlqTopic()

	return &KafkaManager{
		cfg:               cfg,
		topicManager:      topicManager,
		mu:                &sync.Mutex{},
		groupConsumersMap: make(map[string][]*GroupConsumer),
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

// =====================================================================================================================
// 下面的方法均是在job中使用

// SetBatchHandler 在定时任务那里设置
func (km *KafkaManager) SetBatchHandler(handler BatchMessagesHandler) {
	km.batchHandler = handler
}

func (km *KafkaManager) StartGroupConsumers() error {
	// 首先需要设置处理器handler，如果没有则无法消费
	if km.batchHandler == nil {
		return errors.New("batch handler is not set")
	}

	// 保证启动时只创建一个map
	km.mu.Lock()
	if len(km.groupConsumersMap) > 0 {
		km.mu.Unlock()
		return errors.New("consumers already started; call StopGroupConsumers first")
	}
	km.groupConsumersMap = make(map[string][]*GroupConsumer)
	km.mu.Unlock()

	// 控制并发数 10个
	g, _ := errgroup.WithContext(km.ctx)
	g.SetLimit(10)

	for _, topic := range km.topicManager.topics {
		topicName := topic.Name
		for i := 0; i < topic.PartitionNum; i++ {
			g.Go(func() error {
				consumer, err := NewGroupConsumer(km.cfg, topic.GroupID, topic.Name)
				if err != nil {
					return err
				}

				km.mu.Lock()
				defer km.mu.Unlock()

				if _, ok := km.groupConsumersMap[topicName]; !ok {
					km.groupConsumersMap[topicName] = make([]*GroupConsumer, 0)
				}

				km.groupConsumersMap[topicName] = append(km.groupConsumersMap[topicName], consumer)
				return nil
			})
		}

	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("create group consumers: %w", err)
	}

	// start group consumers
	var multiErr *multierror.Error
	for _, val := range km.groupConsumersMap {
		for _, consumer := range val {
			err := consumer.StartBatchConsume(km.ctx, km.batchHandler)
			if err != nil {
				multiErr = multierror.Append(multiErr, err)
			}
		}
	}

	return multiErr.ErrorOrNil()
}

func (km *KafkaManager) StopGroupConsumers() error {
	km.cancel()

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(10)

	// wait for all group consumers to stop
	for _, val := range km.groupConsumersMap {
		consumers := val

		g.Go(func() error {
			var me *multierror.Error

			for _, consumer := range consumers {
				if err := consumer.Close(); err != nil {
					me = multierror.Append(me, err)
				}
			}

			return me.ErrorOrNil()
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
