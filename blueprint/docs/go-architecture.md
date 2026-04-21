# Go Architecture

## Directly Observed Style

This repository is a multi-binary Go service system. The executable boundary is `cmd/<binary>/main.go`; the service implementation lives primarily in `internal/<service>`; shared runtime and utility code lives in `pkg`; reusable API contracts live in `pkg/api`; generated protobufs live in `internal/pb`.

The architecture is best described as:

**Pragmatic server-centric, package-by-feature Go architecture with shared runtime infrastructure and concrete repository adapters.**

That name is deliberate. The repo has handlers, business orchestration, and repositories, but it does not consistently separate them into controller/service/repository packages. A service package usually owns a `Server` struct that holds config, environment dependencies, renderers, managers, database adapters, and domain helpers. Handler methods on that struct often parse requests and call private methods on the same struct for use-case logic.

## Entrypoints

Observed main packages include:

- HTTP APIs/jobs: `cmd/exposure`, `cmd/export`, `cmd/generate`, `cmd/jwks`, `cmd/key-rotation`, `cmd/mirror`, `cmd/federationin`, `cmd/backup`, `cmd/cleanup-*`, `cmd/export-importer`, `cmd/admin-console`, `cmd/metrics-registrar`.
- gRPC: `cmd/federationout`.
- Operational tools: `cmd/migrate` and many `tools/*`.

The common HTTP entrypoint pattern is:

1. `signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)`.
2. Build logger with `logging.NewLoggerFromEnv().With("build_id", ...).With("build_tag", ...)`.
3. Store logger in context with `logging.WithLogger`.
4. `setup.Setup(ctx, &cfg)` loads env config and shared dependencies.
5. `internal/<service>.NewServer(&cfg, env)` validates and wires service dependencies.
6. `server.New(cfg.Port)` creates a listener.
7. `srv.ServeHTTPHandler(ctx, service.Routes(ctx))` serves with shared HTTP instrumentation and graceful shutdown. In the observed source repo that wrapper is OpenCensus; in the blueprint starter it is upgraded to OpenTelemetry `otelhttp`.

`cmd/federationout` follows the same shape, but constructs a `grpc.Server`, optionally installs TLS and an auth interceptor, registers `federation.RegisterFederationServer`, and calls `srv.ServeGRPC`.

Disposition: **preserve as-is** for new services.

## Package Layout

Observed package clusters:

- `cmd/`: one thin binary per deployable process.
- `internal/<service>`: feature-owned server package. Typical files are `config.go`, `server.go`, handler files, `metrics.go`, service helper files, and tests.
- `internal/<service>/database`: concrete database adapter around `*pkg/database.DB`.
- `internal/<service>/model`: structs and domain logic for the service.
- `internal/setup`: envconfig-driven runtime setup. It recognizes provider interfaces implemented by service configs.
- `internal/serverenv`: shared dependency bag returned by setup.
- `internal/middleware`: mux middleware for recovery, request id, logger population, observability, maintenance, and chaff.
- `internal/jsonutil`: shared JSON request/response helpers for public JSON APIs.
- `pkg/database`: pgx pool wrapper, `InTx`, database config, lock helpers, test database harness.
- `pkg/server`: graceful HTTP/gRPC server wrapper and health handler. The observed source also has an optional Prometheus sidecar server; the blueprint starter upgrades metrics export to OTLP through OpenTelemetry.
- `pkg/logging`: zap sugar logger and context propagation.
- `pkg/render`: JSON renderer used by internal/job-like handlers.
- `pkg/observability`: observed source OpenCensus exporter setup and observability context helpers; blueprint starter OpenTelemetry provider setup.
- `pkg/api/v1`, `pkg/api/v1alpha1`: externally visible API DTOs.

Disposition: **preserve as-is**. New services should copy the shape, not necessarily every package.

## Dependency Direction

The practical direction is:

```text
cmd/<binary>
  -> internal/setup, pkg/logging, pkg/server
  -> internal/<service>
      -> internal/middleware, internal/serverenv
      -> internal/<service>/database, internal/<service>/model
      -> selected sibling packages when the use case requires it
      -> pkg/database, pkg/render, pkg/api, pkg/observability
```

