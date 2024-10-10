package tracker

import (
	"context"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"go.opentelemetry.io/otel"

	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
)

type Metrics struct {
	meter  api.Meter
	gauges map[string]api.Float64ObservableGauge
}

var MeterProvider api.MeterProvider

func (t *Metrics) initMetrics(ctx context.Context, c *Config) error {
	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(c.Host),
		otlpmetricgrpc.WithTemporalitySelector(deltaSelector),
		// Gzip Compression
		otlpmetricgrpc.WithCompressor("gzip"),
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
		consoleExporter, err = stdoutmetric.New(
			stdoutmetric.WithPrettyPrint(),
			stdoutmetric.WithWriter(file),
			stdoutmetric.WithTemporalitySelector(deltaSelector),
		)
		if err != nil {
			log.Println("failed to create debug console exporter for metrics: ", err)
		}
	}

	attributes := []attribute.KeyValue{
		attribute.String("service.name", c.ServiceName),
		attribute.String("library.language", "go"),
		attribute.Bool("mw_agent", true),
		attribute.String("project.name", c.projectName),
		attribute.Bool("runtime.metrics.go", true),
		attribute.String("mw.app.lang", "go"),
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
		log.Println("failed to set resources for metrics:", err)
	}

	if c.debug {
		c.Mp = metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(10*time.Second))),
			metric.WithReader(metric.NewPeriodicReader(consoleExporter, metric.WithInterval(10*time.Second))),
			metric.WithResource(resources))
	} else {
		c.Mp = metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(10*time.Second))),
			metric.WithResource(resources))
	}

	MeterProvider = c.Mp
	// Set the global meter provider, overridding any previous values.
	otel.SetMeterProvider(c.Mp)

	if !c.pauseDefaultMetrics {
		err := runtimemetrics.Start(runtimemetrics.WithMeterProvider(MeterProvider))
		if err != nil {
			log.Println("failed to start runtime metrics:", err)
		}

		metrics := NewMetrics()
		metrics.Initialize()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start collecting metrics every 10 seconds
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-ticker.C:
					// Trigger collection if needed, but it's handled by the callback
				case <-ctx.Done():
					return
				}
			}
		}()

		// Simulate the main application running
		select {}

	}
	return nil
}

func NewMetrics() *Metrics {
	meter := otel.Meter("golang-agent")
	return &Metrics{
		meter:  meter,
		gauges: make(map[string]api.Float64ObservableGauge),
	}
}

func (t *Metrics) Initialize() {
	// Create gauges for all metrics once
	metricNames := []string{
		"num_cpu",
		"num_goroutine",
		"num_cgo_call",
		"mem_stats.alloc",
		"mem_stats.total_alloc",
		"mem_stats.sys",
		"mem_stats.lookups",
		"mem_stats.mallocs",
		"mem_stats.frees",
		"mem_stats.heap_alloc",
		"mem_stats.heap_sys",
		"mem_stats.heap_idle",
		"mem_stats.heap_inuse",
		"mem_stats.heap_released",
		"mem_stats.heap_objects",
		"mem_stats.stack_inuse",
		"mem_stats.stack_sys",
		"mem_stats.m_span_inuse",
		"mem_stats.m_span_sys",
		"mem_stats.m_cache_inuse",
		"mem_stats.m_cache_sys",
		"mem_stats.buck_hash_sys",
		"mem_stats.gc_sys",
		"mem_stats.other_sys",
		"mem_stats.next_gc",
		"mem_stats.last_gc",
		"mem_stats.pause_total_ns",
		"mem_stats.num_gc",
		"mem_stats.num_forced_gc",
		"mem_stats.gc_cpu_fraction",
		"gc_stats.pause_quantiles.min",
		"gc_stats.pause_quantiles.25p",
		"gc_stats.pause_quantiles.50p",
		"gc_stats.pause_quantiles.75p",
		"gc_stats.pause_quantiles.max",
	}

	meter := MeterProvider.Meter("github.com/middleware-labs/golang-apm")

	for _, name := range metricNames {
		gauge, err := meter.Float64ObservableGauge(name, api.WithDescription(name))
		if err != nil {
			log.Println("Failed to create gauge:", err)
			return
		}
		t.gauges[name] = gauge
	}

	// Create a slice to hold the Observable values
	observables := make([]api.Observable, 0, len(t.gauges))
	for _, gauge := range t.gauges {
		observables = append(observables, gauge)
	}

	// Register a single callback for all gauges
	_, err := meter.RegisterCallback(t.collectMetrics, observables...)
	if err != nil {
		log.Println("Failed to register callback:", err)
	}
}

