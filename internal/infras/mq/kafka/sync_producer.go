package kafka

import (
	"backend/core-server/internal/infras/clog"
	"fmt"
	"time"

	"backend/core-server/internal/config"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// 生产者模型搭建 SyncProducer（同步生成者模型）
// 如果需要producer，需要在application中加入SyncProducer实现异步消费
// 生产者需要注意生产者方面的一些知识
//
// 提供2个主要的功能：
// 1. SendMessage  推送单条消息
// 2. SendMessages 批量推送消息

type SyncProducer struct {
	logger   *clog.Log
	producer sarama.SyncProducer
}

func NewSyncProducer(cfg *config.Config, log *clog.Log) (*SyncProducer, error) {
	// 拿到基础配置
	saramaConfig := sarama.NewConfig()

	// 解析Kafka的版本号
	version, err := sarama.ParseKafkaVersion(cfg.Kafka.Version)
	if err != nil {
		return nil, fmt.Errorf("parse kafka version: %w", err)
	}

	// 生产者相关的配置
	saramaConfig.Version = version
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll // 发送应答机制
	saramaConfig.Producer.Retry.Max = cfg.Kafka.MaxRetry
	saramaConfig.Producer.Timeout = time.Duration(cfg.Kafka.Timeout) * time.Second
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true

	// 新建生产者
	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers(), saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("start sync producer: %w", err)
	}

	return &SyncProducer{
		logger:   log,
		producer: producer,
	}, nil
}

func (s *SyncProducer) Close() error {
	return s.producer.Close()
}

func (s *SyncProducer) SendMessage(topic, key string, value []byte) error {
	partition, offset, err := s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key), // 消息key，指定分区（保证同一用户内容发到同一分区保证有序）
		Value: sarama.ByteEncoder(value), // 消息体
	})
	if err != nil {
		return err
	}

	s.logger.Debug("message sent",
		zap.String("topic", topic),
		zap.Int64("offset", offset),
		zap.Int32("partition", partition))
	return nil
}

func (s *SyncProducer) SendMessages(topic string, keyBatch []string, valueBatch [][]byte) error {
	if len(keyBatch) == 0 {
		return nil
	}

	if len(keyBatch) != len(valueBatch) {
		return fmt.Errorf("invalid batch length for key:%d and value:%d", len(keyBatch), len(valueBatch))
	}

	messages := make([]*sarama.ProducerMessage, len(keyBatch))
	for i := range keyBatch {
		messages[i] = &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.ByteEncoder(keyBatch[i]),
			Value: sarama.ByteEncoder(valueBatch[i]),
		}
	}

	return s.producer.SendMessages(messages)
}
