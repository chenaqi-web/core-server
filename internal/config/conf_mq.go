package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Topic 解析后的 topic 配置
type Topic struct {
	GroupID              string
	Name                 string
	PartitionNum         int
	ReplicationFactorNum int
}

type KafkaConfig struct {
	Version                   string `yaml:"Version"`
	Host                      string `yaml:"Host"`
	MaxRetry                  int    `yaml:"MaxRetry"`
	Timeout                   int    `yaml:"Timeout"`
	ConsumerMaxProcessingTime int    `yaml:"ConsumerMaxProcessingTime"`
	DlqTopic                  string `yaml:"DlqTopic"`
	LikeTopic                 string `yaml:"LikeTopic"`
}

// Brokers 可能是kafka集群
func (c *KafkaConfig) Brokers() []string {
	return strings.Split(c.Host, ",")
}

// ParseTopic 解析单条 topic 配置
// 格式: group_id;topic_name;partition_num;replication_factor_num
func ParseTopic(raw string) (*Topic, error) {
	parts := strings.Split(raw, ";")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid topic %q, expected group_id;topic_name;partition_num;replication_factor_num", raw)
	}

	partitionNum, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil || partitionNum <= 0 {
		return nil, fmt.Errorf("invalid partition_num in topic %q: %w", raw, err)
	}

	replicationNum, err := strconv.Atoi(strings.TrimSpace(parts[3]))
	if err != nil || replicationNum <= 0 {
		return nil, fmt.Errorf("invalid replication_factor_num in topic %q: %w", raw, err)
	}

	return &Topic{
		GroupID:              strings.TrimSpace(parts[0]),
		Name:                 strings.TrimSpace(parts[1]),
		PartitionNum:         partitionNum,
		ReplicationFactorNum: replicationNum,
	}, nil
}

// ParseTopics 解析所有 topic 配置，返回业务 topic 列表和 dlq topic
func (c *KafkaConfig) ParseTopics() ([]*Topic, *Topic, error) {
	consumerEntries := []string{
		c.LikeTopic,
	}

	// 解析业务Topic
	topics := make([]*Topic, 0, len(consumerEntries))
	for _, raw := range consumerEntries {
		topic, err := ParseTopic(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("parse topic %q: %w", raw, err)
		}
		topics = append(topics, topic)
	}

	// 解析DLQ Topic
	var dlq *Topic
	if raw := strings.TrimSpace(c.DlqTopic); raw != "" {
		parsed, err := ParseTopic(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("DlqTopic: %w", err)
		}
		dlq = parsed
	}

	return topics, dlq, nil
}

func (c *KafkaConfig) DlqTopicName() string {
	_, dlq, err := c.ParseTopics()
	if err != nil || dlq == nil {
		return ""
	}
	return dlq.Name
}

func (c *KafkaConfig) LikeTopicName() (string, error) {
	topic, err := ParseTopic(c.LikeTopic)
	if err != nil {
		return "", err
	}
	return topic.Name, nil
}