func (t *Metrics) collectMetrics(ctx context.Context, observer api.Observer) error {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	observer.ObserveFloat64(t.gauges["num_cpu"], float64(runtime.NumCPU()))
	observer.ObserveFloat64(t.gauges["num_goroutine"], float64(runtime.NumGoroutine()))
	observer.ObserveFloat64(t.gauges["num_cgo_call"], float64(runtime.NumCgoCall()))

	// Memory stats
	observer.ObserveFloat64(t.gauges["mem_stats.alloc"], float64(ms.Alloc))
	observer.ObserveFloat64(t.gauges["mem_stats.total_alloc"], float64(ms.TotalAlloc))
	observer.ObserveFloat64(t.gauges["mem_stats.sys"], float64(ms.Sys))
	observer.ObserveFloat64(t.gauges["mem_stats.lookups"], float64(ms.Lookups))
	observer.ObserveFloat64(t.gauges["mem_stats.mallocs"], float64(ms.Mallocs))
	observer.ObserveFloat64(t.gauges["mem_stats.frees"], float64(ms.Frees))

	// Heap memory statistics
	observer.ObserveFloat64(t.gauges["mem_stats.heap_alloc"], float64(ms.HeapAlloc))
	observer.ObserveFloat64(t.gauges["mem_stats.heap_sys"], float64(ms.HeapSys))
	observer.ObserveFloat64(t.gauges["mem_stats.heap_idle"], float64(ms.HeapIdle))
	observer.ObserveFloat64(t.gauges["mem_stats.heap_inuse"], float64(ms.HeapInuse))
	observer.ObserveFloat64(t.gauges["mem_stats.heap_released"], float64(ms.HeapReleased))
	observer.ObserveFloat64(t.gauges["mem_stats.heap_objects"], float64(ms.HeapObjects))

	// Stack memory statistics
	observer.ObserveFloat64(t.gauges["mem_stats.stack_inuse"], float64(ms.StackInuse))
	observer.ObserveFloat64(t.gauges["mem_stats.stack_sys"], float64(ms.StackSys))

	// Off-heap memory statistics
	observer.ObserveFloat64(t.gauges["mem_stats.m_span_inuse"], float64(ms.MSpanInuse))
	observer.ObserveFloat64(t.gauges["mem_stats.m_span_sys"], float64(ms.MSpanSys))
	observer.ObserveFloat64(t.gauges["mem_stats.m_cache_inuse"], float64(ms.MCacheInuse))
	observer.ObserveFloat64(t.gauges["mem_stats.m_cache_sys"], float64(ms.MCacheSys))
	observer.ObserveFloat64(t.gauges["mem_stats.buck_hash_sys"], float64(ms.BuckHashSys))
	observer.ObserveFloat64(t.gauges["mem_stats.gc_sys"], float64(ms.GCSys))
	observer.ObserveFloat64(t.gauges["mem_stats.other_sys"], float64(ms.OtherSys))

	// Garbage collector statistics
	observer.ObserveFloat64(t.gauges["mem_stats.next_gc"], float64(ms.NextGC))
	observer.ObserveFloat64(t.gauges["mem_stats.last_gc"], float64(ms.LastGC))
	observer.ObserveFloat64(t.gauges["mem_stats.pause_total_ns"], float64(ms.PauseTotalNs))
	observer.ObserveFloat64(t.gauges["mem_stats.num_gc"], float64(ms.NumGC))
	observer.ObserveFloat64(t.gauges["mem_stats.num_forced_gc"], float64(ms.NumForcedGC))
	observer.ObserveFloat64(t.gauges["mem_stats.gc_cpu_fraction"], ms.GCCPUFraction)

	// Collect GC pause quantiles
	var gc debug.GCStats
	debug.ReadGCStats(&gc)

	// Check the length of PauseQuantiles before accessing
	if len(gc.PauseQuantiles) >= 5 {
		for i, p := range []string{"min", "25p", "50p", "75p", "max"} {
			observer.ObserveFloat64(t.gauges["gc_stats.pause_quantiles."+p], float64(gc.PauseQuantiles[i]))
		}
	}
	return nil
}

func deltaSelector(kind metric.InstrumentKind) metricdata.Temporality {
	switch kind {
	case metric.InstrumentKindCounter,
		metric.InstrumentKindGauge,
		metric.InstrumentKindHistogram,
		metric.InstrumentKindObservableGauge,
		metric.InstrumentKindObservableCounter:
		return metricdata.DeltaTemporality
	case metric.InstrumentKindUpDownCounter,
		metric.InstrumentKindObservableUpDownCounter:
		return metricdata.CumulativeTemporality
	}
	panic("unknown instrument kind")
}
