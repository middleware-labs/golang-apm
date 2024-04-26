package tracker

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/grafana/pyroscope-go"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	ServiceName string

	projectName string

	Host string

	pauseMetrics bool

	pauseTraces bool

	settings map[string]interface{}

	enableProfiling bool

	TenantID string

	AccessToken string

	target string

	fluentHost string

	isServerless string

	Tp *sdktrace.TracerProvider

	Mp *sdkmetric.MeterProvider
}

type Options func(*Config)

func WithConfigTag(k string, v interface{}) Options {
	return func(c *Config) {
		if c.settings == nil {
			c.settings = make(map[string]interface{})
		}
		c.settings[k] = v
	}
}
func doesNotContainHTTP(s string) bool {
	return !(strings.Contains(s, "http://") || strings.Contains(s, "https://"))
}

func newConfig(opts ...Options) *Config {
	c := new(Config)
	c.pauseMetrics = false
	c.pauseTraces = false
	c.enableProfiling = true
	c.fluentHost = "localhost"
	profilingServerUrl := os.Getenv("MW_PROFILING_SERVER_URL")
	if profilingServerUrl == "" {
		profilingServerUrl = "https://profiling.middleware.io"
	}
	authUrl := os.Getenv("MW_AUTH_URL")
	if authUrl == "" {
		authUrl = "https://app.middleware.io/api/v1/auth"
	}

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
	if c.enableProfiling == true {
		if v, ok := c.settings["enableProfiling"]; ok {
			if s, ok := v.(bool); ok {
				c.enableProfiling = s
			}
		}
	}

	if c.ServiceName == "" {
		if v, ok := c.settings["service"]; ok {
			if s, ok := v.(string); ok {
				c.ServiceName = s
			}
		} else {
			c.ServiceName = "Service-" + pid
		}
	}

	if c.target == "" {
		if v, ok := c.settings["target"]; ok {
			if s, ok := v.(string); ok {
				os.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "false")
				c.target = s
				target := s
				if doesNotContainHTTP(target) {
					target = "https://" + target
				}
				parsedURL, err := url.Parse(target)
				if err != nil {
					log.Println("url parse error", err)
				}
				hostnameParts := strings.SplitN(parsedURL.Hostname(), ".", 2)
				c.fluentHost = "fluent." + hostnameParts[len(hostnameParts)-1]
				c.isServerless = "1"
			}
		} else {
			os.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
			c.target = "localhost:9319"
			c.isServerless = "0"
		}
	}

	c.Host = getHostValue("MW_AGENT_SERVICE", c.target)

	if c.projectName == "" {
		if v, ok := c.settings["projectName"]; ok {
			if s, ok := v.(string); ok {
				c.projectName = s
			}
		} else {
			c.projectName = "Project-" + pid
		}
	}

	if c.AccessToken == "" {
		if v, ok := c.settings["accessToken"]; ok {
			if s, ok := v.(string); ok {
				c.AccessToken = s
			}
		}
	}

	if c.enableProfiling && c.AccessToken == "" {
		log.Println("Middleware accessToken is required for Profiling")
	}

	if c.enableProfiling && c.AccessToken != "" {
		req, err := http.NewRequest("POST", authUrl, nil)
		if err != nil {
			log.Println("Error creating request:", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error making auth request")
			return c
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error reading Middleware auth API response")
				return c
			}
			var data map[string]interface{}
			err = json.Unmarshal([]byte(string(body)), &data)
			if err != nil {
				log.Println("Error parsing Middleware JSON")
				return c
			}
			if data["success"] == true {
				account, ok := data["data"].(map[string]interface{})["account"].(string)
				if !ok {
					log.Println("Failed to retrieve TenantID from  api response")
					return c
				}
				c.TenantID = account
				pyroscope.Start(pyroscope.Config{
					ApplicationName: c.ServiceName,
					ServerAddress:   profilingServerUrl,
					TenantID:        c.TenantID,
					ProfileTypes: []pyroscope.ProfileType{
						pyroscope.ProfileCPU,
						pyroscope.ProfileInuseObjects,
						pyroscope.ProfileAllocObjects,
						pyroscope.ProfileInuseSpace,
						pyroscope.ProfileAllocSpace,
					},
				})
			}
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
