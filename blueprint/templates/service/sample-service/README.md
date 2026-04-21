# sample-service

This starter is a runnable extraction of the repository's Go service style:

- thin `cmd/sample-service` entrypoint,
- envconfig-powered `Config`,
- `setup` plus `serverenv`,
- constructor-based `NewServer`,
- mux route registration,
- handler methods on `Server`,
- small use-case service,
- concrete repository adapter using `pkg/database.DB.InTx`,
- context logger,
- OpenTelemetry OTLP/HTTP traces and metrics,
- feature-local custom metrics in `internal/sample/metrics.go`,
- request logs enriched with trace/span ids,
- `/health`,
- Dockerfile and Kubernetes manifests.

Run tests:

```sh
go test ./...
```

Build:

```sh
docker build -t sample-service:dev .
```

Local OTel defaults:

- set `OTEL_ENABLED=true` to enable SDK setup,
- set `OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318`,
- keep `OTEL_TRACE_SAMPLE_RATE` low for high-throughput services,
- keep logs on stdout unless direct OTLP logs are explicitly required.

Custom metrics included:

- `sample.item.requests`
- `sample.item.errors`
- `sample.item.created`
- `sample.item.lookup`
- `sample.item.handler.duration`
