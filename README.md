# New Project Template Example

This folder is a copyable starter scaffold for a new Go service that follows the same high-level shape as this repository.

The files use a `.tmpl` suffix so they do not participate in the current workspace build. To use them:

1. Copy this folder into a new repository or service directory.
2. Rename `*.tmpl` files to their real names.
3. Replace `example.com/my-project` with your module path.
4. Fill in the TODOs for logging, observability, auth, rate limiting, database, and deployment details.

## Layout

```text
examples/new-project-template/
├── .env.example
├── README.md
├── builders/build.yaml.tmpl
├── builders/service.dockerfile.tmpl
├── go.mod.tmpl
├── cmd/adminapi/main.go.tmpl
├── cmd/server/main.go.tmpl
├── cmd/worker/main.go.tmpl
├── internal/envstest/harness.go.tmpl
├── internal/routes/adminapi.go.tmpl
├── internal/routes/server.go.tmpl
├── internal/routes/worker.go.tmpl
├── pkg/cache/cache.go.tmpl
├── pkg/config/adminapi_config.go.tmpl
├── pkg/config/server_config.go.tmpl
├── pkg/config/worker_config.go.tmpl
├── pkg/controller/admin/controller.go.tmpl
├── pkg/controller/example/controller.go.tmpl
├── pkg/controller/example/controller_test.go.tmpl
├── pkg/controller/middleware/request_id.go.tmpl
├── pkg/controller/worker/controller.go.tmpl
├── pkg/database/database.go.tmpl
├── pkg/render/renderer.go.tmpl
├── pkg/worker/example/worker.go.tmpl
├── scripts/build.tmpl
├── terraform/locals.tf.tmpl
├── terraform/main.tf.tmpl
├── terraform/service_server.tf.tmpl
├── terraform/variables.tf.tmpl
└── docs/development.md.tmpl
```

## What This Shows

- a thin `cmd/server` entrypoint with `realMain(ctx)`
- a thin `cmd/adminapi` entrypoint for JSON/admin surfaces
- a thin `cmd/worker` entrypoint for scheduled or internal jobs
- route assembly in `internal/routes`
- constructor-injected controllers in `pkg/controller`
- a sample request ID middleware in `pkg/controller/middleware`
- a reusable worker stub in `pkg/worker`
- a minimal `internal/envstest` harness and example test
- typed environment config in `pkg/config`
- extracted local build, Cloud Build, and Docker packaging seams
- explicit infrastructure stubs for database, cache, render, and Terraform

This is intentionally minimal. It shows the seams to start from scratch without pretending to be a finished service.

## Build And Package

This scaffold follows the same packaging split as this repo: build the Go binary into `bin/`, package it with a thin service Dockerfile, and optionally wrap remote builds through `scripts/build` plus `builders/build.yaml`.

Local binary build:

```sh
mkdir -p bin
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o ./bin/server ./cmd/server
```

Local Docker package:

```sh
docker build --file builders/service.dockerfile --build-arg SERVICE=server --tag my-project/server:dev .
```

Cloud Build wrapper:

```sh
PROJECT_ID=my-project-id SERVICE=server TAG=dev ./scripts/build
```