Database adapters depend on `pkg/database` and their own `model` package. They do not import handlers. Models usually avoid importing database adapters. This avoids the most dangerous cycles.

The repo does allow pragmatic sibling dependencies. Examples include `internal/generate` importing `internal/publish/database` and `internal/publish/model`, and `internal/federationout` using `internal/publish/database` to iterate exposures. These are supported by the codebase, but they should be treated carefully.

Disposition: **preserve with light cleanup**. Copy the feature-owned package layout, but keep sibling package imports rare and explicit.

## Composition Model

`setup.Setup` is the shared runtime composition layer. It processes envconfig, resolves secrets, creates KMS/blob/database/observability providers if the service config implements the relevant provider interfaces, then returns a `serverenv.ServerEnv`.

Service constructors are the local composition root. `NewServer` checks required dependencies and builds service-specific collaborators. Examples:

- `publish.NewServer` requires database and authorized app provider, then creates a transformer, verifier, revision DB/token manager, chaff tracker, and `database.PublishDB`.
- `export.NewServer` requires blobstore, database, and key manager, then creates a renderer.
- `jwks.NewServer` requires database and creates a `Manager`.
- `keyrotation.NewServer` requires database and key manager, then creates `revisiondb.RevisionDB`.

Constructor signatures commonly use:

```go
func NewServer(cfg *Config, env *serverenv.ServerEnv) (*Server, error)
func NewServer(ctx context.Context, cfg *Config, env *serverenv.ServerEnv) (*Server, error)
```

Disposition: **preserve as-is**.

## Interface Placement

Observed interfaces are mostly capability/provider interfaces:

- `setup.DatabaseConfigProvider`, `setup.BlobstoreConfigProvider`, `setup.KeyManagerConfigProvider`, etc. These are owned by the setup package because setup consumes them.
- `storage.Blobstore`, `keys.KeyManager`, `secrets.SecretManager`, `authorizedapp.Provider`. These are provider-level interfaces because multiple implementations exist.
- Generated gRPC interfaces in `internal/pb/federation`.
- Very limited local controller abstraction, such as `internal/admin.Controller`.

Repository interfaces like `type ExposureRepository interface` are not a general repo convention. Database adapters are normally concrete structs with `New(db *database.DB)`.

Disposition:

- Provider/capability interfaces: **preserve as-is**.
- Broad repository/service interfaces for every package: **avoid copying into new services**.
- Narrow private interfaces near the consumer for unit tests: **optional improvement**. The starter uses this lightly because the user requested a repository interface, but it is marked as a cleanup rather than a repo-wide observed norm.

## Business Logic Location

Business logic is hybrid:

- Domain-heavy transformations live in model/service-like helpers, such as `internal/publish/model.Transformer` and `model.ReviseKeys`.
- Use-case orchestration often lives on `Server` methods, such as `publish.Server.process`, `keyrotation.Server.doRotate`, `generate.Server.generateKeysInRegion`, `mirror.Server.processMirror`, and `federationout.Server.fetch`.
- Database consistency rules and transaction sequencing live in concrete database adapters, such as `PublishDB.InsertAndReviseExposures`, `ExportDB.FinalizeBatch`, and `FederationInDB.StartFederationInSync`.

Disposition: **preserve with light cleanup**. Keep orchestration close to the server package, but move complex pure transformations into `model` or helper types.

## What To Copy For A New Service

Copy these conventions:

- `cmd/<service>/main.go` with signal context, logger, `setup.Setup`, `NewServer`, `pkg/server`.
- `internal/<service>/config.go` with env tags and provider methods.
- `internal/<service>/server.go` with `Server`, `NewServer`, `Routes`.
- Handler files named by action, returning `http.Handler`.
- `internal/<service>/database` concrete adapters, with `New(*database.DB)`.
- `internal/<service>/model` for request-independent domain structs/validation.
- Tests next to packages, using table-driven style, `httptest`, and fakes or function injection where useful.
- Docker build by compiling a binary and copying it into a small runtime image.
- Health probes on `/health`.

Avoid copying:

- Large all-knowing `Server` structs when a use case naturally splits into a manager/helper.
- Sibling-package imports just to reuse incidental code.
- Background goroutines that drop request cancellation unless intentionally decoupled.
- Error responses that leak internal details to clients.
