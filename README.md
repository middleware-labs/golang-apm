# Install Golang package

```
go get github.com/middleware-labs/golang-apm@latest
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
		track.WithConfigTag(track.Service, "your-service-name"),
		track.WithConfigTag(track.Token, "your API token"),
	)
```
## Import Application logs

### Open-telemetry Loggers

| Logger                         | Version | Minimal go version |
|--------------------------------|---------|--------------------|
| [mwotelslog](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/bridges/otelslog)       | v0.2.0  | 1.21               |
| [mwotelzap](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/bridges/otelzap)         | v0.0.1  | 1.20               |
| [mwotelzerolog](mwotelzerolog) | v0.0.1  | 1.20               |
| [mwotellogrus](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/bridges/otellogrus)   | v0.2.0  | 1.21               |

### `log/slog`
```go
    config, _ := track.Track(
		track.WithConfigTag(track.Service, "your service name"),
		track.WithConfigTag(track.Project, "your project name"),
	)

	logger := mwotelslog.NewMWOTelLogger(
		config,
		mwotelslog.WithDefaultConsoleLog(), // to enable console log
		mwotelslog.WithName("otelslog"),
	)
	//configure default logger
	slog.SetDefault(logger)
```
Add NewMWOtelHandler with config from tracker config. 

This will start collecting the application log from slog and standard library logs.

See [mwotelslog](https://github.com/middleware-labs/demo-apm/tree/master/golang/features) features sample for more details.

### `zap`
```go
     config, _ := track.Track(
		track.WithConfigTag(track.Service, "your service name"),
		track.WithConfigTag(track.Project, "your project name"),
		track.WithConfigTag(track.Token, "your token"),
	)

	logger := zap.New(zapcore.NewTee(consoleCore, fileCore, mwotelzap.NewMWOTelCore(config, mwotelzap.WithName("otelzaplog"))))
	zap.ReplaceGlobals(logger)
```
Add NewMWOTelCore with config from tracker config. 

This will start collecting the application log from zap.

See [mwotelzap](https://github.com/middleware-labs/demo-apm/tree/master/golang/features)  features sample for more details.

### `zerolog`
```go
    config, _ := track.Track(
		track.WithConfigTag(track.Service, "your service name"),
		track.WithConfigTag(track.Project, "your project name"),
	)
	hook := mwotelzerolog.NewMWOTelHook(config)
	logger := log.Hook(hook)
```
Add NewMWOTelHook with config from tracker config. 

This will start collecting the application log from zerolog.

See [mwotelzerolog](https://github.com/middleware-labs/demo-apm/tree/master/golang/features) features sample for more details.

### `logrus`
```go
     config, _ := track.Track(
		track.WithConfigTag(track.Service, "your service name"),
		track.WithConfigTag(track.Project, "your project name"),
	)

	logHook := otellog.NewMWOTelHook(config, otellog.WithLevels(log.AllLevels), otellog.WithName("otellogrus"))

	// add hook in logrus
	log.AddHook(logHook)
	// set formatter if required
	log.SetFormatter(&log.JSONFormatter{})
```
Add NewMWOTelHook with config from tracker config. 

This will start collecting the application log from logrus.

See [mwotellogrus](https://github.com/middleware-labs/demo-apm/tree/master/golang/features) features sample for more details.

## Collect Application Profiling Data

If you also want to collect profiling data for your application,
simply add this one config to your track.Track() call

```go
track.WithConfigTag(track.Token, "{ACCOUNT_KEY}"),
```

## Custom Logs

To ingest custom logs into Middleware, you can use library functions as given below.

```
"github.com/middleware-labs/golang-apm/logger"

....

logger.Error("Error")
logger.Info("Info")
logger.Warn("Warn")

```

## Stack Error

If you want to record exception in traces then you can use track.RecordError(ctx,error) method.

```
r.GET("/books", func(c *gin.Context) {
    ctx := req.Context()
    if err := db.Ping(ctx); err != nil {
        track.RecordError(ctx, err)
    }
})
```

## Pause Default Metrics

```go
go track.Track(
		track.WithConfigTag(track.PauseDefaultMetrics,true),
	)
```
## Enable Debug Mode with console log

```go
go track.Track(
		track.WithConfigTag(track.Debug, true),
	)
```
## Enable Debug Mode with logs files for Metrics, Traces and Logs

```go
go track.Track(
		track.WithConfigTag(track.Debug, true),
		track.WithConfigTag(track.DebugLogFile, true),
	)
```