package kafka

import (
	"core-server/internal/config"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewKafkaManager(t *testing.T) {
	require.NoError(t, os.Chdir("../../../.."))

	cfg, err := config.Load()
	require.NoError(t, err)
	cfg.Kafka.Enabled = false

	topicManager, err := NewTopicManager(cfg)
	require.NoError(t, err)
	require.Nil(t, topicManager)

	km := NewKafkaManager(cfg, topicManager)
	require.Nil(t, km)

	require.NoError(t, km.Close())
}
