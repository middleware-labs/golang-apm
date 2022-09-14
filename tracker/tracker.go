package tracker

import (
	"log"
)

func Track(opts ...Options) error {
	c := newConfig(opts...)
	if c.pauseTraces == false {
		initTracer(c)
	}
	if c.pauseMetrics == false {
		handler := Tracer{}
		err := handler.init(c)
		if err != nil {
			log.Fatalf("failed to create the collector exporter: %v", err)
			return err
		}
	}
	return nil
}
