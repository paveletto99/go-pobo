# SRE Alerting, SLI, and Dashboard Extraction

This document extracts the reusable SRE intent from the source repository's `terraform/alerting` module and maps it to the blueprint starter. Generated output stays under `/blueprint`; the source Terraform remains evidence only.

## Evidence Inspected

- `terraform/alerting/alerts.tf`
- `terraform/alerting/probers.tf`
- `terraform/alerting/dashboards.tf`
- `terraform/alerting/dashboards/export-batches.json`
- `terraform/alerting/variables.tf`
- `terraform/alerting/notifications.tf`
- service metric definitions such as `internal/export/metrics.go`, `internal/backup/metrics.go`, `internal/cleanup/metrics.go`, `internal/exportimport/metrics.go`
- metric recording call sites such as export worker/batcher, backup, cleanup, jwks, key rotation, mirror, generate, and export importer

## Actual Intent

The alerting folder is not a complete generic SLO platform. It is a pragmatic Google Cloud Monitoring module built around four concerns:

1. Forward progress for scheduled/background jobs.
2. External availability probing for public HTTPS hosts.
3. Log-based operational/security alerts.
4. A narrow dashboard for export batch completion dimensions.

The strongest reusable pattern is forward-progress alerting. A job emits a success counter only after the durable work completes. Terraform creates two alert conditions per job:

- the success metric delta stays below `1` for a service-specific window,
- the success metric is absent for the same window.

That catches both "job is running and failing" and "job never ran or stopped exporting telemetry".

## Observed Alert Types

| Alert | Source behavior | Signal type | Notification class | Reuse guidance |
| --- | --- | --- | --- | --- |
| `ForwardProgress-*` | Checks custom job success counters under `custom.googleapis.com/opencensus/en-server/...` on `generic_task` resources | SLI: job completed at least once per window | Paging | Preserve for scheduled jobs. Convert metric names/export plumbing to OTel/Collector for new services. |
| `HostDown` | Uptime checks `/health` every 60s over HTTPS, alert when pass fraction by host drops below 20% for 60s | SLI: external availability | Paging | Preserve for externally reachable APIs. In Kubernetes, use probes for pod health and external synthetic checks for user-visible availability. |
| `StackdriverExportFailed` | Log-based metric for `jsonPayload.logger="stackdriver"` and message `failed to export metric`; alert if 5m rate is positive for 15m | Telemetry pipeline health | Non-paging | Preserve the concept. For OTel, move this to Collector/exporter error monitoring instead of app Stackdriver exporter logs. |
| `HumanAccessedSecret` | Audit-log metric for non-service-account Secret Manager access | Security control | Paging | Preserve if the platform uses GCP Secret Manager and DATA_READ audit logs. |
| `HumanDecryptedValue` | Audit-log metric for non-service-account Cloud KMS decrypt | Security control | Paging | Preserve if the platform uses Cloud KMS and DATA_READ audit logs. |
| `CloudRunBreakglass` | Audit-log metric for Cloud Run revisions deployed with breakglass | Deployment policy control | Paging | Cloud Run-specific. For Kubernetes, replace with admission/policy bypass alerts if available. |

## Forward Progress Indicators

Observed default job indicators:

| Job | Metric suffix | Window | Intended tolerance |
| --- | --- | --- | --- |
| `backup` | `backup/success` | 8h 10m | backup runs every 4h; page after about 2 missed successes |
| `cleanup-export` | `cleanup/export/success` | 8h 10m | cleanup runs every 4h; page after about 2 missed successes |
| `cleanup-exposure` | `cleanup/exposure/success` | 8h 10m | cleanup runs every 4h; page after about 2 missed successes |
| `export-batcher` | `export/batcher/success` | 18m | batcher runs every 5m; page after about 3 missed successes |
| `export-worker` | `export/worker/success` | 12m | worker runs every 1m but can run up to 5m; page after about 2 failures |
| `export-importer-schedule` | `export-importer/schedule/success` | 35m | scheduler runs every 15m; page after about 2 missed successes |
| `export-importer-import` | `export-importer/import/success` | 18m | importer runs every 5m; page after about 3 missed successes |
| `jwks` | `jwks/success` | 35m | jwks runs every 2m; page after roughly 15 missed successes |
| `key-rotation` | `key-rotation/success` | 8h 12m | key rotation runs every 4h; page after about 2 missed successes |
| `mirror` | `mirror/success` | 35m | mirror runs every 5m with a default 15m lock; page after about 2 missed successes |

