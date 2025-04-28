package tracker

import (
	"context"
	"log"
	"os"

	"github.com/middleware-labs/golang-apm/logger"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"time"
	"strings"
)


func TrackWithCtx(ctx context.Context, opts ...Options) (*Config, error) {

	c := newConfig(opts...)
	logger.InitLogger(c.ServiceName, c.AccessToken, c.fluentHost, c.isServerless)

	if !c.pauseTraces {
		tracesHandler := Traces{}
		errTraces := tracesHandler.initTraces(ctx, c)
		if errTraces != nil {
			log.Println("failed to track traces: ", errTraces)
		}
	}

	if !c.pauseLogs {
		logsHandler := Logs{}
		errLogs := logsHandler.initLogs(ctx, c)
		if errLogs != nil {
			log.Println("failed to track logs: ", errLogs)
		}
	}

	if !c.pauseMetrics {
		metricsHandler := Metrics{}
		go func() {
			errMetrics := metricsHandler.initMetrics(ctx, c)
			if errMetrics != nil {
				log.Println("failed to track metrics: ", errMetrics)
			}
		}()
	}

	return c, nil
}

func Track(opts ...Options) (*Config, error) {
	ctx := context.Background()
	return TrackWithCtx(ctx, opts...)
}


func NewTracerProviderCtx(ctx context.Context, c *Config, serviceName string) *trace.TracerProvider{
	collectorURL := c.Host
	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(collectorURL),
			// Gzip Compression
			otlptracegrpc.WithCompressor("gzip"),
		),
	)

	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	var file *os.File = os.Stdout
	var consoleExporter *stdouttrace.Exporter
	if c.debug {
		if c.debugLogFile {
			file, err = os.OpenFile("./mw-traces.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
			if err != nil {
				log.Println("failed to create exporter file for traces: ", err)
			}
		}
		consoleExporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(file))
		if err != nil {
			log.Println("failed to create debug console exporter for traces: ", err)
		}
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("telemetry.sdk.language", "go"),
			attribute.Bool("mw_agent", true),
			attribute.String("project.name", c.projectName),
			attribute.String("mw.account_key", c.AccessToken),
			attribute.String("mw_serverless", c.isServerless),
	
		),
	)
	
	for key, value := range c.customResourceAttributes {
		switch v := value.(type) {
		case string:
			res, err = resource.Merge(res, resource.NewSchemaless(
				attribute.String(key, v),
			))
			if err != nil {
				log.Println("failed to create resource: %v", err)
			}
		case bool:
			res, err = resource.Merge(res, resource.NewSchemaless(
				attribute.Bool(key, v),
			))
			if err != nil {
				log.Println("failed to create resource: %v", err)
			}
		case int:
			res, err = resource.Merge(res, resource.NewSchemaless(
				attribute.Int(key, v),
			))
			if err != nil {
				log.Println("failed to create resource: %v", err)
			}

		case int64:
			res, err = resource.Merge(res, resource.NewSchemaless(
				attribute.Int64(key, v),
			))
			if err != nil {
				log.Println("failed to create resource: %v", err)
			}
		case float64:
			res, err = resource.Merge(res, resource.NewSchemaless(
				attribute.Float64(key, v),
			))
			if err != nil {
				log.Println("failed to create resource: %v", err)
			}

		case float32:
			res, err = resource.Merge(res, resource.NewSchemaless(
				attribute.Float64(key, float64(v)),
			))
			if err != nil {
				log.Println("failed to create resource: %v", err)
			}

		case []string:
			for _, s := range v {
				res, err = resource.Merge(res, resource.NewSchemaless(
					attribute.String(key, s),
				))
				if err != nil {
					log.Println("failed to create resource: %v", err)
				}
			}
		case []int:
			for _, i := range v {
				res, err = resource.Merge(res, resource.NewSchemaless(
					attribute.Int(key, i),
				))
				if err != nil {
					log.Println("failed to create resource: %v", err)
				}
			}
		case []float64:
			for _, f := range v {
				res, err = resource.Merge(res, resource.NewSchemaless(
					attribute.Float64(key, f),
				))
				if err != nil {
					log.Println("failed to create resource: %v", err)
				}
			}
		default:
			log.Printf("Unsupported attribute type for key: %s\n", key)
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
			res, err = resource.Merge(res, resource.NewSchemaless(
				attribute.String(key, value),
			))
			if err != nil {
				log.Println("failed to create resource: %v", err)
			}
		}
	}
	
	

	var tp *trace.TracerProvider

	if c.debug {
		tp = trace.NewTracerProvider(
			trace.WithResource(res),
			trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exporter,
				trace.WithMaxExportBatchSize(10000), trace.WithBatchTimeout(10*time.Second))),
				trace.WithSpanProcessor(trace.NewSimpleSpanProcessor(consoleExporter)),
		)
	} else {
		tp = trace.NewTracerProvider(
			trace.WithResource(res),
			trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exporter,
				trace.WithMaxExportBatchSize(10000), trace.WithBatchTimeout(10*time.Second))),
		)
	}
	return tp
}