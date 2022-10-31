package tracker

import (
	"os"
	"strconv"
)

type config struct {
	serviceName string

	projectName string

	host string

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
	pid := strconv.Itoa(os.Getpid())
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
	if c.serviceName == "" {
		if v, ok := c.settings["service"]; ok {
			if s, ok := v.(string); ok {
				c.serviceName = s
			}
		} else {
			c.serviceName = "Service-" + pid
		}
	}

	c.host = getHostValue("MW_AGENT_SERVICE", "localhost:9319")

	if c.projectName == "" {
		if v, ok := c.settings["projectName"]; ok {
			if s, ok := v.(string); ok {
				c.projectName = s
			}
		} else {
			c.projectName = "Project-" + pid
		}
	}
	return c
}

func getHostValue(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value + ":9319"
}
