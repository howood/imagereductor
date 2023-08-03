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

var log *zap.Logger

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

// Debug log output with DEBUG
func Debug(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Debug(fmt.Sprintf("%v", msg[0]), zap.String("file", file), zap.String("PackegeName", packegeName), zap.Any(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey())), zap.Any("messages", msg))
}

// Info log output with Info
func Info(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Info(fmt.Sprintf("%v", msg[0]), zap.String("file", file), zap.String("PackegeName", packegeName), zap.Any(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey())), zap.Any("messages", msg))
}

// Warn log output with Warn
func Warn(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Warn(fmt.Sprintf("%v", msg[0]), zap.String("file", file), zap.String("PackegeName", packegeName), zap.Any(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey())), zap.Any("messages", msg))
}

// Error log output with Error
func Error(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Error(fmt.Sprintf("%v", msg[0]), zap.String("file", file), zap.String("PackegeName", packegeName), zap.Any(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey())), zap.Any("messages", msg))
}

// Panic log output with Panic
func Panic(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Panic(fmt.Sprintf("%v", msg[0]), zap.String("file", file), zap.String("PackegeName", packegeName), zap.Any(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey())), zap.Any("messages", msg))
}

// Fatal log output with Fatal
func Fatal(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	file := filename + ":" + strconv.Itoa(line)
	log.Fatal(fmt.Sprintf("%v", msg[0]), zap.String("file", file), zap.String("PackegeName", packegeName), zap.Any(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey())), zap.Any("messages", msg))
}
