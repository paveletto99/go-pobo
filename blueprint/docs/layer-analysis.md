# Layer Analysis

## Actual Layering Model

The repo does not use a textbook controller/service/repository architecture. The real model is a hybrid:

```text
Transport route
  -> Server handler method
  -> private Server use-case method or manager/helper
  -> concrete database adapter and/or external provider
  -> pkg/database.DB / blobstore / KMS / HTTP / gRPC
```

The closest accurate label is:

**Server-centric handler/use-case packages with concrete repository adapters.**

It is package-by-feature at the service level (`internal/publish`, `internal/export`, `internal/jwks`) and package-by-layer inside some features (`database`, `model`). Shared runtime infrastructure is package-by-capability (`pkg/server`, `pkg/database`, `internal/setup`).

## Handler / Controller Layer

### Observed

- HTTP handlers are methods on `*Server` and return `http.Handler`.
- Routes are registered in `Routes(ctx)` using `gorilla/mux` for most services.
- `internal/admin` is the exception: it uses `gin` and has a small `Controller` interface for admin pages.
- Middleware order commonly includes recovery, request ID, observability, logger, and sometimes maintenance/chaff.
- Request context is always taken from `r.Context()` and passed down.
- Job endpoints often parse query strings or no body and call a server method.
- Public JSON APIs use `internal/jsonutil.Unmarshal` with content-type check, `MaxBytesReader`, `DisallowUnknownFields`, and consistent parse errors.
- Simple job APIs often use `pkg/render.RenderJSON`.

### Preserve

- `Routes(ctx)` owns route registration and middleware.
- Handler functions should be thin enough to parse, set timeouts/locks, log, call one use-case method, map errors, and render.
- Use `r.Context()` as the root for downstream work.
- Use context logger from middleware, not package globals.
- Use `httptest` directly in handler tests.

### Preserve With Light Cleanup

- Split large handler/use-case methods when they grow like `publish.Server.process`. Keep orchestration in the service package, but move pure transformations into `model` or a helper.
- Keep response mapping near the transport boundary. For public APIs, avoid returning raw `err.Error()` for server faults.

### Avoid Copying

- Do not copy `gin` admin patterns into API services unless building an admin HTML console.
- Do not create generic `Controller` packages for APIs; the repo does not support that as a norm.

## Service / Use-Case Layer

### Observed

The service layer is mostly embodied by methods on `Server` or named helpers/managers:

- `publish.Server.process` performs authorization lookup, verification, transformations, revision token handling, database insert/revise, metrics, and response construction.
- `keyrotation.Server.doRotate` orchestrates revision key rotation.
- `generate.Server.generateKeysInRegion` creates synthetic exposure data and persists it.
- `jwks.Manager.UpdateAll` uses bounded concurrency for external JWKS calls and database updates.
- `federationout.Server.fetch` handles gRPC request validation, auth constraints, pagination/cursors, database iteration, and partial responses.

This is orchestration-heavy rather than a pure domain-service layer.

### Preserve

- Keep use-case logic in the owning feature package.
- Introduce a manager/helper when orchestration has its own lifecycle or concurrency model (`jwks.Manager` is a good example).
- Keep domain-heavy, request-independent rules in `model`.
- Pass `context.Context` to all IO and long-running work.

### Optional Improvement

- Use narrow private interfaces for dependencies when they simplify unit tests. This is consistent with `federationin` function fields, but it is not a repo-wide repository-interface convention.

### Avoid Copying

- Avoid "god services" that know too much about multiple sibling features.
- Avoid broad `Service` interfaces exported by provider packages unless multiple implementations are expected.

## Repository / Database Layer

### Observed

- Repository-like adapters are concrete structs: `PublishDB`, `ExportDB`, `FederationInDB`, `FederationOutDB`, `MirrorDB`, `AuthorizedAppDB`, `RevisionDB`, `HealthAuthorityDB`.
- Constructors are usually `func New(db *database.DB) *XDB`.
- Methods accept `context.Context` first.
- SQL is handwritten with pgx.
- Transactions are owned by database adapters through `db.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error { ... })`.
- Some methods accept a `pgx.Tx` for helper methods within a transaction, such as `UpdateStatsInTx` or `ReadExposures`.
- Sentinel errors live in repository or shared database packages, such as `pkg/database.ErrNotFound`, `pkg/database.ErrKeyConflict`, and publish-specific revision errors.
- `pgx.ErrNoRows` is mapped to `pkg/database.ErrNotFound`.
- Locking is implemented through `pkg/database.DB.Lock` and `MultiLock`.

### Preserve

- Concrete adapters around `*pkg/database.DB`.
- `context.Context` as first argument.
- Transaction closure ownership in the repository adapter.
- Map driver errors to package-level sentinel errors.
- Wrap errors with operation context using `%w`.
- Keep SQL close to the adapter that owns the table/model.

### Preserve With Light Cleanup

- Expose narrow interfaces at consumer boundary only when testing needs them.
- Keep external side effects outside DB transactions unless the transaction explicitly models a lease/finalize handoff.

### Avoid Copying

- Avoid database methods that spawn background goroutines unless clearly intentional. `PublishDB.InsertAndReviseExposures` updates stats in a goroutine with `context.Background()`; this avoids blocking the publish path but drops request cancellation and makes failure less visible.
- Avoid letting repositories return transport-specific errors or HTTP statuses.

## Error Mapping

### Observed

- Handlers translate domain/repository errors to HTTP status and response payloads.
- `publish` has explicit mapping for authorization, verification, revision token, transform, and database failures.
- Job endpoints often return HTTP 200 for "already locked/too early" so schedulers do not retry.
- gRPC auth failures use `status.Errorf(codes.Unauthenticated, ...)`; internal fetch failures are sometimes intentionally hidden as `"internal error"`.

### Preserve

- Map errors at transport boundary.
- Use `errors.Is` and `errors.As`.
- Treat lock contention as a non-failing scheduled-job result when retried schedulers would be harmful.

### Avoid Copying

- Do not expose internal database or verification error details to public clients unless the API contract requires it.

## Pattern Disposition Summary

| Pattern | Disposition |
| --- | --- |
| `cmd` thin main + `setup.Setup` + `NewServer` + `pkg/server` | Preserve as-is |
| Feature-owned `Server` with handler methods | Preserve as-is |
| Concrete `database.XDB` adapters | Preserve as-is |
| `InTx` closure per repository operation | Preserve as-is |
| Provider interfaces for real pluggable infrastructure | Preserve as-is |
| Broad service/repository interfaces everywhere | Avoid copying |
| Narrow private consumer interfaces for tests | Optional improvement |
| Large use-case method on `Server` | Preserve with light cleanup |
| Sibling feature imports | Preserve only when justified |
| Background goroutines in DB adapters | Avoid copying unless deliberately decoupled |
