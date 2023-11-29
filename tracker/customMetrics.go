package tracker

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"

	api "go.opentelemetry.io/otel/metric"
)

func CollectInt64Counter(name string) api.Int64Counter {
	meter := otel.Meter("golang-agent")
	counter, err := meter.Int64Counter(name, api.WithDescription(""))
	if err != nil {
		log.Println(err)
	}
	return counter
}

func CollectFloat64Counter(name string) api.Float64Counter {
	meter := otel.Meter("golang-agent")
	counter, err := meter.Float64Counter(name, api.WithDescription(""))
	if err != nil {
		log.Println(err)
	}

	return counter
}

func CollectInt64UpDownCounter(name string) api.Int64UpDownCounter {
	meter := otel.Meter("golang-agent")
	counter, err := meter.Int64UpDownCounter(name, api.WithDescription(""))
	if err != nil {
		log.Println(err)
	}

	return counter
}

func CollectFloat64UpDownCounter(name string) api.Float64UpDownCounter {
	meter := otel.Meter("golang-agent")
	counter, err := meter.Float64UpDownCounter(name, api.WithDescription(""))
	if err != nil {
		log.Println(err)
	}
	return counter
}

func CollectInt64Histogram(name string) api.Int64Histogram {
	meter := otel.Meter("golang-agent")
	counter, err := meter.Int64Histogram(name, api.WithDescription(""))
	if err != nil {
		log.Println(err)
	}
	return counter
}

func CollectFloat64Histogram(name string) api.Float64Histogram {
	meter := otel.Meter("golang-agent")
	counter, err := meter.Float64Histogram(name, api.WithDescription(""))
	if err != nil {
		log.Println(err)
	}
	return counter
}

func CollectInt64Gauge(name string, value int64) api.Int64ObservableGauge {
	meter := otel.Meter("golang-agent")
	gauge, err := meter.Int64ObservableGauge(name, api.WithDescription(""))
	if err != nil {
		log.Println(err)
	}
	_, err = meter.RegisterCallback(func(_ context.Context, o api.Observer) error {
		o.ObserveInt64(gauge, value)
		return nil
	}, gauge)
	if err != nil {
		log.Println(err)
	}
	return gauge
}

func CollectFloat64Gauge(name string, value float64) api.Float64ObservableGauge {
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
	return gauge
}
