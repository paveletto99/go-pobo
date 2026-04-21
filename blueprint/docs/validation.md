# Validation

This file is updated when the blueprint is changed.

Initial validation targets:

- `go list ./...` from the repository root.
- `go test -short ./...` from the repository root when feasible.
- `go test ./...` inside `blueprint/templates/service/sample-service`.
- YAML parsing for Kubernetes and Skaffold files when tooling is available.
- `skaffold render` when Skaffold is installed and a Kubernetes context is available.

Results are summarized in the final response for this extraction.

## 2026-04-20 Extraction Validation

Validated:

- `go list ./...` from repository root: passed. The nested starter module under `blueprint/templates/service/sample-service` did not pollute the root module package list.
- `go test -short ./...` from repository root: passed.
- `go list ./...` inside `blueprint/templates/service/sample-service`: passed.
- `go test ./...` inside `blueprint/templates/service/sample-service`: passed.
- YAML parsing for all `blueprint/**/*.yaml` with Ruby `YAML.load_stream`: passed.
- `kubectl kustomize blueprint/deploy/kubernetes/sample-service`: passed and rendered ConfigMap, Secret, Service, and Deployment.

Not validated locally:

- `skaffold render --filename blueprint/skaffold.yaml`: not run because `skaffold` is not installed in this environment (`command not found`).
- Live Kubernetes deployment: not attempted.

Remaining assumptions:

- The starter expects a reachable Postgres-compatible database at runtime because it preserves the repo's DB-backed `/health` style.
- The Kubernetes Secret and ConfigMap values are examples and must be replaced for real environments.
- The starter's private repository interface is a light testability improvement; the source repo mostly uses concrete repository adapters.

## 2026-04-21 OpenTelemetry Blueprint Update

Validated:

- `go test ./...` inside `blueprint/templates/service/sample-service`: passed.
- `go list ./...` from repository root: passed.
- `make yaml-check` under `blueprint`: passed.
- `kubectl kustomize blueprint/deploy/kubernetes/otel-collector`: passed.
- `kubectl kustomize blueprint/deploy/kubernetes/sample-service`: passed.

Not validated locally:

- `skaffold render --filename blueprint/skaffold.yaml`: not run because `skaffold` is still not installed in this environment.
- Collector runtime export to a real backend: the local Collector template uses the `debug` exporter intentionally.

Open assumptions:

- The starter now uses Go 1.20 to support the selected stable OpenTelemetry traces/metrics SDK/exporter set.
- The default log approach is stdout JSON with `trace_id` and `span_id`, not direct OTLP log export, to keep application-side resource use low.

## 2026-04-21 Custom Metrics Blueprint Update

Validated:

- `go test ./...` inside `blueprint/templates/service/sample-service`: passed.
- `go list ./...` inside `blueprint/templates/service/sample-service`: passed.
- `make yaml-check` under `blueprint`: passed. Ruby reported the existing local `ffi` extension warning, but YAML parsing completed successfully.

Open assumptions:

- Custom metrics stay feature-local and rely on the shared `pkg/observability` OTel provider initialized by `internal/setup`.
- Attribute sets remain intentionally low-cardinality; future services should add dimensions only when the value set is bounded.
