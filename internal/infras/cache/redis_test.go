package cache

import (
	"backend/core-server/internal/config"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedisClient(t *testing.T) {
	require.NoError(t, os.Chdir("../../.."))
	cfg, err := config.Load()
	require.NoError(t, err)
	NewClient(cfg)
}
