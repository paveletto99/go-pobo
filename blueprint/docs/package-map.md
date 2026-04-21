# Package Map

This map classifies major packages by role and by what a new service should do with them.

## Top-Level

| Package/Path | Role | Inbound | Outbound | Guidance |
| --- | --- | --- | --- | --- |
| `cmd/<service>` | Executable boundary and composition root | Process manager, Cloud Run/Kubernetes | `internal/<service>`, `internal/setup`, `pkg/server`, `pkg/logging` | **Preserve as-is**. Keep thin. |
| `tools/*` | Local/admin utilities | Developers/operators | Internal packages and public APIs | **Avoid copying** into service starter unless you need an operational CLI. |
| `migrations/` | Shared Postgres schema | `cmd/migrate`, tests | SQL only | **Preserve** for DB-backed services. New service should own migrations if it owns tables. |
| `builders/*` | Cloud Build/Docker packaging | CI/CD | Docker, gcloud | **Preserve idea**, but starter uses Skaffold/Kubernetes instead of Cloud Build. |

## Shared Runtime

| Package | Purpose | Key Dependencies | Guidance |
| --- | --- | --- | --- |
| `internal/setup` | Loads env config, secrets, key manager, blobstore, DB, authorized app provider, observability | `go-envconfig`, `serverenv`, `database`, `keys`, `secrets`, `storage` | **Preserve as-is**. Config-provider interfaces are a core pattern. |
| `internal/serverenv` | Runtime dependency bag with option constructors and `Close` | shared providers | **Preserve as-is**. New services should request dependencies through config provider methods. |
| `pkg/server` | Listener creation, HTTP/gRPC graceful shutdown, observed source OpenCensus HTTP wrapper, health, Prometheus sidecar | `net/http`, `grpc`, `ochttp`, `database` | **Preserve structure; upgrade instrumentation to OTel in new services**. |
| `pkg/logging` | zap logger setup and context propagation | `zap` | **Preserve as-is**. Use context logger throughout. |
| `internal/middleware` | mux middlewares for recovery, request id, logger, observability, maintenance | `gorilla/mux`, `logging`, `observability` | **Preserve as-is** for HTTP services. |
| `pkg/render` | JSON response renderer for internal/job handlers | `encoding/json`, `multierror` | **Preserve with light cleanup**. Good for simple job endpoints. |
| `internal/jsonutil` | Strict JSON parsing and public API response marshaling | `encoding/json`, `net/http` | **Preserve as-is** for public JSON APIs. |
| `pkg/database` | pgx pool, config, `InTx`, locks, test DB harness | `pgx`, `dockertest`, `migrate` | **Preserve as-is**. |

## Feature Packages

| Package | Purpose | Architectural Role | Guidance |
| --- | --- | --- | --- |
| `internal/publish` | Public exposure publish API | Rich server/use-case package | **Study and preserve patterns**, but avoid copying all complexity. |
| `internal/publish/model` | TEK transformation and revision domain rules | Domain/model helper | **Preserve** for complex pure logic. |
| `internal/publish/database` | Exposure and stats persistence | Concrete repository adapter | **Preserve** transaction and sentinel-error style. |
| `internal/export` | Export batch job API | Job server plus worker logic | **Preserve job endpoint + lock pattern** for scheduled jobs. |
| `internal/export/database` | Export config, batch, file persistence | Concrete repository adapter | **Preserve** for lease/finalize transaction style. |
| `internal/federationin` | Pulls remote federation data | HTTP job handler plus pull orchestration | **Preserve with cleanup**. It uses function fields for testability. |
| `internal/federationout` | gRPC federation server | gRPC service plus auth interceptor | **Preserve** for gRPC starter variants. |
| `internal/jwks` | JWKS refresh job | Server + manager | **Preserve** manager extraction and bounded worker concurrency. |
| `internal/keyrotation` | Revision key rotation job | Server + DB/KMS orchestration | **Preserve** lock + renderer + concrete DB adapter style. |
| `internal/admin` | Admin console | Gin MVC-ish package | **Avoid copying** for API services; it is a special UI/admin path. |
| `internal/storage`, `pkg/keys`, `pkg/secrets` | Provider registries with interfaces and multiple implementations | Infrastructure adapters | **Preserve provider registry pattern** when a capability truly has multiple backends. |

## Dependency Risks

- Some feature packages import sibling feature packages. This is supported by the repo but can become coupling. Example: `generate` and `federationout` reuse `publish` database/model types.
- The `Server` struct can accumulate many collaborators. This is acceptable for small services but should be split into a manager/helper when orchestration grows.
- Interfaces are not systematically consumer-owned. Do not introduce broad interfaces unless they reduce real coupling or make tests materially simpler.
