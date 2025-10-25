package logger

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewRespectsLogLevel(t *testing.T) {
	log, err := New("error")
	if err != nil {
		t.Fatalf("unexpected error creating logger: %v", err)
	}
	defer Sync(log)

	if log.Check(zapcore.DebugLevel, "debug") != nil {
		t.Fatalf("expected debug level to be disabled for error logger")
	}
	if log.Check(zapcore.ErrorLevel, "error") == nil {
		t.Fatalf("expected error level to be enabled")
	}
}

func TestWithContextAndFromContext(t *testing.T) {
	base := context.Background()
	log := zap.NewExample()
	defer Sync(log)

	ctx := WithContext(base, log)
	retrieved := FromContext(ctx)

	if retrieved != log {
		t.Fatalf("expected to retrieve same logger instance from context")
	}

	//nolint:staticcheck // verify nil context falls back to noop logger
	if FromContext(nil) == nil {
		t.Fatalf("expected non-nil logger for nil context")
	}

	if FromContext(context.Background()) == nil {
		t.Fatalf("expected fallback logger for empty context")
	}
}

func TestSyncHandlesNilLogger(t *testing.T) {
	Sync(nil)
}
