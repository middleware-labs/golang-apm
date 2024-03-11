package tracker

import (
	"github.com/middleware-labs/golang-apm/logger"
)

func Track(opts ...Options) (*Config, error) {
	c := newConfig(opts...)
	logger.InitLogger(c.ServiceName, c.AccessToken, c.fluentHost, c.isServerless)
	if c.pauseTraces == false {
		initTracer(c)
	}
	handler := Metrics{}
	go handler.init(c)
	return c, nil
}
