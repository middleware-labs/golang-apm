package logger

import (
	"github.com/fluent/fluent-logger-golang/fluent"
	"os"
)

var (
	host      = getEnv("MW_AGENT_SERVICE", "localhost")
	logger, _ = fluent.New(fluent.Config{FluentPort: 8006, FluentHost: host})
	tag       = "go.app"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

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
