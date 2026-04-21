package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
}

func New(ctx context.Context, cfg *Config) (*Provider, error) {
	if cfg == nil || !cfg.Enabled {
		return &Provider{}, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(resourceAttributes(cfg)...),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating otel resource: %w", err)
	}

	provider := &Provider{}

	if cfg.TracesEnabled {
		traceExporter, err := otlptracehttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating otlp trace exporter: %w", err)
		}

		provider.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(clampSampleRate(cfg.TraceSampleRate)))),
			sdktrace.WithBatcher(traceExporter),
		)
		otel.SetTracerProvider(provider.tracerProvider)
	}

	if cfg.MetricsEnabled {
		metricExporter, err := otlpmetrichttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating otlp metric exporter: %w", err)
		}

		interval := cfg.MetricInterval
		if interval <= 0 {
			interval = 60 * time.Second
		}
		reader := sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(interval))
		provider.meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(reader),
		)
		otel.SetMeterProvider(provider.meterProvider)
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider, nil
}

func (p *Provider) Close(ctx context.Context) error {
	if p == nil {
		return nil
	}

	shutdownCtx, done := context.WithTimeout(ctx, 5*time.Second)
	defer done()

	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutting down meter provider: %w", err)
		}
	}
	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutting down tracer provider: %w", err)
		}
	}
	return nil
}

func resourceAttributes(cfg *Config) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", cfg.ServiceVersion),
	}
	if cfg.Environment != "" {
		attrs = append(attrs, attribute.String("deployment.environment.name", cfg.Environment))
	}
	return attrs
}

func clampSampleRate(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
