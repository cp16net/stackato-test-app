package common

import (
	"os"

	"github.com/Sirupsen/logrus"
)

const (
	// DEBUG level messages will be outputed to log
	DEBUG = "debug"

	// INFO level messages will be outputed to log
	INFO = "info"

	// WARN level messages and above output to log
	WARN = "warn"

	// ERROR level messages will be outputed to log
	ERROR = "error"

	// FATAL level messages will be outputed to log
	FATAL = "fatal"
)

// NewLogger creates a new logger
func NewLogger(level string) *logrus.Logger {

	minLogLevel := logrus.DebugLevel
	switch level {
	case INFO:
		minLogLevel = logrus.InfoLevel
	case WARN:
		minLogLevel = logrus.WarnLevel
	case ERROR:
		minLogLevel = logrus.ErrorLevel
	case FATAL:
		minLogLevel = logrus.FatalLevel
	case DEBUG:
		minLogLevel = logrus.DebugLevel
	}

	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Level = minLogLevel

	return logger
}
