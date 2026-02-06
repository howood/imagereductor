package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/howood/imagereductor/infrastructure/requestid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const packageName = "imagereductor"

const (
	logModeFew    = "few"
	logModeMedium = "minimum"
)

// Supported log levels for LOG_LEVEL environment variable.
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

//nolint:gochecknoglobals
var log *zap.Logger

//nolint:gochecknoinits,cyclop
func init() {
	level := zap.InfoLevel // default to Info level

	// Priority 1: LOG_LEVEL environment variable (recommended)
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		switch logLevel {
		case LogLevelDebug:
			level = zap.DebugLevel
		case LogLevelInfo:
			level = zap.InfoLevel
		case LogLevelWarn:
			level = zap.WarnLevel
		case LogLevelError:
			level = zap.ErrorLevel
		case LogLevelFatal:
			level = zap.FatalLevel
		default:
			// Invalid LOG_LEVEL, use default Info
			fmt.Fprintf(os.Stderr, "Invalid LOG_LEVEL '%s', using 'info'. Valid values: debug, info, warn, error, fatal\n", logLevel)
		}
	} else if os.Getenv("VERIFY_MODE") == "enable" {
		// Priority 2: VERIFY_MODE=enable forces Debug level
		level = zap.DebugLevel
	} else if logMode := os.Getenv("LOG_MODE"); logMode != "" {
		// Priority 3: Legacy LOG_MODE (deprecated, kept for backward compatibility)
		switch logMode {
		case logModeFew:
			level = zap.WarnLevel
		case logModeMedium:
			level = zap.ErrorLevel
		default:
			level = zap.InfoLevel
		}
	}
	conf := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "name",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	var err error
	log, err = conf.Build()
	if err != nil {
		panic(err)
	}
}

// Debug log output with Debug.
func Debug(ctx context.Context, msg ...any) {
	_, filename, line, _ := runtime.Caller(2)
	file := filename + ":" + strconv.Itoa(line)
	log.Debug(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

// Info log output with Info.
func Info(ctx context.Context, msg ...any) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Info(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

// Warn log output with Warn.
func Warn(ctx context.Context, msg ...any) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Warn(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

// Error log output with Error.
func Error(ctx context.Context, msg ...any) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Error(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

// Panic log output with Panic.
func Panic(ctx context.Context, msg ...any) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Panic(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

// Fatal log output with Fatal.
func Fatal(ctx context.Context, msg ...any) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Fatal(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

func metadataFields(ctx context.Context, file string, msgs []any) []zap.Field {
	messages := make([]any, 0)
	if len(msgs) > 1 {
		messages = msgs[1:]
	}
	return []zap.Field{
		zap.String("PackageName", packageName),
		zap.String("file", file),
		zap.Any(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey())),
		zap.Any("messages", messages),
	}
}
