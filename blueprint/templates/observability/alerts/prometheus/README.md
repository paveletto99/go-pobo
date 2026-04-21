# Prometheus Alert Examples

These rules are examples for environments that export OTel metrics to Prometheus.

Assumptions:

- OTel metric names are converted to Prometheus-style names, for example `sample.item.requests` becomes `sample_item_requests_total`.
- OTel histogram names are converted to Prometheus histogram series, for example `sample.item.handler.duration` becomes `sample_item_handler_duration_bucket`.
- The scheduled-job example emits `sample.job.success` as `sample_job_success_total` with `job_name="sample-job"`.
- External synthetic checks use a `probe_success` metric, such as from a blackbox-style prober.

Tune windows and thresholds per service. The source Terraform pattern used job-specific windows derived from schedule frequency and tolerated missed runs.
