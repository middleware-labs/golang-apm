# golang-apm

~~~~# middleware-labs/golang-apm

go get github.com/middleware-labs/golang-apm


```golang

import (
	track "github.com/middleware-labs/golang-apm/tracker"
)

func main() {
	track.Track(
		track.WithConfigTag("service", "service1"),
		track.WithConfigTag("host", "localhost:4320"),
		track.WithConfigTag("projectName", "demo-agent-apm"),
		track.WithConfigTag("pauseMetrics", false),
		track.WithConfigTag("pauseTraces", false),
	)
}


```