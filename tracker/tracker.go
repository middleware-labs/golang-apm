package tracker

import (
	"context"
	"log"

	"github.com/middleware-labs/golang-apm/logger"
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
