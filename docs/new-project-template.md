# New Project Template

This template is a stripped-down starting point for a new Go service or sibling project that should feel native to this repository.

See `examples/new-project-template/` for a copyable scaffold. The files in that folder use a `.tmpl` suffix so they stay out of the current repo's build and test paths until you rename them in a new project.

The scaffold now includes three optional extra surfaces beyond the base web server: an admin API, an HTTP-triggered worker, and a small test harness patterned after `internal/envstest`.

It also includes a basic request ID middleware template so route setup starts with a real `pkg/controller/middleware` seam instead of comments only.

For shipping services, the scaffold also includes an extracted build pipeline: `builders/service.dockerfile`, `builders/build.yaml`, and `scripts/build`, modeled on this repo's binary-first Docker packaging flow.

## Suggested layout

```text
my-project/
├── cmd/
│   └── server/main.go
├── internal/
│   └── routes/server.go
├── pkg/
│   ├── config/server_config.go
│   ├── controller/example/controller.go
│   ├── database/
│   ├── cache/
│   └── render/
├── terraform/
│   ├── main.tf
│   ├── locals.tf
│   ├── variables.tf
│   └── service_server.tf
└── docs/
    └── development.md
```

## Startup shape

```go
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := logging.NewLoggerFromEnv()
	ctx = logging.WithLogger(ctx, logger)

	if err := realMain(ctx); err != nil {
		logger.Fatal(err)
	}
}

func realMain(ctx context.Context) error {
	cfg, err := config.NewServerConfig(ctx)
	if err != nil { return fmt.Errorf("load config: %w", err) }

	db, err := cfg.Database.Load(ctx)
	if err != nil { return fmt.Errorf("load db config: %w", err) }
	defer db.Close()

	mux, err := routes.Server(ctx, cfg, db)
	if err != nil { return fmt.Errorf("build routes: %w", err) }

	srv, err := server.New(cfg.Port)
	if err != nil { return fmt.Errorf("create server: %w", err) }
	return srv.ServeHTTPHandler(ctx, mux)
}
```

## Route and controller shape

```go
type Controller struct {
	db *database.Database
	h  *render.Renderer
}

func New(db *database.Database, h *render.Renderer) *Controller {
	return &Controller{db: db, h: h}
}

func (c *Controller) HandleShow() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controller.RenderJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}
```

Register controllers in `internal/routes/server.go`, not in `main.go`. Put shared middleware near the router root, then add auth or rate limiting only on the subrouter that needs it.

## Terraform shape

Use the same layering as this repo:

- shared infrastructure in `main.tf`, `locals.tf`, and domain files such as database, cache, and keys
- one service definition per `service_*.tf`
- shared environment maps in locals
- `_all` overrides merged before service-specific overrides

```hcl
dynamic "env" {
  for_each = merge(
    local.database_config,
    local.cache_config,
    lookup(var.service_environment, "_all", {}),
    lookup(var.service_environment, "server", {}),
  )

  content {
    name  = env.key
    value = env.value
  }
}
```

## First-project checklist

1. Add typed config with validation.
2. Keep startup wiring in `cmd/` only.
3. Inject dependencies into controllers.
4. Use context-aware middleware for request-scoped data.
5. Add a minimal Terraform service with a dedicated service account and explicit secret access.
6. Add admin API, worker, and envstest harness files only when the project actually needs those surfaces.
7. Add the Docker build files only when the project will ship a container image.
