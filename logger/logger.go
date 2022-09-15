package logger

import (
	"github.com/fluent/fluent-logger-golang/fluent"
)

var (
	logger, _ = fluent.New(fluent.Config{FluentPort: 8006, FluentHost: "localhost"})
	tag       = "go.app"
)

func Error(message string) {
	var data = map[string]string{
		"level":   "error",
		"message": message,
	}
	logger.Post(tag, data)
}

func Info(message string) {
	var data = map[string]string{
		"level":   "info",
		"message": message,
	}
	logger.Post(tag, data)
}

func Warn(message string) {
	var data = map[string]string{
		"level":   "warn",
		"message": message,
	}
	logger.Post(tag, data)
}
