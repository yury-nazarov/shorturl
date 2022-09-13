package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func New() *logrus.Logger {
	return &logrus.Logger{
		Formatter: &logrus.JSONFormatter{},
		Out: os.Stdout,
		//ReportCaller: true,
		Level: logrus.InfoLevel,
	}
}