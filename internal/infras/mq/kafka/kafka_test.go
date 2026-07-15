package kafka

import (
	"backend/core-server/internal/config"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewKafkaManager(t *testing.T) {
	require.NoError(t, os.Chdir("../../../.."))

	cfg, err := config.Load()
	require.NoError(t, err)

	topicManager, err := NewTopicManager(cfg)
	require.NoError(t, err)

	km := NewKafkaManager(cfg, topicManager)

	require.NoError(t, km.Close())
}
