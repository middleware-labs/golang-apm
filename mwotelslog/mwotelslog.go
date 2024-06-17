package mwotelslog

import (
	"log/slog"
	"os"

	"github.com/middleware-labs/golang-apm/tracker"
	sm "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	otellog "go.opentelemetry.io/otel/sdk/log"
)

const loggerName = "mwslog"
const MWTraceID = "traceId"
const MWSpanID = "spanId"

type config struct {
	provider    *otellog.LoggerProvider
	name        string
	consoleLog  bool
	handlerOpts *slog.HandlerOptions
}

type Option interface {
	apply(config) config
}

type optFunc func(config) config

func (f optFunc) apply(c config) config { return f(c) }

func WithName(name string) Option {
	return optFunc(func(c config) config {
		c.name = name
		return c
	})
}

func WithConsoleLog(ho *slog.HandlerOptions) Option {
	return optFunc(func(c config) config {
		c.consoleLog = true
		c.handlerOpts = ho
		return c
	})
}
func WithDefaultConsoleLog() Option {
	return optFunc(func(c config) config {
		c.consoleLog = true
		c.handlerOpts = &slog.HandlerOptions{}
		return c
	})
}

func newConfig(cfg *tracker.Config, options []Option) config {
	var c config
	for _, opt := range options {
		c = opt.apply(c)
	}

	if c.name == "" {
		c.name = loggerName
	}
	if c.provider == nil {
		c.provider = cfg.Lp
	}

	return c
}

func NewMWOTelLogger(config *tracker.Config, options ...Option) *slog.Logger {
	cfg := newConfig(config, options)
	if cfg.consoleLog {
		return slog.New(
			sm.Fanout(
				slog.NewTextHandler(os.Stderr, cfg.handlerOpts),
				otelslog.NewLogger(cfg.name, otelslog.WithLoggerProvider(cfg.provider)).Handler(),
			),
		)
	}
	return otelslog.NewLogger(cfg.name, otelslog.WithLoggerProvider(cfg.provider))
}