Reusable rule:

```text
alert_window = expected_interval * tolerated_missed_runs + execution_slack + scheduling_slack
```

For each scheduled job, record a monotonic success counter only after the durable unit of work completes. Alert on both:

- no positive delta within `alert_window`,
- metric absence within `alert_window`.

## Dashboard Extraction

The observed dashboard is intentionally narrow:

- display name: `Export batches`
- metric: `custom.googleapis.com/opencensus/en-server/batch_completion`
- resource: `generic_task`
- alignment: `ALIGN_DELTA` over 60s, then `ALIGN_SUM`
- charts:
  - export batches by `config_id`
  - export batches by `region`
  - export batches by `includes_travelers`

Reusable dashboard pattern:

- chart durable business throughput, not only HTTP traffic,
- group by bounded business dimensions,
- use stacked bars for per-window completion counts,
- keep dashboard-specific labels low cardinality and operationally useful.

For a new service, copy the idea, not the export-domain labels. Examples:

- `operation` for API actions,
- `result` for `success`, `error`, `not_found`, `skipped`,
- `job` for scheduled job names,
- `region` only when the value set is controlled.

Avoid labels containing IDs, emails, tokens, full paths, request bodies, raw error strings, or unbounded resource names.

## Blueprint SLI Model

For API services:

| SLI | Preferred source | Required metrics |
| --- | --- | --- |
| Availability | External `/health` synthetic checks plus HTTP server telemetry | uptime check pass/fail, HTTP request count by status |
| Error ratio | HTTP server telemetry and service-owned error counters | request counter, error counter, bounded reason/result labels |
| Latency | HTTP server duration histogram and service-owned handler histogram | duration histogram by operation/route |
| Traffic | HTTP server request count and business event counters | request counter, durable event counters |
| Forward progress | Only for scheduled jobs or critical background loops | success counter plus absent-metric detection |

For scheduled/background services:

| SLI | Preferred source | Required metrics |
| --- | --- | --- |
| Completion | job success counter | `service.job.success` or `<service>.<job>.success` |
| Failure reason | job error counter | bounded `reason` label |
| Duration | job duration histogram | bounded `job` and optional `result` labels |
| Work volume | completed/skipped/failed work counters | bounded `result` and controlled dimensions |
| Staleness | timestamp/gauge or absent success counter | last success time or missing success metric |

## OTel Mapping For New Services

The source repo uses OpenCensus metric paths under `custom.googleapis.com/opencensus/en-server`. The blueprint starter uses OpenTelemetry exported to an OTLP Collector. Preserve the SRE semantics while modernizing names.

Recommended OTel instruments:

| Source repo pattern | OTel blueprint equivalent |
| --- | --- |
| `en-server/<job>/success` | `<service>.<job>.success` counter |
| `en-server/export/batch_completion` | `<service>.<domain>.completed` counter with bounded dimensions |
| OpenCensus views in service package | package-private OTel instruments in `internal/<service>/metrics.go` |
| Stackdriver metric exporter failure logs | Collector exporter error metrics/logs |
| GCP uptime check `/health` | Kubernetes probes plus external synthetic `/health` checks |

Use seconds for duration histograms to align with common OTel semantic convention units.

## Alert Templates To Reuse

Template assets:

- `blueprint/templates/observability/alerts/backend-neutral-alerts.yaml` provides backend-neutral alert skeletons.
- `blueprint/templates/observability/alerts/prometheus/alert-rules.yaml` provides Prometheus examples.
- `blueprint/templates/observability/alerts/grafana/sample-service-sli-dashboard.json` provides a Grafana dashboard example.

Forward progress:

