package kafka

import (
	"fmt"
	"log/slog"
	"time"

	"backend/core-server/internal/config"

	"github.com/IBM/sarama"
)

type SyncProducer struct {
	logger   *slog.Logger
	producer sarama.SyncProducer
}

func NewSyncProducer(cfg *config.Config) (*SyncProducer, error) {
	saramaConfig := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion(cfg.Kafka.Version)
	if err != nil {
		return nil, fmt.Errorf("parse kafka version: %w", err)
	}

	saramaConfig.Version = version
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = cfg.Kafka.MaxRetry
	saramaConfig.Producer.Timeout = time.Duration(cfg.Kafka.Timeout) * time.Second
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers(), saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("start sync producer: %w", err)
	}

	return &SyncProducer{
		logger:   slog.Default().With("component", "kafka_sync_producer"),
		producer: producer,
	}, nil
}

func (s *SyncProducer) Close() error {
	return s.producer.Close()
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

func (s *SyncProducer) SendMessage(topic, key string, value []byte) error {
	partition, offset, err := s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	})
	if err != nil {
		return err
	}

	s.logger.Debug("message sent", "topic", topic, "partition", partition, "offset", offset)
	return nil
}
