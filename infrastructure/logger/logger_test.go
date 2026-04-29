package logger_test

import (
	"context"
	"testing"

	logger "github.com/howood/imagereductor/infrastructure/logger"
)

// These tests primarily exercise the log functions for coverage.
// We avoid Fatal which would terminate the test process.
func Test_Logger_AllLevels(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger.Debug(ctx, "debug message")
	logger.Info(ctx, "info message", "extra1", "extra2")
	logger.Warn(ctx, "warn message")
	logger.Error(ctx, "error message")
}

func Test_Logger_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	logger.Panic(context.Background(), "panic test message")
}
