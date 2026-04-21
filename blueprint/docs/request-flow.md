# Request Flow

## Publish API Flow

Representative path: `POST /v1/publish`.

```text
cmd/exposure/main.go
  -> setup.Setup(ctx, &publish.Config)
  -> publish.NewServer(ctx, &cfg, env)
  -> server.New(cfg.Port)
  -> ServeHTTPHandler(ctx, publishServer.Routes(ctx))
  -> mux route /v1/publish
  -> publish.Server.handlePublishV1
  -> handleRequest: jsonutil.Unmarshal into pkg/api/v1.Publish
  -> process:
       authorizedapp.Provider.AppConfig
       verification.Verifier.VerifyDiagnosisCertificate
       revision.TokenManager.UnmarshalRevisionToken
       publish/model.Transformer.TransformPublish
       publish/database.PublishDB.InsertAndReviseExposures
       revision.TokenManager.MakeRevisionToken
       metrics/logging response
  -> jsonutil.MarshalResponse
```

Important details:

- The handler maps JSON parsing errors to response codes before business logic.
- `process` maps known domain/repository errors with `errors.Is` and `errors.As`.
- `InsertAndReviseExposures` owns the database transaction and performs read/merge/insert/update in one `InTx`.
- Context from `r.Context()` reaches verification, token manager, transformer, database, logging, tracing, and metrics.

Disposition: **preserve with light cleanup**. This is the richest example, but new services should avoid putting this much logic into one method unless the use case is truly cohesive.

## Scheduled Job Flow

Representative path: `POST/GET /rotate-keys` in `keyrotation`.

```text
cmd/key-rotation/main.go
  -> setup.Setup
  -> keyrotation.NewServer
  -> Routes
  -> handleRotateKeys
       ctx := r.Context()
       db.Lock(ctx, "key-rotation-lock", ttl)
       doRotate(ctx)
       render.RenderJSON
  -> revision/database.RevisionDB
  -> KMS through serverenv key manager
```

Important details:

- Lock contention maps to HTTP 200 with `"too early"` because retrying a scheduler is not useful.
- The lock is released with a deferred unlock and logged on failure.
- The use-case method `doRotate` is testable without HTTP once dependencies are in the `Server`.

Disposition: **preserve as-is** for scheduler-triggered jobs.

## Export Worker Flow

Representative path: `/do-work`.

```text
export.Server.handleDoWork
  -> context.WithTimeout(r.Context(), cfg.WorkerTimeout)
  -> export/database.ExportDB.LeaseBatch
  -> processBatch / exportBatch
  -> publish/database.PublishDB.IterateExposures
  -> blobstore writes
  -> export/database.ExportDB.FinalizeBatch
```

Important details:

- Work is leased in the database before external blob writes.
- Finalization writes export-file records and marks the batch complete in a DB transaction.
- External blob operations use their own timeouts.

Disposition: **preserve** for multi-step jobs with leases.

## Federation gRPC Flow

Representative path: `Federation.Fetch`.

```text
cmd/federationout/main.go
  -> setup.Setup
  -> federationout.NewServer(env, &cfg)
  -> grpc.NewServer(options)
  -> optional AuthInterceptor
  -> federation.RegisterFederationServer
  -> server.ServeGRPC
  -> Server.Fetch(ctx, req)
       context.WithTimeout(ctx, cfg.Timeout)
       fetch(ctx, req, publishdb.IterateExposures, fetchUntil)
       map failures to gRPC errors or hidden internal errors
```

Important details:

- gRPC auth is an interceptor, not route middleware.
- `Fetch` creates a timeout context.
- The service accepts a function parameter for iteration in internal helper code, making core logic testable.

Disposition: **preserve** for gRPC services. Function injection is a useful testability pattern.

## Repository Transaction Flow

Representative method: `PublishDB.InsertAndReviseExposures`.

```text
PublishDB.InsertAndReviseExposures(ctx, req)
  -> validate request
  -> db.db.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error {
       encode incoming keys
       ReadExposures(ctx, tx, keys)
       validate revision token and metadata
       model.ReviseKeys(ctx, existing, incoming)
       prepare insert/update statements
       execute inserts/revisions
       return nil or sentinel/wrapped error
     })
  -> return response counters
```

Important details:

- Transaction boundary is in the database adapter.
- Domain merge logic is delegated to `model.ReviseKeys`.
- Known domain errors are returned as sentinels/types and mapped by the handler/use-case layer.

Disposition: **preserve as-is**.

## Template Request Flow

The starter implements this distilled flow:

```text
cmd/sample-service
  -> setup.Setup
  -> sample.NewServer
  -> Routes
  -> handleCreateItem / handleGetItem
  -> sample.Service
  -> sample/database.ItemDB
  -> pkg/database.DB.InTx
```

The starter adds narrow private interfaces around service/repository dependencies for tests. This is an **optional improvement** rather than a universal repo pattern.
