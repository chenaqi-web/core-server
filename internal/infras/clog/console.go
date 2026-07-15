package clog

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newConsoleCore(level zapcore.Level) zapcore.Core {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = encodeBracketTime
	cfg.EncodeLevel = encodeBracketColorLevel
	cfg.EncodeCaller = encodeBracketCaller
	cfg.ConsoleSeparator = " "

	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(os.Stdout),
		level,
	)
}

// 终端输出形如: [2026-07-15T15:37:00.000+08:00] [INFO] [logger.go:42] msg key=value
func encodeBracketTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + t.Format("2006-01-02T15:04:05.000Z07:00") + "]")
}

func encodeBracketColorLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var color uint8
	switch level {
	case zapcore.DebugLevel:
		color = 35 // magenta
	case zapcore.InfoLevel:
		color = 34 // blue
	case zapcore.WarnLevel:
		color = 33 // yellow
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		color = 31 // red
	default:
		color = 37 // white
	}
	enc.AppendString(fmt.Sprintf("\x1b[%dm[%s]\x1b[0m", color, strings.ToUpper(level.String())))
}

func encodeBracketCaller(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + caller.TrimmedPath() + "]")
}
