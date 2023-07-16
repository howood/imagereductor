package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/howood/imagereductor/infrastructure/requestid"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

const packegeName = "imagereductor"

const (
	logModeFew    = "few"
	logModeMedium = "minimum"
)

var log *logrus.Entry

// PlainFormatter struct
type PlainFormatter struct {
	TimestampFormat string
	LevelDesc       []string
}

func init() {
	plainFormatter := new(PlainFormatter)
	plainFormatter.TimestampFormat = "2006-01-02 15:04:05.999+00:00"
	plainFormatter.LevelDesc = []string{"PANC", "FATAL", "ERROR", "WARN", "INFO", "DEBUG"}
	logrus.SetFormatter(plainFormatter)

	logrus.SetOutput(colorable.NewColorableStdout())

	if os.Getenv("VERIFY_MODE") == "enable" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		switch os.Getenv("LOG_MODE") {
		case logModeFew:
			logrus.SetLevel(logrus.WarnLevel)
		case logModeMedium:
			logrus.SetLevel(logrus.ErrorLevel)
		default:
			logrus.SetLevel(logrus.InfoLevel)
		}
	}

	log = logrus.WithFields(logrus.Fields{})
}

// Debug log output with DEBUG
func Debug(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log = logrus.WithField(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey()))
	log.Debug("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

// Info log output with Info
func Info(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log = logrus.WithField(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey()))
	log.Info("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

// Warn log output with Warn
func Warn(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log = logrus.WithField(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey()))
	log.Warn("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

// Error log output with Error
func Error(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log = logrus.WithField(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey()))
	log.Error("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

// Panic log output with Panic
func Panic(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log = logrus.WithField(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey()))
	log.Panic("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

// Fatal log output with Fatal
func Fatal(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log = logrus.WithField(requestid.KeyRequestID, ctx.Value(requestid.GetRequestIDKey()))
	log.Fatal("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

// Format is formatted log output
func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(f.TimestampFormat)
	return []byte(fmt.Sprintf("[%s] [%s] [%s] [%s] %s \n", timestamp, f.LevelDesc[entry.Level], packegeName, entry.Data[requestid.KeyRequestID], entry.Message)), nil
}
