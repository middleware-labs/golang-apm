package tracker

func Track(opts ...Options) (*Config, error) {
	c := newConfig(opts...)
	if c.pauseTraces == false {
		initTracer(c)
	}
	if c.pauseMetrics == false {
		handler := Tracer{}
		go handler.init(c)
	}
	return c, nil
}
