package mwotelzerolog

import (
	"github.com/middleware-labs/golang-apm/tracker"
	"github.com/sirupsen/logrus"
	ol "go.opentelemetry.io/contrib/bridges/otellogrus"
	otellog "go.opentelemetry.io/otel/sdk/log"
)

const loggerName = "mwlogrus"
const MWTraceID = "traceId"
const MWSpanID = "spanId"

type config struct {
	provider *otellog.LoggerProvider
	name     string
	levels   []logrus.Level
}

type Option interface {
	apply(config) config
}

type optFunc func(config) config

func (f optFunc) apply(c config) config { return f(c) }

func WithLevels(l []logrus.Level) Option {
	return optFunc(func(c config) config {
		c.levels = l
		return c
	})
}

func WithName(name string) Option {
	return optFunc(func(c config) config {
		c.name = name
		return c
	})
}

func newConfig(cfg *tracker.Config, options []Option) config {
	var c config
	for _, opt := range options {
		c = opt.apply(c)
	}

	if c.levels == nil {
		c.levels = logrus.AllLevels
	}

	if c.name == "" {
		c.name = loggerName
	}
	if c.provider == nil {
		c.provider = cfg.Lp
	}

	return c
}

func NewMWOTelHook(config *tracker.Config, options ...Option) *ol.Hook {
	cfg := newConfig(config, options)
	return ol.NewHook(cfg.name, ol.WithLevels(cfg.levels), ol.WithLoggerProvider(cfg.provider))
}
