package clog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newFileCore(level zapcore.Level, filename string) (zapcore.Core, error) {
	if strings.TrimSpace(filename) == "" {
		return nil, fmt.Errorf("log filename is required when mode is file")
	}

	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file %s: %w", filename, err)
	}

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		zapcore.AddSync(f),
		level,
	), nil
}
