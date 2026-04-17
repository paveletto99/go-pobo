# Feature Development Wiki

This page captures the main implementation patterns already used in this repository so new features match the codebase instead of introducing a parallel style.

## Repository shape

- `cmd/` owns process bootstrap and shutdown only.
- `internal/routes/` composes routers, subrouters, and middleware stacks.
- `pkg/controller/` contains controllers and middleware.
- `pkg/config/` is the single place for environment-backed configuration.
- `pkg/database/`, `pkg/cache/`, `pkg/ratelimit/`, and `pkg/render/` own infrastructure adapters.
- `terraform/` keeps shared infrastructure in shared files and service deployment in `service_*.tf` files.

Keep dependencies flowing inward: entrypoints wire infrastructure, routes compose handlers, controllers coordinate work, and lower packages implement shared concerns.

## Server setup, handlers, and lifecycle

The canonical startup is visible in `cmd/server/main.go` and repeated in `cmd/apiserver/main.go` and `cmd/adminapi/main.go`:

1. Create a signal-aware root context with `signal.NotifyContext`.
2. Build the logger once and attach build metadata.
3. Load typed config from `pkg/config`.
4. Start observability before serving traffic.
5. Create cache, database, key managers, rate limiter, and auth providers once.
6. Inject those dependencies into the route builder.
7. Create the HTTP server and block on `ServeHTTPHandler`.
8. `defer` every cleanup in the same scope that allocated it.

Use `main()` as a thin shell around `realMain(ctx)`. Expensive objects belong to process startup, not to request handlers.

Controllers follow constructor injection. A controller struct stores only the dependencies it needs, and handler methods return `http.Handler`. That pattern is used across `pkg/controller/codes`, `pkg/controller/issueapi`, `pkg/controller/login`, and related packages.

## Go patterns and optimizations

The strongest Go pattern in this repo is strict boundary placement.

- Put executable wiring in `cmd/`.
- Keep non-exported routing and glue code in `internal/`.
- Put reusable infrastructure and transport helpers in `pkg/`.

Other patterns to preserve:

- Load env-backed config through typed structs and validation in `pkg/config`.
- Prefer small provider interfaces at package boundaries such as cache, email, and auth providers.
- Wrap errors with `%w` so callers keep context and error chaining.
- Thread `context.Context` through long-lived operations and request paths.
- Keep shared rendering and encoding in helpers instead of duplicating response code inside handlers.

The main explicit optimization pattern is in `pkg/render/renderer.go`: templates are cached, guarded by an `RWMutex`, and rendered through a `sync.Pool` of buffers to reduce allocations and prevent partial responses. Keep this style for any new high-frequency rendering or encoding path.

When adding code, avoid broad interfaces, hidden globals, and package cycles. If a new dependency is only needed by one controller, inject it there instead of adding it to a shared god package.

## Routers and middleware

Routing is built with Gorilla Mux and subrouters. The pattern is:

1. Create a base router.
2. Install global concerns that truly apply everywhere.
3. Fork subrouters by surface area.
4. Add the smallest middleware stack that enforces the feature's rules.

The UI server in `internal/routes/server.go` shows the fullest stack:

- static assets mounted before heavy request middleware
- request ID, trace ID, logger, recovery, and observability early
- template variables, locale, debug, session, and CSRF in the shared web stack
- auth, membership, MFA, firewall, and rate limit per protected route group

The API server in `internal/routes/apiserver.go` shows two critical rules:

- keep `/health` on a minimal stack
- put chaff middleware before rate limiting so padding traffic does not spend quota

If you add endpoints that need form-based `PUT` or `PATCH`, preserve the existing `MutateMethod` wrapper pattern at the outer edge.

## Terraform best practices

The Terraform layout is service-oriented but shares configuration through locals.

- `main.tf`, `database.tf`, `redis.tf`, `network.tf`, and `keys.tf` define shared infrastructure.
- `services.tf` aggregates environment maps such as cache, database, signing, and feature flags.
- Each `service_*.tf` file creates one service account, IAM bindings, Cloud Run service, network exposure, and outputs.

Patterns worth keeping:

- model secrets as Secret Manager references, not inline values
- keep KMS, database, cache, and observability wiring in shared locals
- merge `_all` overrides first and service-specific overrides last
- use explicit `depends_on` when service readiness depends on APIs, IAM, migrations, or secrets
- ignore deploy-managed Cloud Run fields in `lifecycle.ignore_changes`

The main improvement opportunity is repetition. If new infrastructure repeats the same Cloud Run and IAM shape, prefer a reusable service module instead of copying another `service_*.tf` file unchanged.

## Feature checklist

Use this sequence when adding a feature:

1. Pick the correct surface: UI server, API server, admin API, background job, or Terraform only.
2. Add or extend typed config first if the feature is environment-driven.
3. Create or extend the controller with constructor-injected dependencies.
4. Register the route in `internal/routes/*` and attach only the middleware the route needs.
5. Reuse renderer, bind, auth, rate-limit, and error helpers before inventing a new pattern.
6. Add tests close to the touched package and validate the smallest affected slice.
7. If the feature needs deployment changes, add shared locals first and then the per-service Terraform updates.

This repo scales best when new work follows the existing dependency direction and operational wiring rather than adding alternate startup, routing, or deployment conventions.
