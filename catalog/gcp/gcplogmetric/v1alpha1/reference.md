# GcpLogMetric

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpLogMetricSpec defines a Cloud Logging log-based metric — the bridge
from logs to monitoring: entries matching `filter` are counted (or a
value is extracted from them) into a Cloud Monitoring metric that
dashboards chart and alert policies watch.

Two forms, chosen by the metric_descriptor:
  - COUNTER (metric_kind DELTA, value_type INT64): count matching
    entries — "how many 5xx responses". The default and simplest form.
  - DISTRIBUTION (value_type DISTRIBUTION): extract a numeric value
    from each entry with value_extractor and histogram it with
    bucket_options — "request latency from the logs".

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpLogMetric
metadata:
  name: my-sample-log-metric
spec:
  # The logging filter selecting the entries that feed the metric
  # (https://cloud.google.com/logging/docs/view/logging-query-language).
  # Lead with the cheapest selective term (resource.type) and scope to
  # the service — an over-broad filter counts the whole project.
  filter: resource.type="cloud_run_revision" AND severity>=ERROR

  # The metric's shape: DELTA/INT64 is the counter form (the workhorse).
  # DISTRIBUTION adds valueExtractor + bucketOptions for histograms.
  metricDescriptor:
    metricKind: DELTA
    valueType: INT64
    unit: "1"
    labels:
      - key: status
        description: HTTP status code class
        valueType: STRING

  # Populate the declared labels from each matching entry. Keep label
  # values BOUNDED (status classes, methods) — every distinct
  # combination is a billed time series.
  labelExtractors:
    status: EXTRACT(httpRequest.status)

  # What this metric measures (shown in the metrics explorer).
  description: Error entries per service — the input to error-rate alerts

  # What a destroy does: DELETE (default — alert policies watching the
  # metric silently stop evaluating), PREVENT (the posture once alerts
  # depend on it), or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.metricName` | `string` |  |  |  |
| `spec.filter` | `string` | yes |  |  |
| `spec.bucketName` | `string \| valueFrom` |  |  | GcpLogBucket (`status.outputs.bucket_name`) |
| `spec.description` | `string` |  |  |  |
| `spec.disabled` | `bool` |  |  |  |
| `spec.metricDescriptor` | `GcpLogMetricDescriptor` |  |  |  |
| `spec.metricDescriptor.metricKind` | `string` | yes |  |  |
| `spec.metricDescriptor.valueType` | `string` | yes |  |  |
| `spec.metricDescriptor.unit` | `string` |  |  |  |
| `spec.metricDescriptor.displayName` | `string` |  |  |  |
| `spec.metricDescriptor.labels` | `[]GcpLogMetricLabel` |  |  |  |
| `spec.metricDescriptor.labels[].key` | `string` | yes |  |  |
| `spec.metricDescriptor.labels[].description` | `string` |  |  |  |
| `spec.metricDescriptor.labels[].valueType` | `string` |  |  |  |
| `spec.valueExtractor` | `string` |  |  |  |
| `spec.labelExtractors` | `map<string, string>` |  |  |  |
| `spec.bucketOptions` | `GcpLogMetricBucketOptions` |  |  |  |
| `spec.bucketOptions.explicitBuckets` | `GcpLogMetricExplicitBuckets` |  |  |  |
| `spec.bucketOptions.explicitBuckets.bounds` | `[]double` | yes |  |  |
| `spec.bucketOptions.exponentialBuckets` | `GcpLogMetricExponentialBuckets` |  |  |  |
| `spec.bucketOptions.exponentialBuckets.numFiniteBuckets` | `int32` |  |  |  |
| `spec.bucketOptions.exponentialBuckets.growthFactor` | `double` |  |  |  |
| `spec.bucketOptions.exponentialBuckets.scale` | `double` |  |  |  |
| `spec.bucketOptions.linearBuckets` | `GcpLogMetricLinearBuckets` |  |  |  |
| `spec.bucketOptions.linearBuckets.numFiniteBuckets` | `int32` |  |  |  |
| `spec.bucketOptions.linearBuckets.offset` | `double` |  |  |  |
| `spec.bucketOptions.linearBuckets.width` | `double` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project whose logs feed the metric. Can be a literal project
ID or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.metricName

`string`

The metric name — how the metric is addressed from Monitoring as
logging.googleapis.com/user/{metric_name}. Defaults to metadata.name.
May include forward slashes for namespacing (e.g. "checkout/errors");
the API forbids characters that need URL-encoding.

### spec.filter

`string` · required

