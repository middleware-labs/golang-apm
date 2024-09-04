package tracker

import (
	"context"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/otel"

	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type Metrics struct{}

var MeterProvider metric.MeterProvider

func (t *Metrics) initMetrics(ctx context.Context, c *Config) error {
	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(c.Host),
	)
	if err != nil {
		log.Println("failed to create exporter for metrics: ", err)
	}

	var file *os.File = os.Stdout
	var consoleExporter metric.Exporter
	if c.debug {
		if c.debugLogFile {
			file, err = os.OpenFile("./mw-metrics.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
			if err != nil {
				log.Println("failed to create exporter file for metrics: ", err)
			}
		}
		consoleExporter, err = stdoutmetric.New(stdoutmetric.WithPrettyPrint(), stdoutmetric.WithWriter(file))
		if err != nil {
			log.Println("failed to create debug console exporter for metrics: ", err)
		}
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", c.ServiceName),
			attribute.String("library.language", "go"),
			attribute.Bool("mw_agent", true),
			attribute.String("project.name", c.projectName),
			attribute.Bool("runtime.metrics.go", true),
			attribute.String("mw.app.lang", "go"),
			attribute.String("mw.account_key", c.AccessToken),
			attribute.String("mw_serverless", c.isServerless),
		),
	)
	if err != nil {
		log.Println("failed to set resources for metrics:", err)
	}

	if c.debug {
		MeterProvider = *metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(10*time.Second))),
			metric.WithReader(metric.NewPeriodicReader(consoleExporter, metric.WithInterval(10*time.Second))),
			metric.WithResource(resources))
	} else {
		MeterProvider = *metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(10*time.Second))),
			metric.WithResource(resources))
	}

	c.Mp = &MeterProvider

	otel.SetMeterProvider(&MeterProvider)

	if !c.pauseDefaultMetrics {
		err := runtimemetrics.Start(runtimemetrics.WithMeterProvider(&MeterProvider))
		if err != nil {
			log.Println("failed to start runtime metrics:", err)
		}

		tick := time.NewTicker(10 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				t.collectMetrics(ctx)
			}
		}
	}
	return nil
}

func (t *Metrics) collectMetrics(ctx context.Context) {
	var ms runtime.MemStats
	gc := debug.GCStats{
		PauseQuantiles: make([]time.Duration, 5),
	}
	debug.ReadGCStats(&gc)
	runtime.ReadMemStats(&ms)
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

	//  Heap memory statistics
	t.createMetric("mem_stats.heap_alloc", float64(ms.HeapAlloc))
	t.createMetric("mem_stats.heap_sys", float64(ms.HeapSys))
	t.createMetric("mem_stats.heap_idle", float64(ms.HeapIdle))
	t.createMetric("mem_stats.heap_inuse", float64(ms.HeapInuse))
	t.createMetric("mem_stats.heap_released", float64(ms.HeapReleased))
	t.createMetric("mem_stats.heap_objects", float64(ms.HeapObjects))
	// Stack memory statistics
	t.createMetric("mem_stats.stack_inuse", float64(ms.StackInuse))
	t.createMetric("mem_stats.stack_sys", float64(ms.StackSys))
	// Off-heap memory statistics
	t.createMetric("mem_stats.m_span_inuse", float64(ms.MSpanInuse))
	t.createMetric("mem_stats.m_span_sys", float64(ms.MSpanSys))
	t.createMetric("mem_stats.m_cache_inuse", float64(ms.MCacheInuse))
	t.createMetric("mem_stats.m_cache_sys", float64(ms.MCacheSys))
	t.createMetric("mem_stats.buck_hash_sys", float64(ms.BuckHashSys))
	t.createMetric("mem_stats.gc_sys", float64(ms.GCSys))
	t.createMetric("mem_stats.other_sys", float64(ms.OtherSys))
	// Garbage collector statistics
	t.createMetric("mem_stats.next_gc", float64(ms.NextGC))
	t.createMetric("mem_stats.last_gc", float64(ms.LastGC))
	t.createMetric("mem_stats.pause_total_ns", float64(ms.PauseTotalNs))
	t.createMetric("mem_stats.num_gc", float64(ms.NumGC))
	t.createMetric("mem_stats.num_forced_gc", float64(ms.NumForcedGC))
	t.createMetric("mem_stats.gc_cpu_fraction", ms.GCCPUFraction)
	for i, p := range []string{"min", "25p", "50p", "75p", "max"} {
		t.createMetric("gc_stats.pause_quantiles."+p, float64(gc.PauseQuantiles[i]))
	}
}

func (t *Metrics) createMetric(name string, value float64) {
	meter := otel.Meter("golang-agent")
	gauge, err := meter.Float64ObservableGauge(name, api.WithDescription(""))
	if err != nil {
		log.Println(err)
	}
	_, err = meter.RegisterCallback(func(_ context.Context, o api.Observer) error {
		o.ObserveFloat64(gauge, value)
		return nil
	}, gauge)
	if err != nil {
		log.Println(err)
	}
}
