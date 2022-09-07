package tracker

import (
	"github.com/middleware-labs/golang-apm/packages/metrics"
	"log"
	"time"
)

func Track() error {
	handler := metrics.Tracer{}
	err := handler.Init()
	if err != nil {
		log.Fatalf("failed to create the collector exporter: %v", err)
	}
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			handler.CollectMetrics()
		}
	}
	return nil
}
