package kafka_test

import (
	"testing"

	"backend/core-server/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTopic(t *testing.T) {
	topic, err := config.ParseTopic("like_group;like_topic;3;1")
	require.NoError(t, err)
	assert.Equal(t, "like_group", topic.GroupID)
	assert.Equal(t, "like_topic", topic.Name)
	assert.Equal(t, 3, topic.PartitionNum)
	assert.Equal(t, 1, topic.ReplicationFactorNum)
}

func TestKafkaConfigParseTopics(t *testing.T) {
	cfg := config.KafkaConfig{
		LikeTopic: "like_group;like_topic;3;1",
		DlqTopic:  "dlq_group;dlq_topic;1;1",
	}

	topics, dlq, err := cfg.ParseTopics()
	require.NoError(t, err)
	require.Len(t, topics, 1)
	assert.Equal(t, "like_topic", topics[0].Name)
	require.NotNil(t, dlq)
	assert.Equal(t, "dlq_topic", dlq.Name)
	assert.Equal(t, "dlq_topic", cfg.DlqTopicName())
}
