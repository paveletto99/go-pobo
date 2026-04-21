# Service Template Guide

## Create The Service

Use the script:

```sh
blueprint/scripts/scaffold-service.sh my-service ./my-service
```

This copies `templates/service/sample-service`, renames `sample-service`, and rewrites package/import placeholders.

## Folder Shape

Recommended shape:

```text
cmd/<service>/main.go
internal/<package>/config.go
internal/<package>/server.go
internal/<package>/handle_<action>.go
internal/<package>/service.go
internal/<package>/database/<thing>.go
internal/<package>/model/<thing>.go
pkg/database
pkg/logging
pkg/render
pkg/server
```

In this repository, `pkg` utilities are shared by many binaries. In a new standalone service, copy only the minimal shared runtime utilities you actually need.

## Define Config

Follow repo style:

- `Port string \`env:"PORT, default=8080"\``.
- Nested shared configs, such as `Database database.Config`.
- Provider methods consumed by setup, such as `DatabaseConfig() *database.Config`.
- Compile-time assertions for provider contracts.
- A `Validate() error` method when there are cross-field rules.

## Wire Dependencies

Follow this composition order:

1. `main` creates signal context and logger.
2. `main` calls `setup.Setup(ctx, &cfg)`.
3. `main` defers `env.Close(ctx)`.
4. `main` calls `<package>.NewServer(&cfg, env)`.
5. `NewServer` checks required dependencies.
6. `NewServer` constructs database adapters, services/managers, renderers, and stores them on `Server`.
7. `Routes(ctx)` registers middleware and handlers.

## Define Handlers

Use handler methods on `*Server`:

```go
func (s *Server) handleCreateThing() http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // parse, validate, call service, map error, render
  })
}
```

Do:

- use `r.Context()`,
- derive loggers from context,
- set per-request timeouts for long-running jobs,
- keep transport error mapping near the handler,
- use `jsonutil` for public JSON APIs or `render.RenderJSON` for simple internal/job APIs.

Do not:

- open DB transactions in handlers,
- log secrets or full request bodies,
- introduce generic controller packages.

## Define Service Logic

Use a small service/helper when logic is more than trivial. The repo often puts this on `Server`; the starter uses `Service` to make the shape easier to extend.

Use narrow private interfaces for repository dependencies only if tests need them:

```go
type itemRepository interface {
  InsertItem(context.Context, *model.Item) error
  GetItem(context.Context, string) (*model.Item, error)
}
```

This is an **optional improvement**, not a broad repo convention.

## Define Repository

Follow concrete adapter style:

```go
type ItemDB struct {
  db *database.DB
}

func New(db *database.DB) *ItemDB { return &ItemDB{db: db} }
```

Repository methods should:

- accept `context.Context` first,
- own `InTx` transaction boundaries,
- use pgx,
- map `pgx.ErrNoRows` to `pkg/database.ErrNotFound`,
- return sentinel errors for expected branches,
- wrap unexpected errors with `%w`.

## Register Routes

Use `gorilla/mux`:

```go
r := mux.NewRouter()
r.Use(middleware.Recovery())
r.Use(middleware.PopulateRequestID())
r.Use(middleware.PopulateLogger(logger))
r.Handle("/health", server.HandleHealthz(s.env.Database()))
r.Handle("/v1/items", s.handleCreateItem()).Methods(http.MethodPost)
```

## Add Tests

Minimum:

- `server_test.go` checks missing dependencies and route availability.
- Handler tests with `httptest`.
- Service tests with fakes.
- Repository integration tests when a real table/query is added.

## Add Skaffold/Kubernetes

For each service:

- add a Skaffold artifact pointing at the service directory and Dockerfile,
- add Deployment, Service, ConfigMap, Secret example,
- set `PORT=8080`,
- add `/health` readiness/liveness/startup probes,
- add `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`, sampling, and metric interval env vars,
- point local/dev services at the included `otel-collector` service,
- keep DB secrets separate from ConfigMap.

## Add Observability

Preserve the source repo's structure:

- keep observability setup in a shared package,
- expose it through config provider methods,
- attach lifecycle to `serverenv.Close`,
- instrument HTTP in `pkg/server`,
- enrich request logs in middleware.

For new services, default to:

- OTLP/HTTP traces and metrics to the Collector,
- parent-based ratio sampling,
- 60-second metric export interval,
- stdout JSON logs with `trace_id` and `span_id`,
- no direct OTLP logs unless explicitly required.

## Add Custom Metrics

Follow the sample service pattern:

- add `internal/<service>/metrics.go`,
- create instruments with `otel.Meter("example.com/<service>/internal/<service>")`,
- keep the metrics struct and helper methods package-private,
- construct metrics in `NewServer`,
- record request-level outcomes in handlers and durable business events in services,
- use request context for every metric call,
- use low-cardinality attributes only.

Copyable examples from `sample-service`:

- `sample.item.requests` counter for handler traffic by `operation`,
- `sample.item.errors` counter for stable error categories,
- `sample.item.created` counter for a successful business event,
- `sample.item.lookup` counter for found/not-found lookup outcomes,
- `sample.item.handler.duration` histogram for request duration.

Avoid:

- item ids, user ids, names, tokens, error strings, and request paths as attributes,
- process-local Prometheus exporters unless your platform requires them,
- custom metrics in `pkg/observability`; that package should stay focused on provider/exporter lifecycle.

## Do / Don't

Do:

- preserve `cmd -> setup -> serverenv -> NewServer -> Routes`,
- keep package ownership clear,
- keep transaction ownership in repositories,
- use concrete adapters by default,
- pass context everywhere,
- keep custom metrics feature-local and labels low-cardinality,
- make lock contention explicit for scheduled jobs.

Don't:

- force textbook clean architecture labels,
- create broad interfaces for every dependency,
- put SQL in handlers,
- hide runtime dependencies in globals,
- copy admin console patterns into API services,
- introduce Kubernetes complexity not required by the service.

## Patterns To Modernize Carefully

- Consider narrow private interfaces for unit-test seams.
- Consider a small use-case service when `Server` starts becoming too large.
- Avoid request-path goroutines unless you can explain cancellation and failure handling.
- Keep public error messages stable and less detailed than internal logs.
