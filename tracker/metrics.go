package tracker

import (
	"context"
	"go.opentelemetry.io/otel"
	"log"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type ClientInterface interface {
	Init(ServiceName string) error
	CollectMetrics()
	createMetric(name string, value float64)
}

type Metrics struct{}

func (t *Metrics) init(c *Config) error {
	ctx := context.Background()
	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(c.host),
	)
	if err != nil {
		panic(err)
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", c.ServiceName),
			attribute.String("library.language", "go"),
			attribute.Bool("mw_agent", true),
			attribute.String("project.name", c.projectName),
		),
	)

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exp)),
		metric.WithResource(resources))

	defer func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	otel.SetMeterProvider(meterProvider)
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			t.collectMetrics()
		}
	}
	return nil
}

func (t *Metrics) collectMetrics() {
	var ms runtime.MemStats

	t.createMetric("num_cpu", float64(runtime.NumCPU()))
	t.createMetric("num_goroutine", float64(runtime.NumGoroutine()))
	t.createMetric("num_cgo_call", float64(runtime.NumCgoCall()))

	// Memory stats
	t.createMetric("mem_stats.alloc", float64(ms.Alloc))
	t.createMetric("mem_stats.total_alloc", float64(ms.TotalAlloc))
	t.createMetric("mem_stats.sys", float64(ms.Sys))
	t.createMetric("mem_stats.lookups", float64(ms.Lookups))
	t.createMetric("mem_stats.mallocs", float64(ms.Mallocs))
	t.createMetric("mem_stats.frees", float64(ms.Frees))
}

func (t *Metrics) createMetric(name string, value float64) {
	ctx := context.Background()
	meter := otel.Meter("golang-agent")
	counter, err := meter.Float64Counter(name)
	if err != nil {
		log.Fatalf("Failed to create the instrument: %v", err)
	}
	counter.Add(ctx, value)
}
