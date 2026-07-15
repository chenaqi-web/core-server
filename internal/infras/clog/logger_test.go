package clog

import (
	"os"
	"path/filepath"
	"testing"

	"backend/core-server/internal/config"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewLogConsole(t *testing.T) {
	logger, err := NewLog(&config.Config{
		Log: config.LogConfig{
			Level: "debug",
			Mode:  ModeConsole,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.Info("console log ok", zap.String("mode", ModeConsole), zap.Int("n", 1))
	logger.Debug("debug field", zap.Bool("ok", true))
	logger.Warn("warn field", zap.String("k", "v"))
	logger.Error("error field", zap.String("err", "demo"))
	_ = logger.Sync()
}

func TestNewLogFile(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "app.log")

	logger, err := NewLog(&config.Config{
		Log: config.LogConfig{
			Level:    "info",
			Mode:     ModeFile,
			Filename: filename,
		},
	})
	require.NoError(t, err)

	logger.Info("file log ok", zap.String("mode", ModeFile))
	_ = logger.Sync()

	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Contains(t, string(data), `"msg":"file log ok"`)
	require.Contains(t, string(data), `"mode":"file"`)
}
