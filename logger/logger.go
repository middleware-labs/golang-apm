package logger

import (
	"github.com/fluent/fluent-logger-golang/fluent"
	"os"
)

var (
	host        = getEnv("MW_AGENT_SERVICE", "localhost")
	logger, _   = fluent.New(fluent.Config{FluentPort: 8006, FluentHost: host})
	tag         = "go.app"
	projectName = ""
	serviceName = ""
)

func InitLogger(project string, service string) {
	projectName = project
	serviceName = service
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func Error(message string) {
	var data = map[string]string{
		"level":        "error",
		"message":      message,
		"project.name": projectName,
		"service.name": serviceName,
	}
	logger.Post(tag, data)
}

func Info(message string) {
	var data = map[string]string{
		"level":        "info",
		"message":      message,
		"project.name": projectName,
		"service.name": serviceName,
	}
	logger.Post(tag, data)
}

func Warn(message string) {
	var data = map[string]string{
		"level":        "warn",
		"message":      message,
		"project.name": projectName,
		"service.name": serviceName,
	}
	logger.Post(tag, data)
}

func Debug(message string) {
	var data = map[string]string{
		"level":        "debug",
		"message":      message,
		"project.name": projectName,
		"service.name": serviceName,
	}
	logger.Post(tag, data)
}
