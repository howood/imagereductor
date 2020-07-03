package logger

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

const PackegeName = "imagereductor"

const (
	LogModeFew    = "few"
	LogModeMedium = "minimum"
)

var log *logrus.Entry

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

func Debug(msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log.Debug("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

func Info(msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log.Info("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

func Warn(msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log.Warn("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

func Error(msg ...interface{}) {
	_, filename, line, _ := runtime.Caller(1)
	log.Error("["+filename+":"+strconv.Itoa(line)+"] ", msg)
}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := fmt.Sprintf(entry.Time.Format(f.TimestampFormat))
	return []byte(fmt.Sprintf("[%s] [%s] [%s] %s \n", timestamp, f.LevelDesc[entry.Level], PackegeName, entry.Message)), nil
}
