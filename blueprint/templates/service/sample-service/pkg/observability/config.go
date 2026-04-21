package observability

import "time"

type Config struct {
	Enabled bool `env:"OTEL_ENABLED, default=false"`

	ServiceName    string `env:"OTEL_SERVICE_NAME, default=sample-service"`
	ServiceVersion string `env:"OTEL_SERVICE_VERSION, default=dev"`
	Environment    string `env:"OTEL_DEPLOYMENT_ENVIRONMENT"`

	TracesEnabled   bool    `env:"OTEL_TRACES_ENABLED, default=true"`
	TraceSampleRate float64 `env:"OTEL_TRACE_SAMPLE_RATE, default=0.01"`

	MetricsEnabled bool          `env:"OTEL_METRICS_ENABLED, default=true"`
	MetricInterval time.Duration `env:"OTEL_METRIC_EXPORT_INTERVAL, default=60s"`
}
