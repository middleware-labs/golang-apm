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
		track.WithConfigTag("projectName", "demo-agent-apm"),
	)
	
	logger.Error("Error")
	
	logger.Info("Info")
	
	logger.Warn("Warn")
}