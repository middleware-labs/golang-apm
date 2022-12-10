package tracker

import (
	"github.com/middleware-labs/golang-apm/logger"
)

func Track(opts ...Options) (*Config, error) {
	c := newConfig(opts...)
	logger.InitLogger(c.projectName, c.ServiceName)
	if c.pauseTraces == false {
		initTracer(c)
	}
	if c.pauseMetrics == false {
		handler := Metrics{}
		go handler.init(c)
	}
	return c, nil
}
