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
- **IAM**: the deploying identity needs `roles/logging.configWriter` or broader.

## Deploy

### Console

Open the deployment store, find **GCP Log Metric**, and click **Deploy**. The creation wizard walks you through the target project, the filter, the metric descriptor, and the extraction and histogram settings. Start from the **Error Counter** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpLogMetric
metadata:
  name: checkout-errors
  org: acme-corp
  env: prod
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

The metric then charts (and alerts) as `logging.googleapis.com/user/checkout-errors`. An Infra Job tracks the provisioning in real time.

### InfraChart

When scoping the metric to a log bucket deployed in the same InfraPipeline, wire the reference with ValueFromRef:

```yaml
spec:
  filter: severity>=ERROR
  bucketName:
    valueFrom:
      kind: GcpLogBucket
      name: audit-logs
      fieldPath: status.outputs.bucket_name
```

The InfraPipeline deploys the bucket first, then the metric against it. The same chart typically carries a GcpMonitoringAlertPolicy whose threshold condition filters on `logging.googleapis.com/user/{metric}` — deploy both together and log spikes page on-call.

## Key Configuration

These are the most important decisions when configuring a log-based metric. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**filter** -- the Cloud Logging query selecting the entries that feed the metric. Scope it tightly (resource type + service) — an over-broad filter counts the whole project.

**metricDescriptor** -- the metric's shape: `DELTA`/`INT64` for counters (the workhorse), `DISTRIBUTION` for extracted-value histograms. Declare `labels` for every key `labelExtractors` populates. Label keys and value types REPLACE the metric on change — the schema is append-only in the API.

**valueExtractor + bucketOptions** -- the DISTRIBUTION surface: `EXTRACT(field)` or `REGEXP_EXTRACT(field, regex)` pulls the number; explicit/exponential/linear layouts histogram it (exponential is the usual choice for latencies).

**bucketName** -- scope the metric to a specific GcpLogBucket instead of the project's `_Default` (same-project only).

**disabled** -- pause ingestion without losing the configuration or history — the safe way to silence a misconfigured extractor while fixing it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpLogBucket** (optional) | `bucketName` | `status.outputs.bucket_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `metric_name` | The metric name | Alert filters and dashboard widgets as `logging.googleapis.com/user/{metric_name}` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Error counter** -- count error entries per service; the input to error-rate alerts. Start from the **Error Counter** preset.

**Latency distribution** -- extract request latency from access logs into a histogram; percentile charts from logs alone. Start from the **Latency Distribution** preset.

## Works With

- [**GCP Monitoring Alert Policy**](/infra-catalog/gcp-monitoring-alert-policy) -- turns log patterns into pages
- [**GCP Monitoring Dashboard**](/infra-catalog/gcp-monitoring-dashboard) -- charts the metric
- [**GCP Monitoring SLO**](/infra-catalog/gcp-monitoring-slo) -- log-based counters feed good/total SLIs
- [**GCP Log Bucket**](/infra-catalog/gcp-log-bucket) -- scopes the metric to a specific bucket
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project whose logs feed the metric
