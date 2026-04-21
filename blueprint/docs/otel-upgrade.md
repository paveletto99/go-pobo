# OpenTelemetry Upgrade Notes

## Scope

This file describes the blueprint-only OpenTelemetry upgrade. It does not change the source repository outside `/blueprint`.

## What Changed In The Starter

- Added `pkg/observability`.
- Added `ObservabilityConfigProvider` to `internal/setup`.
- Added observability provider lifecycle to `internal/serverenv`.
- Replaced OpenCensus HTTP wrapping with `otelhttp.NewHandler`.
- Removed the sample's per-process Prometheus metrics server.
- Added OTLP/HTTP trace and metric exporters.
- Added sampling and metric export interval configuration.
- Added trace/span IDs to request-scoped zap logs.
- Added a minimal OpenTelemetry Collector Kubernetes template.

## Resource-Conscious Defaults

- `OTEL_ENABLED=false` in code defaults; Kubernetes sample sets it to `true`.
- OTLP/HTTP is used rather than OTLP/gRPC to avoid adding gRPC exporter dependencies to the app path.
- Trace sampling defaults to `0.01`.
- Metric export interval defaults to `60s`.
- `/health` is filtered from HTTP telemetry.
- Logs stay as JSON stdout with trace/span correlation by default.
- Collector uses `memory_limiter` and `batch`.

## Why HTTP OTLP

The official Go exporter docs list OTLP traces and metrics over HTTP with binary protobuf payloads. HTTP keeps the application dependency graph smaller than gRPC exporters and is enough for a Kubernetes-local Collector service.

## Why Not Direct Logs By Default

OpenTelemetry supports direct-to-Collector logs, but the official logs spec frames it as network output from the application, typically through logging-library add-ons. That moves buffering and export work into the application. For this blueprint, preserving zap stdout logs with trace correlation is the lower-resource default.

## When To Enable Direct OTLP Logs

Enable direct OTLP logs only when:

- stdout log collection is unavailable or unacceptable,
- your backend requires OTLP log records directly from the app,
- you have measured the overhead,
- you can bound queues and failure behavior.

Otherwise, prefer container stdout logs enriched with `trace_id` and `span_id`.
