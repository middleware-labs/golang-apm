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
go track.Track(
		track.WithConfigTag("service", "your service name"),
		track.WithConfigTag("projectName", "your project name"),
                track.WithConfigTag("accessToken", "your API key"),
	)
```
## Import Application logs

### Open-telemetry Loggers

| Logger                         | Version | Minimal go version |
|--------------------------------|---------|--------------------|
| [mwotelslog](mwotelslog)       | v0.1.0  | 1.21               |
| [mwotelzap](mwotelzap)         | v0.2.1  | 1.20               |
| [mwotelzerolog](mwotelzerolog) | v0.0.1  | 1.20               |

### `log/slog`
```go
    config, _ := track.Track(
		track.WithConfigTag("service", "your service name"),
		track.WithConfigTag("projectName", "your project name"),
	)

	logger := slog.New(
	//use slog-multi if logging in console is needed with stderr handler.
		sm.Fanout(
			slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}),
			mwotelslog.NewMWOtelHandler(config, mwotelslog.HandlerOptions{}),
		),
	)
	//configure default logger
	slog.SetDefault(logger)
```
Add NewMWOtelHandler with config from tracker config. 

This will start collecting the application log from slog and standard library logs.

See [mwotelslog]() features sample for more details.
### `zap`
```go
    config, _ := track.Track(
		track.WithConfigTag("service", "your service name"),
		track.WithConfigTag("projectName", "your project name"),
	)

	logger := zap.New(zapcore.NewTee(consoleCore, fileCore, mwotelzap.NewMWOtelCore(config)))
	zap.ReplaceGlobals(logger)
```
Add NewMWOtelCore with config from tracker config. 

This will start collecting the application log from zap.

See [mwotelzap]()  features sample for more details.

### `zerolog`
```go
    config, _ := track.Track(
		track.WithConfigTag("service", "your service name"),
		track.WithConfigTag("projectName", "your project name"),
	)
	hook := mwotelzerolog.NewMWOtelHook(config)
	logger := log.Hook(hook)
```
Add NewMWOtelHook with config from tracker config. 

This will start collecting the application log from zerolog.

See [mwotelzerolog]() features sample for more details.

## Collect Application Profiling Data

If you also want to collect profiling data for your application,
simply add this one config to your track.Track() call

```go
track.WithConfigTag("accessToken", "{ACCOUNT_KEY}")
```

## Add custom logs

```go
"github.com/middleware-labs/golang-apm/logger"

....

logger.Error("Error")
logger.Info("Info")
logger.Warn("Warn")
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

