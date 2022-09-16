# golang-apm

go get github.com/middleware-labs/golang-apm


```golang

import (
	track "github.com/middleware-labs/golang-apm/tracker"
	"github.com/middleware-labs/golang-apm/logger"
)

func main() {
	track.Track(
		track.WithConfigTag("service", "service1"),
		track.WithConfigTag("host", "localhost:4320"),
		track.WithConfigTag("apiKey", "36kb1q8i2aqxdpw4keb5z6aq3fz0zayl4"),
		track.WithConfigTag("projectName", "demo-agent-apm"),
		track.WithConfigTag("pauseMetrics", false),
		track.WithConfigTag("pauseTraces", false),
	)
	
	logger.Error("Error")
	
	logger.Info("Info")
	
	logger.Warn("Warn")
}