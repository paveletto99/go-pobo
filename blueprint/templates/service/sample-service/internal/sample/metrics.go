package sample

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "example.com/sample-service/internal/sample"

type metrics struct {
	itemRequests       metric.Int64Counter
	itemErrors         metric.Int64Counter
	itemsCreated       metric.Int64Counter
	itemLookups        metric.Int64Counter
	itemHandlerLatency metric.Float64Histogram
}

func newMetrics() (*metrics, error) {
	meter := otel.Meter(meterName)

	itemRequests, err := meter.Int64Counter(
		"sample.item.requests",
		metric.WithDescription("Number of item API requests handled by operation."),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	itemErrors, err := meter.Int64Counter(
		"sample.item.errors",
		metric.WithDescription("Number of item API request errors by operation and reason."),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return nil, err
	}

	itemsCreated, err := meter.Int64Counter(
		"sample.item.created",
		metric.WithDescription("Number of items created successfully."),
		metric.WithUnit("{item}"),
	)
	if err != nil {
		return nil, err
	}

	itemLookups, err := meter.Int64Counter(
		"sample.item.lookup",
		metric.WithDescription("Number of item lookups by result."),
		metric.WithUnit("{lookup}"),
	)
	if err != nil {
		return nil, err
	}

	itemHandlerLatency, err := meter.Float64Histogram(
		"sample.item.handler.duration",
		metric.WithDescription("Item handler duration by operation."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return &metrics{
		itemRequests:       itemRequests,
		itemErrors:         itemErrors,
		itemsCreated:       itemsCreated,
		itemLookups:        itemLookups,
		itemHandlerLatency: itemHandlerLatency,
	}, nil
}

func (m *metrics) recordRequest(ctx context.Context, operation string) {
	if m == nil {
		return
	}
	m.itemRequests.Add(ctx, 1, metric.WithAttributes(operationAttr(operation)))
}

func (m *metrics) recordError(ctx context.Context, operation, reason string) {
	if m == nil {
		return
	}
	m.itemErrors.Add(ctx, 1, metric.WithAttributes(
		operationAttr(operation),
		attribute.String("reason", reason),
	))
}

func (m *metrics) recordItemCreated(ctx context.Context) {
	if m == nil {
		return
	}
	m.itemsCreated.Add(ctx, 1)
}

func (m *metrics) recordLookup(ctx context.Context, found bool) {
	if m == nil {
		return
	}
	result := "not_found"
	if found {
		result = "found"
	}
	m.itemLookups.Add(ctx, 1, metric.WithAttributes(attribute.String("result", result)))
}

func (m *metrics) recordHandlerLatency(ctx context.Context, operation string, start time.Time) {
	if m == nil {
		return
	}
	m.itemHandlerLatency.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(operationAttr(operation)))
}

func operationAttr(operation string) attribute.KeyValue {
	return attribute.String("operation", operation)
}
