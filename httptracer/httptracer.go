package httptracer

import (
	"strings"
    "context"
    "fmt"
    "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/attribute"
	"os"
)

type ConfigTag string

type Config struct {
	ServiceName string

	AccessToken string

	target string

	settings map[ConfigTag]interface{}

}

func WithConfigTag(k ConfigTag, v interface{}) Options {
	return func(c *Config) {
		if c.settings == nil {
			c.settings = make(map[ConfigTag]interface{})
		}
		c.settings[k] = v
	}
}

type Options func(*Config)

func newConfig(opts ...Options) *Config {
	c := new(Config)
	for _, fn := range opts {
		fn(c)
	}

	if c.ServiceName == "" {
		if v, ok := c.settings["service"]; ok {
			if s, ok := v.(string); ok {
				c.ServiceName = s
			}
		} else {
			c.ServiceName = "Defualt-Service"
		}
	}
	if c.target == "" {
		if v, ok := c.settings["target"]; ok {
			if s, ok := v.(string); ok {
				c.target = s
			}
		} else {
			os.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
			c.target = "localhost:9319"
		}
	}
	
	c.target = removeHTTP(c.target)

	if c.AccessToken == "" {
		if v, ok := c.settings["accessToken"]; ok {
			if s, ok := v.(string); ok {
				c.AccessToken = s
			}
		}
	}
	return c
}

func removeHTTP(s string) string {
	if strings.Contains(s, "http://") {
		s = strings.ReplaceAll(s, "http://", "")
	}
	if strings.Contains(s, "https://") {
		s = strings.ReplaceAll(s, "https://", "")
	}
	return s
}

// Initialize TracerProvider
func Initialize(opts ...Options) (*trace.TracerProvider, error) {
	c := newConfig(opts...)
    ctx := context.Background()
	
	
	// Create an OTLP HTTP exporter
    exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(c.target))
    if err != nil {
        return nil, fmt.Errorf("failed to create OTLP exporter: %v", err)
    }

    // Create a TracerProvider with the OTLP exporter
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(resource.NewSchemaless(
            attribute.String("service.name", c.ServiceName),
			attribute.String("library.language", "go"),
			attribute.String("mw.account_key", c.AccessToken),
        )),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}
