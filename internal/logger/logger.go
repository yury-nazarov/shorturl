package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

func New() *logrus.Logger {
	return &logrus.Logger{
		Formatter: &logrus.JSONFormatter{},
		Out: os.Stdout,
		//ReportCaller: true,
		Level: logrus.InfoLevel,
	}
}