package tracker

import (
	"os"
	"path/filepath"
)

type config struct {
	serviceName string

	projectName string

	host string

	apiKey string

	pauseMetrics bool

	pauseTraces bool

	settings map[string]interface{}
}

type Options func(*config)

func WithConfigTag(k string, v interface{}) Options {
	return func(c *config) {
		if c.settings == nil {
			c.settings = make(map[string]interface{})
		}
		c.settings[k] = v
	}
}

func newConfig(opts ...Options) *config {
	c := new(config)
	c.pauseMetrics = false
	c.pauseTraces = false
	c.apiKey = ""
	for _, fn := range opts {
		fn(c)
	}
	if c.pauseMetrics == false {
		if v, ok := c.settings["pauseMetrics"]; ok {
			if s, ok := v.(bool); ok {
				c.pauseMetrics = s
			}
		}
	}
	if c.pauseTraces == false {
		if v, ok := c.settings["pauseTraces"]; ok {
			if s, ok := v.(bool); ok {
				c.pauseTraces = s
			}
		}
	}
	if c.apiKey == "" {
		if v, ok := c.settings["apiKey"]; ok {
			if s, ok := v.(string); ok {
				c.apiKey = s
			}
		}
	}
	if c.serviceName == "" {
		if v, ok := c.settings["service"]; ok {
			if s, ok := v.(string); ok {
				c.serviceName = s
			}
		} else {
			c.serviceName = filepath.Base(os.Args[0])
		}
	}
	if c.host == "" {
		if v, ok := c.settings["host"]; ok {
			if s, ok := v.(string); ok {
				c.host = s
			}
		} else {
			c.host = "localhost:4320"
		}
	}
	if c.projectName == "" {
		if v, ok := c.settings["projectName"]; ok {
			if s, ok := v.(string); ok {
				c.projectName = s
			}
		} else {
			c.projectName = filepath.Base(os.Args[0])
		}
	}
	return c
}
