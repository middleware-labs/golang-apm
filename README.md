# Install Golang package

```
go get github.com/middleware-labs/golang-apm
```

# Import Tracker

Add this import statement in your project.

```
import (
    track "github.com/middleware-labs/golang-apm/tracker"
)
```

Add this snippet in your main function

```
track.Track(
		track.WithConfigTag("service", "your service name"),
		track.WithConfigTag("projectName", "your project name"),
                track.WithConfigTag("accessToken", "your API key"),
	)
```

# Custom Logs

To ingest custom logs into Middleware, you can use library functions as given below.

```
"github.com/middleware-labs/golang-apm/logger"

....

logger.Error("Error")
logger.Info("Info")
logger.Warn("Warn")

```

# Stack Error

If you want to record exception in traces then you can use track.RecordError(ctx,error) method.

```
r.GET("/books", func(c *gin.Context) {
    ctx := req.Context()
    if err := db.Ping(ctx); err != nil {
        track.RecordError(ctx, err)
    }
})
```

