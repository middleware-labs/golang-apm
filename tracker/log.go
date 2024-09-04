package tracker

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/propagation"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

type Logs struct{}

var LogProvider otellog.LoggerProvider

func (t *Logs) initLogs(ctx context.Context, c *Config) error {
	host := GetHostLog(c.Host)
	exp, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(fmt.Sprint("http://"+host+"/v1/logs")),
	)
	if err != nil {
		log.Println("failed to create exporter for logs: ", err)
	}

	var file *os.File = os.Stdout
	var consoleExporter *stdoutlog.Exporter
	if c.debug {
		if c.debugLogFile {
			file, err = os.OpenFile("./mw-logs.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
			if err != nil {
				log.Println("failed to create exporter file for logs: ", err)
			}
		}
		consoleExporter, err = stdoutlog.New(stdoutlog.WithPrettyPrint(), stdoutlog.WithWriter(file))
		if err != nil {
			log.Println("failed to create debug console exporter for logs: ", err)
		}
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", c.ServiceName),
			attribute.String("library.language", "go"),
			attribute.Bool("mw_agent", true),
			attribute.String("project.name", c.projectName),
			attribute.String("mw.app.lang", "go"),
			attribute.String("mw.account_key", c.AccessToken),
			attribute.String("mw_serverless", c.isServerless),
		),
	)

	if err != nil {
		log.Println("failed to set resources for logs:", err)
	}

	if c.debug {
		LogProvider = *otellog.NewLoggerProvider(
			otellog.WithResource(resources),
			otellog.WithProcessor(otellog.NewBatchProcessor(consoleExporter)),
			otellog.WithProcessor(otellog.NewBatchProcessor(exp)),
		)
	} else {
		LogProvider = *otellog.NewLoggerProvider(
			otellog.WithResource(resources),
			otellog.WithProcessor(otellog.NewBatchProcessor(exp)),
		)
	}

	c.Lp = &LogProvider

	p := b3.New()
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			p,
			propagation.TraceContext{},
			propagation.Baggage{}),
	)

	return err
}

func GetHostLog(s string) string {
	suffix := ":9319"
	s = s[:len(s)-len(suffix)] + ":9320"
	return s
}
