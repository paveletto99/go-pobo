# Observability

This blueprint keeps the original repository structure but upgrades the starter from OpenCensus-style HTTP wrapping and Prometheus sidecar metrics toward OpenTelemetry.

Official OpenTelemetry sources used:

- Go signal status: https://opentelemetry.io/docs/languages/go/
- Go OTLP exporters: https://opentelemetry.io/docs/languages/go/exporters/
- Collector rationale: https://opentelemetry.io/docs/collector/
- Collector config: https://opentelemetry.io/docs/collector/configuration/
- Collector config best practices: https://opentelemetry.io/docs/security/config-best-practices/
- Logs specification: https://opentelemetry.io/docs/specs/otel/logs/
- Semantic conventions: https://opentelemetry.io/docs/specs/otel/semantic-conventions/

## Decision

Recommended low-overhead default:

```text
Go service
  -> OTLP/HTTP traces directly to local/gateway OpenTelemetry Collector
  -> OTLP/HTTP metrics directly to local/gateway OpenTelemetry Collector
  -> JSON stdout logs enriched with trace_id/span_id
  -> platform or Collector log pipeline forwards stdout logs when required
```

Why:

- Official Go docs list traces and metrics as stable, while Go logs are beta.
- Official Go exporter docs support OTLP traces and metrics over HTTP.
- Official Collector docs recommend using a Collector alongside the service so the app can offload quickly and the Collector can handle retries, batching, encryption, and filtering.
- Official logs guidance supports direct-to-Collector logs, but that path adds app-side network queuing/exporting. For lower application resource use, keep zap JSON stdout as the default and add trace/span correlation fields.

Direct OTLP logs remain an optional improvement when requirements demand fully in-process log export. Do not make direct log export the default unless the extra CPU/memory/queue cost is acceptable.

## Starter Implementation

The starter now includes:

- `pkg/observability.Config`
- `pkg/observability.New(ctx, cfg)`
- OTLP/HTTP trace exporter
- OTLP/HTTP metric exporter
- `BatchSpanProcessor` through `sdktrace.WithBatcher`
- `PeriodicReader` for metrics with `OTEL_METRIC_EXPORT_INTERVAL`
- parent-based trace-id-ratio sampling with `OTEL_TRACE_SAMPLE_RATE`
- W3C TraceContext and baggage propagators
- HTTP instrumentation through `otelhttp.NewHandler`
- `/health` excluded from HTTP telemetry
- zap logs enriched with `trace_id` and `span_id` by request middleware
- a minimal OpenTelemetry Collector deployment under `deploy/kubernetes/otel-collector`

## Configuration Contract

Application env vars:

| Variable | Default | Purpose |
| --- | --- | --- |
| `OTEL_ENABLED` | `false` | Enables OTel SDK setup. |
| `OTEL_SERVICE_NAME` | `sample-service` | Sets `service.name`. |
| `OTEL_SERVICE_VERSION` | `dev` | Sets `service.version`. |
| `OTEL_DEPLOYMENT_ENVIRONMENT` | empty | Sets `deployment.environment.name`. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTel exporter default | OTLP collector endpoint, for example `http://otel-collector:4318`. |
| `OTEL_EXPORTER_OTLP_TIMEOUT` | OTel exporter default | Export timeout, example `5s`. |
| `OTEL_TRACES_ENABLED` | `true` | Enables trace provider/exporter when OTel is enabled. |
| `OTEL_TRACE_SAMPLE_RATE` | `0.01` | Parent-based probability sampling ratio. |
| `OTEL_METRICS_ENABLED` | `true` | Enables metric provider/exporter when OTel is enabled. |
| `OTEL_METRIC_EXPORT_INTERVAL` | `60s` | Periodic metric export interval. |

The Kubernetes sample sets `OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318`.

## Metrics Upgrade

The original repo pattern used OpenCensus metrics and an optional Prometheus endpoint. The starter upgrades the template to OpenTelemetry metrics exported over OTLP/HTTP.

Preserve:

- metrics stay part of shared runtime setup,
- HTTP server instrumentation lives in `pkg/server`,
- service-specific custom metrics should live near the service package.

Upgrade:

