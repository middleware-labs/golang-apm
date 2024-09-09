module github.com/middleware-labs/golang-apm

go 1.21

toolchain go1.22.2

require (
	github.com/agoda-com/opentelemetry-go/otelzerolog v0.0.1
	github.com/agoda-com/opentelemetry-logs-go v0.5.0
	github.com/fluent/fluent-logger-golang v1.9.0
	github.com/go-errors/errors v1.5.1
	github.com/grafana/pyroscope-go v1.1.2
	github.com/samber/slog-multi v1.1.0
	github.com/sirupsen/logrus v1.9.3
	go.opentelemetry.io/contrib/bridges/otellogrus v0.2.0
	go.opentelemetry.io/contrib/bridges/otelslog v0.2.0
	go.opentelemetry.io/contrib/bridges/otelzap v0.0.0-20240611215918-b368fc0c6318
	go.opentelemetry.io/contrib/instrumentation/runtime v0.51.0
	go.opentelemetry.io/contrib/propagators/b3 v1.22.0
	go.opentelemetry.io/otel v1.29.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.3.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.45.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.27.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.27.0
	go.opentelemetry.io/otel/metric v1.29.0
	go.opentelemetry.io/otel/sdk v1.29.0
	go.opentelemetry.io/otel/sdk/log v0.5.0
	go.opentelemetry.io/otel/sdk/metric v1.29.0
	go.opentelemetry.io/otel/trace v1.29.0
)

require (
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.8 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/rs/zerolog v1.30.0 // indirect
	github.com/samber/lo v1.38.1 // indirect
	github.com/tinylib/msgp v1.1.9 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.5.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.29.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.29.0 // indirect
	go.opentelemetry.io/otel/log v0.5.0 // indirect
	go.opentelemetry.io/proto/otlp v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/exp v0.0.0-20220303212507-bbda1eaf7a17 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240520151616-dc85e6b867a5 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117 // indirect
	google.golang.org/grpc v1.64.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)
