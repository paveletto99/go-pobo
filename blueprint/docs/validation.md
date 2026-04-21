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

## 2026-04-21 SRE Alerting Extraction

Validated:

- Extracted alerting intent from `terraform/alerting` into `blueprint/docs/sre-alerting-sli-extraction.md`.
- Checked sample-service metrics against the extracted API SLI signals.

Open assumptions:

- The starter remains an HTTP API example, so scheduled-job forward-progress metrics are documented but not implemented in sample-service.
- Concrete alert rules depend on the target backend after the OTel Collector. The blueprint preserves the SRE semantics but does not force a specific Prometheus, Cloud Monitoring, or vendor query language.

## 2026-04-21 Alert Templates And Job Metrics

Validated:

- `go test ./...` inside `blueprint/templates/service/sample-service`: passed.
- `go list ./...` inside `blueprint/templates/service/sample-service`: passed.
- `go build ./cmd/sample-job ./cmd/sample-service` inside `blueprint/templates/service/sample-service`: passed.
- `go test ./cmd/sample-job ./internal/jobs/samplejob`: passed.
- `make yaml-check` under `blueprint`: passed. Ruby reported the existing local `ffi` extension warning, but YAML parsing completed successfully.
- Alert YAML parsed with Ruby YAML for `backend-neutral-alerts.yaml` and Prometheus `alert-rules.yaml`: passed.
- Grafana dashboard JSON parsed with Ruby JSON: passed.
- `kubectl kustomize blueprint/templates/service/sample-service/deploy/kubernetes/sample-job`: passed.

Open assumptions:

- Prometheus examples assume OTel metrics are exported with Prometheus-style names such as `sample_item_requests_total`.
- Grafana examples assume a Prometheus datasource named `Prometheus`.

Not validated locally:

- `promtool check rules`: not run because `promtool` is not installed in this environment.
- `skaffold render`: not run because `skaffold` is not installed in this environment.
