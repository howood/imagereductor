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

const packegeName = "imagereductor"

const (
	logModeFew    = "few"
	logModeMedium = "minimum"
)

//nolint:gochecknoglobals
var log *zap.Logger

//nolint:gochecknoinits
func init() {
	level := zap.DebugLevel
	if os.Getenv("VERIFY_MODE") != "enable" {
		switch os.Getenv("LOG_MODE") {
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
func Debug(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Debug(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

// Info log output with Info.
func Info(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Info(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

// Warn log output with Warn.
func Warn(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Warn(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

// Error log output with Error.
func Error(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Error(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

// Panic log output with Panic.
func Panic(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Panic(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

// Fatal log output with Fatal.
func Fatal(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Fatal(fmt.Sprintf("%v", msg[0]), metadataFields(ctx, file, msg)...)
}

func metadataFields(ctx context.Context, file string, msgs []interface{}) []zap.Field {
	messages := make([]interface{}, 0)
	if len(msgs) > 1 {
		messages = msgs[1:]
	}
	return []zap.Field{
		zap.String("PackegeName", packegeName),
		zap.String("file", file),
		zap.Any(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey())),
		zap.Any("messages", messages),
	}
}
