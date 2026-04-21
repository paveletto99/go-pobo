# Exposure Notifications Go Service Blueprint

This blueprint is an evidence-driven extraction of the Go service architecture in this repository. It is not a generic clean-architecture template. The codebase is a pragmatic multi-binary Go application with shared runtime packages, feature-owned server packages, and concrete database adapters.

## What Was Observed

- `cmd/<service>/main.go` packages are thin composition roots. They install signal handling, create a context logger, call `setup.Setup`, construct an `internal/<service>.Server`, create `pkg/server.Server`, and serve routes or gRPC.
- Most HTTP services live in `internal/<service>` with `config.go`, `server.go`, handler files, `metrics.go`, and sometimes `database/` and `model/` subpackages.
- The dominant layer shape is `main -> setup/serverenv -> internal/<service>.Server -> handler methods -> service/helper methods -> internal/<feature>/database adapters -> pkg/database.DB -> Postgres/external systems`.
- The repo does not consistently use separate controller/service/repository packages. Handlers and use-case orchestration often live on the same `Server` struct. Database adapters are concrete structs like `PublishDB`, `ExportDB`, `FederationInDB`, and `MirrorDB`.
- Interfaces are used sparingly. They are common for infrastructure capabilities (`storage.Blobstore`, `keys.KeyManager`, `secrets.SecretManager`, `authorizedapp.Provider`) and config-provider contracts in `internal/setup`, but not for every service or repository.
- Transactions are normally owned by repository/database adapters through `pkg/database.DB.InTx(ctx, pgx.ReadCommitted, func(tx pgx.Tx) error { ... })`.
- Context flows from OS signal context to setup/server startup, and from `r.Context()` through handlers, domain helpers, database calls, logging, tracing, and metrics.

## Contents

- `docs/go-architecture.md`: package layout, executable boundaries, dependency direction, and composition model.
- `docs/layer-analysis.md`: actual handler/service/repository hybrid model and what to preserve.
- `docs/package-map.md`: major packages, roles, dependencies, and copy/wrap/avoid guidance.
- `docs/request-flow.md`: representative HTTP, gRPC, job, and database request flows.
- `docs/service-template-guide.md`: practical guide for creating a new service in this style.
- `docs/observability.md`: logging, metrics, tracing, health, and runtime hooks.
- `docs/otel-upgrade.md`: OpenTelemetry upgrade notes and low-resource defaults.
- `docs/sre-alerting-sli-extraction.md`: Terraform alerting intent, reusable SLI/SLO signals, dashboard patterns, and sample-service metric coverage.
- `docs/config-and-runtime.md`: envconfig, setup, serverenv, ports, shutdown, secrets, and Docker behavior.
- `docs/testing-and-quality.md`: tests, fakes, database harness, error/validation conventions, and quality gates.
- `docs/runbook.md`: daily commands and local development flow.
- `templates/service/sample-service`: a runnable starter module that distills the repository style.
- `templates/observability/alerts`: backend-neutral alert skeletons plus Prometheus and Grafana examples for the extracted SLIs.
- `skaffold.yaml` and `deploy/kubernetes/sample-service`: a minimal Skaffold/Kubernetes wrapper for the starter.
- `scripts/scaffold-service.sh`: copies the starter into a new directory and renames the sample service.

## Naming the Architecture

The most accurate name is:

**Pragmatic server-centric, package-by-feature Go architecture with shared runtime infrastructure and concrete repository adapters.**

It is not strict Clean Architecture. It is not pure controller/service/repository. It is also not fully hexagonal, although provider interfaces for storage, KMS, secrets, and authorized apps are adapter-like.

## Pattern Disposition

Every observed pattern is classified in the docs as one of:

- **Preserve as-is**: strongly supported and worth copying.
- **Preserve with light cleanup**: supported, but the starter tightens sharp edges.
- **Avoid copying into new services**: present in the repo but risky or legacy-specific.
- **Optional improvement**: not universal in the repo, but consistent with its direction and useful for new services.

## Quick Start With The Starter

From the repo root:

```sh
cd blueprint
make test-template
make skaffold-render
```

To scaffold a service from the starter:

```sh
blueprint/scripts/scaffold-service.sh my-service /tmp/my-service
```

To run the included starter through Skaffold:

```sh
cd blueprint
skaffold dev
```

The starter is intentionally small, but it preserves the repo's important service shape: constructor-based composition, envconfig config loading, context-aware logging, mux routing, handler/use-case/database flow, `InTx` transaction ownership, health probes, Docker build, Kubernetes manifests, table-driven tests, and OpenTelemetry traces/metrics exported to an OTLP Collector.

It also includes an optional separate scheduled-job example in `templates/service/sample-service/cmd/sample-job` and `internal/jobs/samplejob`. Build it with `docker build --build-arg SERVICE=sample-job -t sample-job:dev templates/service/sample-service` when you need forward-progress job metrics.
