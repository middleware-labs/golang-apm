package tracker

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/resource"
	"log"
	"runtime"
	"time"

	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

type ClientInterface interface {
	Init(serviceName string) error
	CollectMetrics()
	createMetric(name string, value float64)
}

type Tracer struct{}

func (t *Tracer) init(c *config) error {
	client := otlpmetricgrpc.NewClient(
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(c.host),
	)
	ctx := context.Background()
	exp, err := otlpmetric.New(ctx, client)
	if err != nil {
		log.Fatalf("failed to create the collector exporter: %v", err)
		return err
	}

	defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := exp.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}()
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", c.serviceName),
			attribute.String("library.language", "go"),
			attribute.Bool("mw_agent", true),
			attribute.String("project.name", c.projectName),
		),
	)
	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			exp,
		),
		controller.WithExporter(exp),
		controller.WithCollectPeriod(2*time.Second),
		controller.WithResource(resources),
	)

	global.SetMeterProvider(pusher)

	if err := pusher.Start(ctx); err != nil {
		log.Fatalf("could not start metric controller: %v", err)
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := pusher.Stop(ctx); err != nil {
			otel.Handle(err)
		}
	}()

	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			t.collectMetrics()
		}
	}
	return nil
}

func (t *Tracer) collectMetrics() {
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

func (t *Tracer) createMetric(name string, value float64) {
	ctx := context.Background()
	meter := global.Meter("golang-agent")
	counter, err := meter.SyncFloat64().Counter(name)
	if err != nil {
		log.Fatalf("Failed to create the instrument: %v", err)
	}
	counter.Add(ctx, value)
}
