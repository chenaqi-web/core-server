package repo

import (
	"core-server/internal/config"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDBClient(t *testing.T) {
	require.NoError(t, os.Chdir("../../.."))

	cfg, err := config.Load()
	require.NoError(t, err)

	_, err = NewDBClient(cfg)
	require.NoError(t, err)
}