- prefer OTLP metrics to a Collector over a per-process Prometheus scrape endpoint for new services,
- keep export interval at `60s` unless you have a clear SLO reason to reduce it,
- use low-cardinality attributes only,
- keep `/health` out of HTTP metrics to avoid noisy probe traffic.

Optional:

- If your backend is Prometheus-first, expose Prometheus from the Collector instead of every Go process. This keeps app processes smaller and centralizes scrape/export behavior.

## Custom Metrics Pattern

The sample service now includes service-owned custom metrics in `internal/sample/metrics.go`.

Observed/source-aligned structure:

- service-specific metrics live beside the service package, not in global runtime setup,
- shared runtime setup owns exporter/provider lifecycle,
- handlers and use-case code record business events using the request `context.Context`,
- the HTTP server wrapper still provides generic request telemetry through `otelhttp`.

Starter instruments:

| Instrument | Type | Attributes | Purpose |
| --- | --- | --- | --- |
| `sample.item.requests` | counter | `operation` | Counts create/get item requests. |
| `sample.item.errors` | counter | `operation`, `reason` | Counts bounded error categories such as `decode`, `invalid`, `internal`. |
| `sample.item.created` | counter | none | Counts successful item creation. |
| `sample.item.lookup` | counter | `result` | Counts `found` vs `not_found` lookups. |
| `sample.item.handler.duration` | histogram | `operation` | Measures handler duration in seconds. |

Guidance for new services:

- create a small `internal/<service>/metrics.go` with package-private instruments and helper methods,
- instantiate metrics in `NewServer` after config validation and before returning the server,
- record in handlers for request lifecycle outcomes and in services for durable business events,
- use counters for monotonic events and histograms for durations or sizes,
- pass `ctx` from the incoming request to every metric call,
- keep attributes bounded and business-useful.

Do not use high-cardinality labels:

- IDs,
- names,
- emails,
- tokens,
- request bodies,
- error strings,
- free-form paths,
- database query text.

Use stable categories instead: `operation=create`, `reason=invalid`, `result=not_found`, `source=database`.

## Tracing Upgrade

The starter uses `otelhttp.NewHandler` around the mux, matching the original repository's shared `pkg/server.ServeHTTPHandler` composition point.

Preserve:

- request context propagation,
- middleware logger enrichment,
- graceful shutdown.

Upgrade:

- use W3C TraceContext and baggage,
- sample in-process with parent-based ratio sampling,
- use batch exporting, not simple synchronous exporting,
- send spans to the Collector over OTLP/HTTP.

## Logs

Default:

- keep zap JSON logs on stderr/stdout,
- add `trace_id` and `span_id` in request middleware when a valid span exists,
- let platform logging or a Collector log pipeline collect container logs.

Why not direct OTLP logs by default:

- Go logs are currently beta in official OTel Go status.
- Official logs spec supports direct-to-Collector logs, but that means application-side network export and buffering.
- For low-resource services, stdout JSON plus trace correlation gives useful correlation with less app-side machinery.

Optional direct log export:

- add an OTel zap bridge and OTLP log exporter only for services that require app-originated OTLP logs,
- keep it behind `OTEL_LOGS_ENABLED`,
- batch logs and set bounded queues,
- benchmark before enabling for high-throughput paths.

## Collector

The included Collector config is intentionally small:

- OTLP/HTTP receiver on `4318`,
- `memory_limiter`,
- `batch`,
- `debug` exporter for local development.

For production, replace `debug` with your backend exporter and configure auth/TLS. Keep `memory_limiter` and `batch`; official Collector best-practice docs call out batching and memory limits as protections against OOM and usage spikes.

## Pattern Disposition

| Pattern | Disposition |
| --- | --- |
| Shared `pkg/observability` setup called from `internal/setup` | Preserve with OTel upgrade |
| `pkg/server` as HTTP instrumentation point | Preserve as-is |
| OpenCensus `ochttp` wrapper | Avoid copying into new services |
| Per-process Prometheus metrics endpoint | Optional, prefer Collector-managed export |
| OTLP/HTTP traces and metrics to Collector | Preserve as new default |
| Feature-local custom metrics in `internal/<service>/metrics.go` | Preserve with low-cardinality attributes |
| Direct OTLP logs from Go process | Optional improvement, disabled by default |
| JSON logs with `trace_id` and `span_id` | Preserve as low-overhead default |
