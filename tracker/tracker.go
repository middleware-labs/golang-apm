package tracker

import (
	"github.com/middleware-labs/golang-apm/tracker/metrics"
	"log"
)

func Track(serviceName string) error {
	handler := metrics.Tracer{}
	err := handler.Init(serviceName)
	if err != nil {
		log.Fatalf("failed to create the collector exporter: %v", err)
	}
	return nil
}
