# Config And Runtime

## Configuration Loading

Observed:

- Config structs live in `internal/<service>/config.go`.
- Environment variables are declared with `env` tags from `github.com/sethvargo/go-envconfig`.
- Service configs embed shared configs like `database.Config`, `secrets.Config`, `keys.Config`, `storage.Config`, `observability.Config`.
- Compile-time assertions verify provider contracts, for example:

```go
var _ setup.DatabaseConfigProvider = (*Config)(nil)
```

- Config methods expose nested provider configs:

```go
func (c *Config) DatabaseConfig() *database.Config { return &c.Database }
```

Disposition: **preserve as-is**.

## Setup

`internal/setup.SetupWith` processes configuration in phases:

1. Configure secret manager first so secret references can be expanded.
2. Configure key manager.
3. Process the full config with envconfig.
4. Configure observability exporter.
5. Configure blobstore.
6. Configure database.
7. Configure authorized app provider after database.
8. Return `serverenv.New(ctx, opts...)`.

Disposition: **preserve as-is**. This is the repo's main dependency construction convention.

## Server Environment

`internal/serverenv.ServerEnv` is a dependency bag built with options:

- database,
- authorized app provider,
- blobstore,
- metrics exporter,
- key manager,
- secret manager,
- observability exporter.

Services validate required dependencies in `NewServer`, not in `main`.

Disposition: **preserve as-is**.

## Runtime Ports

Observed:

- Most HTTP services use `PORT` with default `8080`.
- Prometheus metrics use `METRICS_PORT` only when `OBSERVABILITY_EXPORTER=prometheus`.
- gRPC federationout also uses `PORT`, with optional TLS cert/key config.

Guidance: **preserve** `PORT=8080` and `/health`.

The blueprint starter adds OpenTelemetry runtime configuration while preserving the same setup/serverenv structure:

- `OTEL_ENABLED` enables OTel SDK setup.
- `OTEL_EXPORTER_OTLP_ENDPOINT` points at the Collector, for example `http://otel-collector:4318`.
- `OTEL_TRACE_SAMPLE_RATE` controls parent-based ratio sampling.
- `OTEL_METRIC_EXPORT_INTERVAL` controls periodic metric export.
- Logs remain zap JSON logs; request middleware adds `trace_id` and `span_id`.

## Shutdown

Observed:

- Main creates a signal context for SIGINT/SIGTERM.
- `pkg/server.ServeHTTP` watches the context and shuts down the HTTP server with a 5-second timeout.
- `ServeHTTPHandler` sets `ReadHeaderTimeout: 10 * time.Second`.
- gRPC uses `GracefulStop`.
- `env.Close(ctx)` closes database and observability exporters.

Disposition: **preserve as-is**.

## Secrets

Observed:

- Secret manager is selected by config and registered provider implementations.
- `secrets.Resolver` can mutate envconfig values from secret references.
- DB password is hidden from JSON logging with `json:"-"`.

Disposition: **preserve**. Do not log secrets or full configs with sensitive fields unless fields are explicitly redacted.

## Docker

Observed:

- Cloud Build compiles all `cmd/...` binaries into `./bin`.
- `builders/service.dockerfile` copies `./bin/${SERVICE}` into a minimal `scratch` image.
- It sets `USER nobody` and `ENV PORT 8080`.

Starter:

- Uses a multi-stage Dockerfile inside the sample service for local Skaffold friendliness.
- Keeps the same runtime principles: static binary, non-root user, `PORT=8080`, `/server` entrypoint.
- Uses Go 1.20 in the starter module because the stable OTel metrics SDK/exporter dependency set used by the template requires Go 1.20 or newer.

Disposition: **preserve with Skaffold-oriented packaging**.
