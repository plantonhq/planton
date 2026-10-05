# GCP Log Metric

Creates a Cloud Logging log-based metric — the bridge from logs to monitoring: entries matching a filter are counted (or a value is extracted from them) into a Cloud Monitoring metric that dashboards chart and alert policies watch. The pattern that turns "we log errors" into "we page on error rate".

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Log-based metric** -- a `logging.Metric` with the configured filter, descriptor, extractors, and histogram layout
- **Logging API enablement** -- `logging.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** whose logs feed the metric (directly or via a GcpProject reference).
- **Deployer permissions**: the least-privilege permission set the IaC runner's principal needs lives in [`iac/permissions.yaml`](iac/permissions.yaml).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpLogMetric
metadata:
  name: checkout-errors
spec:
  filter: resource.type="cloud_run_revision" AND resource.labels.service_name="checkout" AND severity>=ERROR
  metricDescriptor:
    metricKind: DELTA
    valueType: INT64
    unit: "1"
```

```shell
planton apply -f log-metric.yaml
```

The metric then charts (and alerts) as `logging.googleapis.com/user/checkout-errors`.

## The two forms

- **Counter** (DELTA/INT64) -- count matching entries: error counts, login attempts, job completions. The default and simplest form.
- **Distribution** (DISTRIBUTION + `valueExtractor` + `bucketOptions`) -- extract a number from each entry and histogram it: request latency from access logs, payload sizes, queue depths — percentile charts from logs alone.

## Outputs

| Output | Description |
|--------|-------------|
| `metric_name` | The metric name — address it from Monitoring as `logging.googleapis.com/user/{metric_name}` |

## Works With

- **GcpMonitoringAlertPolicy** -- threshold conditions on `logging.googleapis.com/user/{metric_name}` turn log patterns into pages
- **GcpMonitoringDashboard** -- chart the metric on dashboard widgets
- **GcpMonitoringSlo** -- log-based counters feed good/total ratio SLIs
- **GcpLogBucket** -- `bucketName` scopes the metric to a specific bucket instead of the project's `_Default`
- **GcpProject** -- provides the GCP project whose logs feed the metric

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
