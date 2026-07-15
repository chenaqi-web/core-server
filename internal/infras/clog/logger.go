package clog

import (
	"fmt"
	"strings"

	"backend/core-server/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ModeConsole = "console"
	ModeFile    = "file"
)

// Log 可注入的结构化日志对象
type Log struct {
	zl *zap.Logger
}

// NewLog 根据配置创建 Log，供依赖注入使用。
// Mode=console: 终端彩色结构化输出
// Mode=file:    写入文件（JSON）
func NewLog(cfg *config.Config) (*Log, error) {
	level, err := parseLevel(cfg.Log.Level)
	if err != nil {
		return nil, err
	}

	mode := strings.ToLower(strings.TrimSpace(cfg.Log.Mode))
	if mode == "" {
		mode = ModeConsole
	}

	var core zapcore.Core
	switch mode {
	case ModeConsole:
		core = newConsoleCore(level)
	case ModeFile:
		fileCore, err := newFileCore(level, cfg.Log.Filename)
		if err != nil {
			return nil, err
		}
		core = fileCore
	default:
		return nil, fmt.Errorf("unsupported log mode %q, want console|file", cfg.Log.Mode)
	}

	zl := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &Log{zl: zl}, nil
}

func (l *Log) Zap() *zap.Logger {
	if l == nil || l.zl == nil {
		return zap.NewNop()
	}
	return l.zl
}

func (l *Log) Sugar() *zap.SugaredLogger {
	return l.Zap().Sugar()
}

// With 绑定字段，返回子 Log（便于按 component 注入）
func (l *Log) With(fields ...zap.Field) *Log {
	return &Log{zl: l.Zap().With(fields...)}
}

// Sync 刷盘，进程退出前调用
func (l *Log) Sync() error {
	if l == nil || l.zl == nil {
		return nil
	}
	return l.zl.Sync()
}

func (l *Log) Debug(msg string, fields ...zap.Field) {
	l.Zap().WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

func (l *Log) Info(msg string, fields ...zap.Field) {
	l.Zap().WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

func (l *Log) Warn(msg string, fields ...zap.Field) {
	l.Zap().WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

func (l *Log) Error(msg string, fields ...zap.Field) {
	l.Zap().WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

func (l *Log) Fatal(msg string, fields ...zap.Field) {
	l.Zap().WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}

func parseLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "info":
		return zapcore.InfoLevel, nil
	case "debug":
		return zapcore.DebugLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		var l zapcore.Level
		if err := l.UnmarshalText([]byte(level)); err != nil {
			return zapcore.InfoLevel, fmt.Errorf("invalid log level %q: %w", level, err)
		}
		return l, nil
	}
}
