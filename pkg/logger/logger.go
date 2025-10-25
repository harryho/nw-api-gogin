package logger

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey struct{}

func New(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Level = parseLevel(level)
	return cfg.Build()
}

func parseLevel(level string) zap.AtomicLevel {
	atomic := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		atomic = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info", "":
		atomic = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn", "warning":
		atomic = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		atomic = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "dpanic":
		atomic = zap.NewAtomicLevelAt(zapcore.DPanicLevel)
	case "panic":
		atomic = zap.NewAtomicLevelAt(zapcore.PanicLevel)
	case "fatal":
		atomic = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	}
	return atomic
}

func WithContext(ctx context.Context, log *zap.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, contextKey{}, log)
}

func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return zap.NewNop()
	}
	if log, ok := ctx.Value(contextKey{}).(*zap.Logger); ok && log != nil {
		return log
	}
	return zap.NewNop()
}

func Sync(log *zap.Logger) {
	if log == nil {
		return
	}
	_ = log.Sync()
}
