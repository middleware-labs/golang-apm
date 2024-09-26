package tracker

import (
	"net/url"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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

	var host string

	if c.isServerless == "0" {
		host, _ = url.JoinPath("http://"+c.LogHost+":9320", "v1", "logs")


	} else {
		host, _ = url.JoinPath("https://"+c.Host, "v1", "logs")
	}

	exp, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(host),
		// Gzip Compression
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
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

	attributes := []attribute.KeyValue{
		attribute.String("service.name", c.ServiceName),
		attribute.String("service.name", c.ServiceName),
		attribute.String("library.language", "go"),
		attribute.Bool("mw_agent", true),
		attribute.String("project.name", c.projectName),
		attribute.String("mw.app.lang", "go"),
		attribute.String("mw.account_key", c.AccessToken),
		attribute.String("mw_serverless", c.isServerless),
		attribute.String("mw.sdk.version", c.SdkVersion),	
	}

	for key, value := range c.customResourceAttributes {
		switch v := value.(type) {
		case string:
			attributes = append(attributes, attribute.String(key, v))
		case bool:
			attributes = append(attributes, attribute.Bool(key, v))
		case int:
			attributes = append(attributes, attribute.Int(key, v)) // handle int
		case int64:
			attributes = append(attributes, attribute.Int64(key, v)) // handle int64
		case float64:
			attributes = append(attributes, attribute.Float64(key, v)) // handle float64
		case float32:
			attributes = append(attributes, attribute.Float64(key, float64(v))) // cast float32 to float64
		case []string:
			for _, s := range v {
				attributes = append(attributes, attribute.String(key, s)) // handle []string by appending each string
			}
		case []int:
			for _, i := range v {
				attributes = append(attributes, attribute.Int(key, i)) // handle []int by appending each int
			}
		case []float64:
			for _, f := range v {
				attributes = append(attributes, attribute.Float64(key, f)) // handle []float64 by appending each float64
			}
		default:
			fmt.Printf("Unsupported attribute type for key: %s\n", key)
		}
	}

	// Get the MW_CUSTOM_RESOURCE_ATTRIBUTES environment variable
	envResourceAttributes := os.Getenv("MW_CUSTOM_RESOURCE_ATTRIBUTES")
	// Split the attributes by comma
	attrs := strings.Split(envResourceAttributes, ",")
	for _, attr := range attrs {
		// Split each attribute by the '=' character
		kv := strings.SplitN(attr, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			attributes = append(attributes, attribute.String(key, value))
		}
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attributes...
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