```text
name: ForwardProgress-<job>
page: true
condition A: increase(<service>.<job>.success[window]) < 1
condition B: absent_over_time(<service>.<job>.success[window])
window: expected interval * tolerated misses + slack
runbook: job-specific playbook
```

API error ratio:

```text
name: HighErrorRatio-<service>
page: depends on user impact
condition: errors / requests > threshold for sustained window
attributes: route or operation, not raw path
```

API latency:

```text
name: HighLatency-<service>
page: usually ticket first unless severe
condition: p95 or p99 duration exceeds SLO threshold
attributes: route or operation
```

External availability:

```text
name: HostDown-<host>
page: true for public user-facing endpoints
condition: synthetic /health pass fraction below threshold
probe period: 60s
timeout: 10s
```

Telemetry pipeline:

```text
name: TelemetryExportFailed
page: false by default
condition: Collector exporter send failures sustained
action: inspect Collector, backend, auth, and network
```

## Sample Service Metric Check

Current sample-service metrics:

| Required SLI signal | Present? | Evidence in blueprint | Notes |
| --- | --- | --- | --- |
| HTTP request telemetry | Yes | `pkg/server` wraps the mux with `otelhttp.NewHandler` | Gives generic HTTP traces/metrics through OTel instrumentation. `/health` is filtered out to avoid probe noise. |
| Health endpoint for probes | Yes | `internal/sample/server.go` registers `/health` and Kubernetes manifests use it | Supports pod probes and external synthetic checks. |
| Request count by operation | Yes | `sample.item.requests` in `internal/sample/metrics.go` | Low-cardinality `operation=create|get`. |
| Error count by operation/reason | Yes | `sample.item.errors` in `internal/sample/metrics.go` | Reasons are bounded: `decode`, `invalid`, `internal`. |
| Handler latency | Yes | `sample.item.handler.duration` in `internal/sample/metrics.go` | Histogram records seconds by operation. |
| Durable business success | Partial | `sample.item.created` | Good for write throughput. It is traffic-dependent, so it should not be used like a scheduled-job forward-progress alert unless expected writes are guaranteed. |
| Lookup result volume | Yes | `sample.item.lookup` | Bounded `result=found|not_found`. |
| Scheduled-job forward progress | Yes, as optional separate example | `internal/jobs/samplejob` and `cmd/sample-job` | Not mixed into the HTTP sample. Emits `sample.job.success`, `sample.job.errors`, `sample.job.duration`, and `sample.job.work.items`. |
| Export-style dashboard dimensions | Not applicable | no export domain in sample-service | Do not copy `config_id`, `region`, or `includes_travelers` unless the new service actually has bounded equivalents. |
| Runtime/saturation metrics | Not yet | no Go runtime instrumentation in starter | Optional improvement: add runtime/process/DB pool metrics if capacity alerts are required. |

Conclusion: the sample service has the metrics needed for API SLI basics: availability endpoint, request count, error count, latency, and bounded business throughput. It also includes an optional separate scheduled-job example for forward-progress metrics, deliberately isolated from the HTTP service path.

## What To Preserve, Clean Up, Avoid, Improve

| Pattern | Disposition |
| --- | --- |
| Forward-progress success counters emitted after durable completion | Preserve as-is for jobs |
| Alert on both low success delta and absent metrics | Preserve as-is |
| Per-job windows based on schedule and tolerated missed runs | Preserve with light cleanup; document the formula beside each job |
| Uptime `/health` checks for externally reachable hosts | Preserve as-is |
| Export batch dashboard grouped by bounded business dimensions | Preserve the pattern, not the domain-specific labels |
| Security audit-log alerts | Preserve when the same GCP controls exist |
| OpenCensus/Stackdriver metric namespace | Avoid copying into new blueprint services; use OTel and Collector |
| Stackdriver app exporter failure alert | Optional improvement; replace with Collector exporter health for OTel |
| High-cardinality labels from logs/audit resources | Avoid in app metrics; acceptable only in narrowly scoped security log metrics |
| Runtime/process/DB saturation metrics | Optional improvement for capacity SLOs |
