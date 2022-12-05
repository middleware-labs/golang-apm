# golang-apm

go get github.com/middleware-labs/golang-apm


```golang

import (
	track "github.com/middleware-labs/golang-apm/tracker"
	"github.com/middleware-labs/golang-apm/logger"
)

func main() {
	track.Track(
		track.WithConfigTag("service", "your service name"),
		track.WithConfigTag("projectName", "your project name"),
	)
	
	logger.Error("Error")
	
	logger.Info("Info")
	
	logger.Warn("Warn")
	
	logger.Debug("Debug")
}

```

If you want to record exceptions in traces then you can use track.RecordError(ctx,err) method.

```golang

app.get('/hello', (req, res) => {
    ctx := req.Context()
    try {
        throw ("error");
    } catch (error) {
        track.RecordError(ctx, err)
    }
})

```

