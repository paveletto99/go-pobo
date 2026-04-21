# Grafana Dashboard Example

`sample-service-sli-dashboard.json` is a lightweight Grafana dashboard example for the extracted SLI patterns.

It assumes a Prometheus datasource named `Prometheus` and Prometheus-style OTel metric names. Import it as a starting point, then tune thresholds, windows, and datasource names for the target environment.

Panels included:

- API request rate,
- API error ratio,
- handler p95 latency,
- item creation throughput,
- lookup result volume,
- scheduled-job forward progress,
- scheduled-job duration,
- Collector exporter failures,
- external probe success.