The logging filter selecting the entries that feed the metric
(https://cloud.google.com/logging/docs/view/logging-query-language),
e.g. resource.type="cloud_run_revision" AND severity>=ERROR

- rule: {"required":true}

### spec.bucketName

`string | valueFrom`

Scope the metric to a specific log BUCKET instead of the project's
_Default bucket: the metric then counts entries as they land in that
bucket. The full bucket resource name
(projects/{p}/locations/{l}/buckets/{b}) — a literal or a reference to
a GcpLogBucket resource (its bucket_name output is exactly this
value). The bucket must live in the same project as the metric.

- references: GcpLogBucket (`status.outputs.bucket_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpLogBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.description

`string`

What this metric measures and how it is meant to be used (shown in
the metrics explorer). At most 8000 characters (the provider's own
documented cap).

- rule: {"string":{"maxLen":"8000"}}

### spec.disabled

`bool`

Pause the metric: a disabled metric keeps its configuration and
history but ingests no new points — the safe way to silence a
misconfigured extractor while fixing it.

### spec.metricDescriptor

`GcpLogMetricDescriptor`

The shape of the metric this filter produces. Omit for the plain
count-of-entries form (GCP then applies its implicit DELTA/INT64
descriptor). Required whenever value_extractor, label_extractors, or
bucket_options are used.

### spec.metricDescriptor.metricKind

`string` · required

How points accumulate: DELTA (change since the last point — the form
log-based counters use), GAUGE (instantaneous value), or CUMULATIVE
(running total). Log-based metrics are DELTA except in rare GAUGE
cases. Changing the kind REPLACES the metric descriptor.

- rule: metric_kind must be one of: DELTA, GAUGE, CUMULATIVE
- rule: {"required":true}

### spec.metricDescriptor.valueType

`string` · required

The value type of each point: INT64 for counters, DISTRIBUTION for
extracted-value histograms; BOOL, DOUBLE, STRING, and MONEY complete
the API's set.

- rule: value_type must be one of: BOOL, INT64, DOUBLE, STRING, DISTRIBUTION, MONEY
- rule: {"required":true}

### spec.metricDescriptor.unit

`string`

The unit of the values (UCUM syntax, e.g. "ms", "By", "1"). Shown on
chart axes.

### spec.metricDescriptor.displayName

`string`

Display name shown in the metrics explorer. Defaults to
metadata.name when left empty.

### spec.metricDescriptor.labels

`[]GcpLogMetricLabel`

The labels each point carries (filled by label_extractors). Declare
every label the extractors populate.

### spec.metricDescriptor.labels[].key

`string` · required

The label key. Changing a label's key REPLACES the metric (label
schemas are append-only in the API).

- rule: {"required":true}

### spec.metricDescriptor.labels[].description

`string`

What this label carries.

### spec.metricDescriptor.labels[].valueType

`string`

The label's value type: BOOL, INT64, or STRING (empty means STRING —
the API default). Changing it REPLACES the metric.

- rule: label value_type must be one of: BOOL, INT64, STRING

### spec.valueExtractor

`string`

For DISTRIBUTION metrics: the expression extracting the numeric value
from each matching entry — EXTRACT(field) or
REGEXP_EXTRACT(field, regex), e.g.
  REGEXP_EXTRACT(jsonPayload.latency, "(\\d+)ms")

### spec.labelExtractors

`map<string, string>`

Populate metric labels from each matching entry: map of label name
(which must be declared in metric_descriptor.labels) to an EXTRACT /
REGEXP_EXTRACT expression, e.g.
  { "status": "EXTRACT(httpRequest.status)" }

### spec.bucketOptions

`GcpLogMetricBucketOptions`

For DISTRIBUTION metrics: the histogram bucket layout values are
recorded into. At least one layout (the provider's own rule when the
block is present); layouts may be combined.

- rule: set at least one of: explicit_buckets, exponential_buckets, or linear_buckets

### spec.bucketOptions.explicitBuckets

`GcpLogMetricExplicitBuckets`

Hand-picked bucket boundaries.

### spec.bucketOptions.explicitBuckets.bounds

`[]double` · required

The boundary values, ascending. N boundaries define N+1 buckets.

- rule: {"repeated":{"minItems":"1"}}

### spec.bucketOptions.exponentialBuckets

`GcpLogMetricExponentialBuckets`

Buckets that grow geometrically — the usual choice for latencies
(each bucket `growth_factor` times wider than the last).

### spec.bucketOptions.exponentialBuckets.numFiniteBuckets

`int32`

How many finite buckets (there is always an underflow and an overflow
bucket beyond them).

- rule: {"int32":{"gt":0}}

### spec.bucketOptions.exponentialBuckets.growthFactor

`double`

The growth factor between adjacent buckets (must exceed 1).

### spec.bucketOptions.exponentialBuckets.scale

`double`

The scale of the first finite bucket (must be positive).

### spec.bucketOptions.linearBuckets

`GcpLogMetricLinearBuckets`

Equal-width buckets.

### spec.bucketOptions.linearBuckets.numFiniteBuckets

`int32`

How many finite buckets.

- rule: {"int32":{"gt":0}}

### spec.bucketOptions.linearBuckets.offset

`double`

The starting value of the first finite bucket.

### spec.bucketOptions.linearBuckets.width

`double`

The width of each bucket (must be positive).

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the metric is deleted; its historical points expire
               with Monitoring retention
  "PREVENT" -- destroy FAILS; protects a metric that alert policies
               depend on from accidental teardown
  "ABANDON" -- the metric is removed from management but keeps
               ingesting in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpLogMetric, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.metric_name` | `string` | The metric name. Address it from Cloud Monitoring (dashboards, alert policy filters) as metric.type = "logging.googleapis.com/user/{metric_name}". |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.bucketName` | GcpLogBucket | `status.outputs.bucket_name` |

## See Also

- [Overview](../README.md)
