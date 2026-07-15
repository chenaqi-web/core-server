package kafka

import (
	"backend/core-server/internal/config"
	"fmt"

	"github.com/IBM/sarama"
)

type TopicManager struct {

	// 用于管理Topic
	clusterAdmin sarama.ClusterAdmin
	topics       []*config.Topic
	dlqTopic     *config.Topic
}

func NewTopicManager(cfg *config.Config) (*TopicManager, error) {
	// 1. 拿到基础配置
	saramaConfig := sarama.NewConfig()
	version, err := sarama.ParseKafkaVersion(cfg.Kafka.Version)
	if err != nil {
		return nil, fmt.Errorf("parse kafka version: %w", err)
	}
	saramaConfig.Version = version

	// cluster admin
	clusterAdmin, err := sarama.NewClusterAdmin(cfg.Kafka.Brokers(), saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("create cluster admin: %w", err)
	}

	// 解析对应 topics
	topics, dlqTopic, err := cfg.Kafka.ParseTopics()
	if err != nil {
		_ = clusterAdmin.Close()
		return nil, err
	}

	return &TopicManager{
		clusterAdmin: clusterAdmin,
		topics:       topics,
		dlqTopic:     dlqTopic,
	}, nil
}

func (tm *TopicManager) CreateTopics() error {
	existingTopics, err := tm.clusterAdmin.ListTopics()
	if err != nil {
		return fmt.Errorf("list topics: %w", err)
	}

	for _, topic := range tm.topics {
		if _, exists := existingTopics[topic.Name]; exists {
			continue
		}

		detail := &sarama.TopicDetail{
			NumPartitions:     int32(topic.PartitionNum),
			ReplicationFactor: int16(topic.ReplicationFactorNum),
		}

		if err = tm.clusterAdmin.CreateTopic(topic.Name, detail, false); err != nil {
			return fmt.Errorf("create topic %s: %w", topic.Name, err)
		}

	}

	return nil
}

func (tm *TopicManager) CreateDlqTopic() error {
	if tm.dlqTopic == nil {
		return nil
	}

	existingTopics, err := tm.clusterAdmin.ListTopics()
	if err != nil {
		return fmt.Errorf("list topics: %w", err)
	}

	if _, ok := existingTopics[tm.dlqTopic.Name]; ok {
		return nil
	}

	detail := &sarama.TopicDetail{
		NumPartitions:     int32(tm.dlqTopic.PartitionNum),
		ReplicationFactor: int16(tm.dlqTopic.ReplicationFactorNum),
	}

	if err = tm.clusterAdmin.CreateTopic(tm.dlqTopic.Name, detail, false); err != nil {
		return fmt.Errorf("create dlq topic %s: %w", tm.dlqTopic.Name, err)
	}

	return nil
}

func (tm *TopicManager) Close() error {
	return tm.clusterAdmin.Close()
}
