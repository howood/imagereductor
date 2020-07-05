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

const PackegeName = "imagereductor"

const (
	LogModeFew    = "few"
	LogModeMedium = "minimum"
)

var log *logrus.Entry

type LogEntry logrus.Entry
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
		case LogModeFew:
			logrus.SetLevel(logrus.WarnLevel)
		case LogModeMedium:
			logrus.SetLevel(logrus.ErrorLevel)
		default:
			logrus.SetLevel(logrus.InfoLevel)
		}
	}

	log = logrus.WithFields(logrus.Fields{})
}

func GetLogger(xReqID string) *logrus.Entry {
	return logrus.WithField(requestid.KeyRequestID, xReqID)
}

func Debug(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log = logrus.WithField(requestid.KeyRequestID, ctx.Value(requestid.KeyRequestID))
	log.Debug("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

func Info(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log = logrus.WithField(requestid.KeyRequestID, ctx.Value(requestid.KeyRequestID))
	log.Info("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

func Warn(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log = logrus.WithField(requestid.KeyRequestID, ctx.Value(requestid.KeyRequestID))
	log.Warn("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

func Error(ctx context.Context, msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log = logrus.WithField(requestid.KeyRequestID, ctx.Value(requestid.KeyRequestID))
	log.Error("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := fmt.Sprintf(entry.Time.Format(f.TimestampFormat))
	return []byte(fmt.Sprintf("[%s] [%s] [%s] [%s] %s \n", timestamp, f.LevelDesc[entry.Level], PackegeName, entry.Data[requestid.KeyRequestID], entry.Message)), nil
}
