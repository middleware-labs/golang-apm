package tracker

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"strings"

	goErros "github.com/go-errors/errors"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	"log"
	"time"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Traces struct{}

var TraceProvider sdktrace.TracerProvider

func (t *Traces) initTraces(ctx context.Context, c *Config) error {
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
		log.Println("failed to create exporter for traces: ", err)
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

	attributes := []attribute.KeyValue{
		attribute.String("service.name", c.ServiceName),
		attribute.String("library.language", "go"),
		attribute.Bool("mw_agent", true),
		attribute.String("project.name", c.projectName),
		attribute.String("mw.account_key", c.AccessToken),
		attribute.String("mw_serverless", c.isServerless),
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
			attributes...,
		),
	)

	if err != nil {
		log.Println("failed to set resources for traces:", err)
	}

	if c.debug {
		TraceProvider = *sdktrace.NewTracerProvider(
			sdktrace.WithResource(resources),
			sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter,
				sdktrace.WithMaxExportBatchSize(10000), sdktrace.WithBatchTimeout(10*time.Second))),
			sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(consoleExporter)),
		)
	} else {
		TraceProvider = *sdktrace.NewTracerProvider(
			sdktrace.WithResource(resources),
			sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter,
				sdktrace.WithMaxExportBatchSize(10000), sdktrace.WithBatchTimeout(10*time.Second))),
		)
	}
	otel.SetTracerProvider(&TraceProvider)
	c.Tp = &TraceProvider

	p := b3.New()
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			p,
			propagation.TraceContext{},
			propagation.Baggage{}),
	)
	return err
}

func SpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ""
	}
	return span.SpanContext().SpanID().String()
}

func TraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}

func SpanFromContext(ctx context.Context) trace.Span {
	span := trace.SpanFromContext(ctx)
	return span
}

func WithStackTrace(b bool) trace.SpanEndEventOption {
	return trace.WithStackTrace(b)
}

func ErrorCode() codes.Code {
	return codes.Error
}

func ErrorRecording(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if err != nil {
		errorStack := string(goErros.Wrap(err, 3).Stack())
		span.AddEvent("exception",
			trace.WithAttributes(
				attribute.String("exception.type", "*errors.errorString"),
				attribute.String("exception.stacktrace", errorStack),
				attribute.String("exception.message", err.Error()),
			),
		)
		span.SetStatus(codes.Error, err.Error())
	}
}

func SetAttribute(ctx context.Context, name string, value string) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String(name, value))
}

// Bool creates a attribute.KeyValue with a BOOL Value type.
func Bool(k string, v bool) attribute.KeyValue {
	return attribute.Key(k).Bool(v)
}

// BoolSlice creates a attribute.KeyValue with a BOOLSLICE Value type.
func BoolSlice(k string, v []bool) attribute.KeyValue {
	return attribute.Key(k).BoolSlice(v)
}

// Int creates a attribute.KeyValue with an INT64 Value type.
func Int(k string, v int) attribute.KeyValue {
	return attribute.Key(k).Int(v)
}

// IntSlice creates a attribute.KeyValue with an INT64SLICE Value type.
func IntSlice(k string, v []int) attribute.KeyValue {
	return attribute.Key(k).IntSlice(v)
}

// Int64 creates a attribute.KeyValue with an INT64 Value type.
func Int64(k string, v int64) attribute.KeyValue {
	return attribute.Key(k).Int64(v)
}

// Int64Slice creates a attribute.KeyValue with an INT64SLICE Value type.
func Int64Slice(k string, v []int64) attribute.KeyValue {
	return attribute.Key(k).Int64Slice(v)
}

// Float64 creates a attribute.KeyValue with a FLOAT64 Value type.
func Float64(k string, v float64) attribute.KeyValue {
	return attribute.Key(k).Float64(v)
}

// Float64Slice creates a attribute.KeyValue with a FLOAT64SLICE Value type.
func Float64Slice(k string, v []float64) attribute.KeyValue {
	return attribute.Key(k).Float64Slice(v)
}

// String creates a attribute.KeyValue with a STRING Value type.
func String(k, v string) attribute.KeyValue {
	return attribute.Key(k).String(v)
}

// StringSlice creates a attribute.KeyValue with a STRINGSLICE Value type.
func StringSlice(k string, v []string) attribute.KeyValue {
	return attribute.Key(k).StringSlice(v)
}

// Stringer creates a new key-value pair with a passed name and a string
// value generated by the passed Stringer interface.
func Stringer(k string, v fmt.Stringer) attribute.KeyValue {
	return attribute.Key(k).String(v.String())
}
