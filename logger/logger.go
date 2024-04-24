package logger

import (
	"os"

	"github.com/fluent/fluent-logger-golang/fluent"
)

var (
	host        = "localhost"
	logger, _   = fluent.New(fluent.Config{FluentPort: 8006, FluentHost: host})
	serviceName = "default-service"
	accessToken = ""
	serverless  = "0"
)

func InitLogger(ServiceName string, AccessToken string, fluentHost string, isServerless string) {
	serviceName = ServiceName
	accessToken = AccessToken
	serverless = isServerless
	host = fluentHost
	target := getEnv("MW_AGENT_SERVICE", host)
	logger, _ = fluent.New(fluent.Config{FluentPort: 8006, FluentHost: target})
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func Post(message string, level string) {
	var data = map[string]string{
		"level":          level,
		"message":        message,
		"service.name":   serviceName,
		"mw.account_key": accessToken,
		"mw_serverless":  serverless,
	}
	go logger.Post(serviceName, data)
}

func Error(message string) {
	go Post(message, "error")
}

func Info(message string) {
	go Post(message, "info")
}

func Warn(message string) {
	go Post(message, "warn")
}

func Debug(message string) {
	go Post(message, "debug")
}
