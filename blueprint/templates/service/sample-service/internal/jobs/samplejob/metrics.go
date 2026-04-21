package samplejob

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	meterName = "example.com/sample-service/internal/jobs/samplejob"
	jobName   = "sample-job"
)

type metrics struct {
	success  metric.Int64Counter
	errors   metric.Int64Counter
	duration metric.Float64Histogram
	work     metric.Int64Counter
}

func newMetrics() (*metrics, error) {
	meter := otel.Meter(meterName)

	success, err := meter.Int64Counter(
		"sample.job.success",
		metric.WithDescription("Number of successful sample job completions."),
		metric.WithUnit("{completion}"),
	)
	if err != nil {
		return nil, err
	}

	errors, err := meter.Int64Counter(
		"sample.job.errors",
		metric.WithDescription("Number of sample job failures by reason."),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return nil, err
	}

	duration, err := meter.Float64Histogram(
		"sample.job.duration",
		metric.WithDescription("Sample job run duration."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	work, err := meter.Int64Counter(
		"sample.job.work.items",
		metric.WithDescription("Number of work items processed by the sample job."),
		metric.WithUnit("{item}"),
	)
	if err != nil {
		return nil, err
	}

	return &metrics{
		success:  success,
		errors:   errors,
		duration: duration,
		work:     work,
	}, nil
}

func (m *metrics) recordSuccess(ctx context.Context) {
	if m == nil {
		return
	}
	m.success.Add(ctx, 1, metric.WithAttributes(jobNameAttr()))
}

func (m *metrics) recordError(ctx context.Context, reason string) {
	if m == nil {
		return
	}
	m.errors.Add(ctx, 1, metric.WithAttributes(jobNameAttr(), attribute.String("reason", reason)))
}

func (m *metrics) recordDuration(ctx context.Context, result string, start time.Time) {
	if m == nil {
		return
	}
	m.duration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(jobNameAttr(), attribute.String("result", result)))
}

func (m *metrics) recordWork(ctx context.Context, result string, count int) {
	if m == nil || count <= 0 {
		return
	}
	m.work.Add(ctx, int64(count), metric.WithAttributes(jobNameAttr(), attribute.String("result", result)))
}

func jobNameAttr() attribute.KeyValue {
	return attribute.String("job_name", jobName)
}
