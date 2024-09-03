package tracker

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/grafana/pyroscope-go"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ConfigTag string

const (
	PauseMetrics        ConfigTag = "pauseMetrics"        // Boolean - disable all metrics
	PauseDefaultMetrics ConfigTag = "pauseDefaultMetrics" // Boolean - disable default runtime metrics
	PauseTraces         ConfigTag = "pauseTraces"         // Boolean - disable all traces
	PauseLogs           ConfigTag = "pauseLogs"           // Boolean - disable all logs
	PauseProfiling      ConfigTag = "pauseProfiling"      // Boolean - disable profiling
	Debug               ConfigTag = "debug"               // Boolean - enable debug in console
	DebugLogFile        ConfigTag = "debugLogFile"        // Boolean - get logs files for debug mode
	Service             ConfigTag = "service"             // String - Service Name e.g: "My-Service"
	Target              ConfigTag = "target"              // String - Target e.g: "app.middleware.io:443"
	Project             ConfigTag = "projectName"         // String - Project Name e.g: "My-Project"
	Token               ConfigTag = "accessToken"         // String - Token string found at agent installation
)

type Config struct {
	ServiceName string

	projectName string

	Host string

	pauseMetrics bool

	pauseDefaultMetrics bool

	pauseTraces bool

	pauseLogs bool

	settings map[ConfigTag]interface{}

	pauseProfiling bool

	debug bool

	debugLogFile bool

	TenantID string

	AccessToken string

	target string

	fluentHost string

	isServerless string

	Tp *sdktrace.TracerProvider

	Mp *sdkmetric.MeterProvider

	Lp *sdklog.LoggerProvider
}

type Options func(*Config)

// Add Config Options using ConfigTag e.g: track.WithConfigTag(track.Service, "my-service")
func WithConfigTag(k ConfigTag, v interface{}) Options {
	return func(c *Config) {
		if c.settings == nil {
			c.settings = make(map[ConfigTag]interface{})
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
	c.pauseDefaultMetrics = false
	c.pauseTraces = false
	c.pauseProfiling = false
	c.fluentHost = "localhost"
	profilingServerUrl := os.Getenv("MW_PROFILING_SERVER_URL")
	authUrl := os.Getenv("MW_AUTH_URL")
	if authUrl == "" {
		authUrl = "https://app.middleware.io/api/v1/auth"
	}

	pid := strconv.Itoa(os.Getpid())
	for _, fn := range opts {
		fn(c)
	}
	if !c.pauseMetrics {
		if v, ok := c.settings["pauseMetrics"]; ok {
			if s, ok := v.(bool); ok {
				c.pauseMetrics = s
			}
		}
	}
	if !c.pauseDefaultMetrics {
		if v, ok := c.settings["pauseDefaultMetrics"]; ok {
			if s, ok := v.(bool); ok {
				c.pauseDefaultMetrics = s
			}
		}
	}
	if !c.pauseTraces {
		if v, ok := c.settings["pauseTraces"]; ok {
			if s, ok := v.(bool); ok {
				c.pauseTraces = s
			}
		}
	}
	if !c.pauseLogs {
		if v, ok := c.settings["pauseLogs"]; ok {
			if s, ok := v.(bool); ok {
				c.pauseLogs = s
			}
		}
	}
	if !c.pauseProfiling {
		if v, ok := c.settings["pauseProfiling"]; ok {
			if s, ok := v.(bool); ok {
				c.pauseProfiling = s
			}
		}
	}
	if !c.debug {
		if v, ok := c.settings["debug"]; ok {
			if s, ok := v.(bool); ok {
				c.debug = s
			}
		}
	}
	if !c.debugLogFile {
		if v, ok := c.settings["debugLogFile"]; ok {
			if s, ok := v.(bool); ok {
				c.debugLogFile = s
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

	if !c.pauseProfiling && c.AccessToken == "" {
		log.Println("Middleware accessToken is required for Profiling")
	}

	if !c.pauseProfiling && c.AccessToken != "" {
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
			body, err := io.ReadAll(resp.Body)
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
				if profilingServerUrl == "" {
					profilingServerUrl = fmt.Sprint("https://" + account + ".middleware.io/profiling")
				}
				c.TenantID = account
				_, err := pyroscope.Start(pyroscope.Config{
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
				if err != nil {
					log.Println("failed to enable continuous profiling: ", err)
				}
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
