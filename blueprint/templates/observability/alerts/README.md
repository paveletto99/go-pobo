# Alert Skeletons

This folder contains reusable SRE alert skeletons extracted from the source Terraform alerting intent.

The files are split into:

- `backend-neutral-alerts.yaml`: product-neutral alert definitions and SLI contracts.
- `prometheus/`: Prometheus rule examples for OTel metrics after Prometheus-style name conversion.
- `grafana/`: Grafana dashboard examples for the extracted API and scheduled-job SLIs.

These templates are examples, not a replacement for platform-specific tuning. Keep the SRE semantics stable:

- scheduled jobs emit success counters only after durable completion,
- forward-progress alerts check both missing success and absent telemetry,
- API services alert on availability, error ratio, latency, and traffic,
- telemetry pipeline alerts watch the Collector/backend exporter path,
- labels stay bounded and operationally useful.

Do not copy domain labels like `config_id`, `region`, or `includes_travelers` unless the new service truly has those bounded dimensions.
