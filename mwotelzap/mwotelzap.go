package mwotelzap

import (
	"context"
	"os"

	"github.com/agoda-com/opentelemetry-go/otelzap"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogsgrpc"
	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"github.com/middleware-labs/golang-apm/tracker"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"go.uber.org/zap/zapcore"
)

// configure common attributes for all logs
func newResource(config *tracker.Config) *resource.Resource {
	hostName, _ := os.Hostname()
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
		semconv.HostName(hostName),
	)
}

func NewMWOtelCore(config *tracker.Config) zapcore.Core {
	ctx := context.Background()
	exporter, _ := otlplogs.NewExporter(ctx, otlplogs.WithClient(otlplogsgrpc.NewClient(otlplogsgrpc.WithEndpoint("localhost:9319"))))
	loggerProvider := sdk.NewLoggerProvider(
		sdk.WithBatcher(exporter),
		sdk.WithResource(newResource(config)),
	)

	return otelzap.NewOtelCore(loggerProvider)
}
