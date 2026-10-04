# AwsManagedPrometheus

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsManagedPrometheusSpec defines one Amazon Managed Prometheus (AMP)
workspace with its folded satellites: the workspace configuration
(retention and label-set series limits), workspace logging, the
alert manager definition, name-keyed rule group namespaces, query
logging, the workspace resource policy, and alias-keyed anomaly
detectors.

Scrapers are deliberately NOT here - an AMP scraper can target
CloudWatch with zero AMP workspaces, so it is its own kind
(AwsManagedPrometheusScraper) that references this workspace's ARN.

Two AWS contracts worth knowing before editing: the workspace ALIAS
can never be unset once set (AWS offers no un-alias - both engines
replace the workspace), and the workspace CONFIGURATION persists
after destroy (AWS has no delete API for it - the settings-retention
class; removing the block leaves the last-applied retention/limits
in place).

## Example

```yaml
# Canonical AwsManagedPrometheus example (hack/dev manifest and refgen
# Example source): a workspace with every satellite arm - alias,
# configuration (retention + a label-set series cap), workspace
# logging, the alert manager definition, a recording-rules namespace,
# query logging, the resource policy, and an anomaly detector. Literal
# values stand in for composed references so the offline `tofu plan`
# renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsManagedPrometheus
metadata:
  name: platform-metrics
  id: platform-metrics
  org: test-org
  env: dev
spec:
  region: us-west-2
  alias: platform-metrics
  logging:
    logGroupArn:
      value: arn:aws:logs:us-west-2:123456789012:log-group:/amp/workspace
  configuration:
    retentionPeriodInDays: 90
    limitsPerLabelSet:
      - labelSet:
          team: ingest
        maxSeries: 1000000
  alertManagerDefinition: |
    alertmanager_config: |
      route:
        receiver: default
      receivers:
        - name: default
          sns_configs:
            - topic_arn: arn:aws:sns:us-west-2:123456789012:amp-alerts
              sigv4:
                region: us-west-2
  ruleGroupNamespaces:
    - name: slo-rules
      data: |
        groups:
          - name: request-rates
            rules:
              - record: job:http_requests:rate5m
                expr: sum(rate(http_requests_total[5m])) by (job)
  queryLogging:
    destinations:
      - logGroupArn:
          value: arn:aws:logs:us-west-2:123456789012:log-group:/amp/queries
        qspThreshold: 10000
  # Statements carry no Resource member: the modules compose the
  # workspace's own ARN (AMP accepts no other Resource value). A single
  # action is authored as a string - AMP's stored form collapses
  # one-element Action arrays to the scalar.
  resourcePolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          AWS: arn:aws:iam::210987654321:root
        Action: aps:RemoteWrite
  anomalyDetectors:
    - alias: request-rate
      query: sum(rate(http_requests_total[5m]))
      evaluationIntervalInSeconds: 60
      missingDataAction: SKIP
      ignoreNearExpectedFromAbove:
        ratio: 0.1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.alias` | `string` |  |  |  |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.logging` | `AwsManagedPrometheusLogging` |  |  |  |
| `spec.logging.logGroupArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.configuration` | `AwsManagedPrometheusWorkspaceConfiguration` |  |  |  |
| `spec.configuration.retentionPeriodInDays` | `int32` |  |  |  |
| `spec.configuration.outOfOrderTimeWindowInSeconds` | `int32` |  |  |  |
| `spec.configuration.ruleQueryOffsetInSeconds` | `int32` |  |  |  |
| `spec.configuration.limitsPerLabelSet` | `[]AwsManagedPrometheusLabelSetLimit` |  |  |  |
| `spec.configuration.limitsPerLabelSet[].labelSet` | `map<string, string>` |  |  |  |
| `spec.configuration.limitsPerLabelSet[].maxSeries` | `int64` |  |  |  |
| `spec.alertManagerDefinition` | `string` |  |  |  |
| `spec.ruleGroupNamespaces` | `[]AwsManagedPrometheusRuleGroupNamespace` |  |  |  |
| `spec.ruleGroupNamespaces[].name` | `string` | yes |  |  |
| `spec.ruleGroupNamespaces[].data` | `string` | yes |  |  |
| `spec.queryLogging` | `AwsManagedPrometheusQueryLogging` |  |  |  |
| `spec.queryLogging.destinations` | `[]AwsManagedPrometheusQueryLoggingDestination` | yes |  |  |
| `spec.queryLogging.destinations[].logGroupArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.queryLogging.destinations[].qspThreshold` | `int64` |  |  |  |
| `spec.resourcePolicy` | `object` |  |  |  |
| `spec.anomalyDetectors` | `[]AwsManagedPrometheusAnomalyDetector` |  |  |  |
| `spec.anomalyDetectors[].alias` | `string` | yes |  |  |
| `spec.anomalyDetectors[].query` | `string` | yes |  |  |
| `spec.anomalyDetectors[].evaluationIntervalInSeconds` | `int32` |  |  |  |
| `spec.anomalyDetectors[].labels` | `map<string, string>` |  |  |  |
| `spec.anomalyDetectors[].sampleSize` | `int32` |  |  |  |
| `spec.anomalyDetectors[].shingleSize` | `int32` |  |  |  |
| `spec.anomalyDetectors[].ignoreNearExpectedFromAbove` | `AwsManagedPrometheusIgnoreNearExpected` |  |  |  |
| `spec.anomalyDetectors[].ignoreNearExpectedFromAbove.amount` | `double` |  |  |  |
| `spec.anomalyDetectors[].ignoreNearExpectedFromAbove.ratio` | `double` |  |  |  |
| `spec.anomalyDetectors[].ignoreNearExpectedFromBelow` | `AwsManagedPrometheusIgnoreNearExpected` |  |  |  |
| `spec.anomalyDetectors[].ignoreNearExpectedFromBelow.amount` | `double` |  |  |  |
| `spec.anomalyDetectors[].ignoreNearExpectedFromBelow.ratio` | `double` |  |  |  |
| `spec.anomalyDetectors[].missingDataAction` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the workspace lives in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.alias

`string`

The workspace's display alias. AWS CONTRACT: once set, an alias
can be changed but never unset - clearing this field replaces the
workspace. Leave it unset only if you never want one.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key encrypting the workspace's metric data.
Reference an AwsKmsKey key_arn output or pass a literal ARN.
Changing it replaces the workspace.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.logging

`AwsManagedPrometheusLogging`

Workspace event logging to a CloudWatch log group (workspace
lifecycle events, rule evaluation failures).

### spec.logging.logGroupArn

`string | valueFrom` · required

The receiving log group. Reference an AwsCloudwatchLogGroup
log_group_arn output or pass a literal group ARN. AWS requires
the ":*" wildcard suffix on this ARN; the log group resource
exports the bare ARN - both engines append ":*" when absent, so
wire the natural output and never hand-craft the suffix.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.configuration

`AwsManagedPrometheusWorkspaceConfiguration`

The workspace configuration: retention and per-label-set series
limits. AWS CONTRACT: this object is created via update and has NO
delete - removing the block leaves the last-applied values in
place (destroy is a no-op at AWS).

- rule: each limits_per_label_set entry needs at least one label in label_set

### spec.configuration.retentionPeriodInDays

`int32` · optional (explicit presence)

How many days ingested metrics are kept (at least 1). Unset keeps
AWS's default (150 days).

- rule: {"int32":{"gte":1}}

### spec.configuration.outOfOrderTimeWindowInSeconds

`int32` · optional (explicit presence)

How far out-of-order samples may arrive and still be ingested, in
seconds (0-600). Unset keeps AWS's default.

- rule: {"int32":{"lte":600,"gte":0}}

### spec.configuration.ruleQueryOffsetInSeconds

`int32` · optional (explicit presence)

How far behind NOW rule queries evaluate, in seconds (0-86400) -
headroom for slow-arriving samples. Unset keeps AWS's default.

- rule: {"int32":{"lte":86400,"gte":0}}

### spec.configuration.limitsPerLabelSet

`[]AwsManagedPrometheusLabelSetLimit`

Active-series limits per label set (e.g. cap series where
{team="ingest"}). Each entry pairs a label set with its
max_series.

### spec.configuration.limitsPerLabelSet[].labelSet

`map<string, string>`

The labels identifying the series population (all must match).

### spec.configuration.limitsPerLabelSet[].maxSeries

`int64`

The active-series cap for that population (0 blocks ingestion for
the set entirely).

- rule: {"int64":{"gte":"0"}}

### spec.alertManagerDefinition

`string`

The workspace's alert manager definition - the alertmanager.yml
document (SNS is the one supported receiver). Strictly one per
workspace.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.ruleGroupNamespaces

`[]AwsManagedPrometheusRuleGroupNamespace`

Recording/alerting rule files, one AWS rule-groups namespace per
entry, keyed by name.

### spec.ruleGroupNamespaces[].name

`string` · required

The namespace's name (its identity within the workspace -
renaming replaces it). Also the key for the outputs map.

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.ruleGroupNamespaces[].data

`string` · required

The rule-groups document - standard Prometheus rules YAML
("groups:" with recording/alerting rules), exactly what you would
put in a rules file on self-managed Prometheus.

- rule: {"string":{"minLen":"1"}}

### spec.queryLogging

`AwsManagedPrometheusQueryLogging`

Query logging: which queries (above a QSP threshold) are logged
and where.

### spec.queryLogging.destinations

`[]AwsManagedPrometheusQueryLoggingDestination` · required

Where logged queries land. AWS accepts multiple destinations.

- rule: {"repeated":{"minItems":"1"}}

### spec.queryLogging.destinations[].logGroupArn

`string | valueFrom` · required

The receiving log group. Reference an AwsCloudwatchLogGroup
log_group_arn output or pass a literal group ARN (both engines
append the ":*" suffix AWS requires when absent).

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.queryLogging.destinations[].qspThreshold

`int64`

Only queries whose Query Samples Processed meet this threshold
are logged (0 logs everything).

- rule: {"int64":{"gte":"0"}}

### spec.resourcePolicy

`object`

The workspace's resource policy - the IAM document granting other
principals/accounts access to this workspace. Write statements
WITHOUT a Resource member: AMP requires every statement's Resource
to be exactly this workspace's own ARN (PutResourcePolicy rejects
anything else with "Resource in policy does not match workspace
resource ARN"), so both modules compose it after the workspace
exists - an authored Resource is replaced, never honored. Author
a single action as a STRING ("Action": "aps:RemoteWrite"), not a
one-element list - AMP's stored form collapses single-element
Action arrays to the scalar, and an adopted (imported) policy
authored as a list diffs forever against that echo. The provider's
revision_id concurrency token is deliberately not modeled (a
state-managed apply-behavior knob, meaningless as declarative
config).

### spec.anomalyDetectors

`[]AwsManagedPrometheusAnomalyDetector`

Metric anomaly detectors, keyed by alias. Random Cut Forest is
AWS's only detection algorithm today; each entry's PromQL query
selects what the model watches.

### spec.anomalyDetectors[].alias

`string` · required

The detector's alias (its display identity and the outputs-map
key).

- rule: {"string":{"minLen":"1","maxLen":"255"}}

### spec.anomalyDetectors[].query

`string` · required

The PromQL query selecting the series the model watches (e.g.
"sum(rate(http_requests_total[5m]))").

- rule: {"string":{"minLen":"1"}}

### spec.anomalyDetectors[].evaluationIntervalInSeconds

`int32` · optional (explicit presence)

How often the detector evaluates, in seconds. Unset keeps AWS's
default.

- rule: {"int32":{"gte":1}}

### spec.anomalyDetectors[].labels

`map<string, string>`

Labels attached to the detector's anomaly results.

### spec.anomalyDetectors[].sampleSize

`int32` · optional (explicit presence)

Random Cut Forest sample size (at least 256). Unset keeps AWS's
default. RCF is AWS's only detection algorithm today; these knobs
tune it.

- rule: {"int32":{"gte":256}}

### spec.anomalyDetectors[].shingleSize

`int32` · optional (explicit presence)

Random Cut Forest shingle size (at least 2). Unset keeps AWS's
default.

- rule: {"int32":{"gte":2}}

### spec.anomalyDetectors[].ignoreNearExpectedFromAbove

`AwsManagedPrometheusIgnoreNearExpected`

Suppress anomalies within a band ABOVE the expected value.

- rule: set exactly one of amount and ratio

### spec.anomalyDetectors[].ignoreNearExpectedFromAbove.amount

`double` · optional (explicit presence)

The band as an absolute value (>= 0).

- rule: {"double":{"gte":0}}

### spec.anomalyDetectors[].ignoreNearExpectedFromAbove.ratio

`double` · optional (explicit presence)

The band as a ratio of the expected value (>= 0, e.g. 0.1 for
10%).

- rule: {"double":{"gte":0}}

### spec.anomalyDetectors[].ignoreNearExpectedFromBelow

`AwsManagedPrometheusIgnoreNearExpected`

Suppress anomalies within a band BELOW the expected value.

- rule: set exactly one of amount and ratio

### spec.anomalyDetectors[].ignoreNearExpectedFromBelow.amount

`double` · optional (explicit presence)

The band as an absolute value (>= 0).

- rule: {"double":{"gte":0}}

### spec.anomalyDetectors[].ignoreNearExpectedFromBelow.ratio

`double` · optional (explicit presence)

The band as a ratio of the expected value (>= 0, e.g. 0.1 for
10%).

- rule: {"double":{"gte":0}}

### spec.anomalyDetectors[].missingDataAction

`string`

What the model does when samples are missing: MARK_AS_ANOMALY
flags the gap, SKIP evaluates nothing.

- rule: {"string":{"in":["MARK_AS_ANOMALY","SKIP"]}}

## Validation Rules

- `spec.rule_namespace_names_unique`: rule_group_namespaces entries must have unique names
- `spec.anomaly_detector_aliases_unique`: anomaly_detectors entries must have unique aliases

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsManagedPrometheus, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.workspace_id` | `string` | The workspace's ID (the provider's import ID for the workspace and most satellites). |
| `status.outputs.workspace_arn` | `string` | The workspace's ARN - what scrapers and remote-write clients reference. |
| `status.outputs.prometheus_endpoint` | `string` | The workspace's Prometheus-compatible query/remote-write endpoint URL. |
| `status.outputs.rule_group_namespace_arns` | `map<string, string>` | Rule-groups namespace ARNs keyed by namespace name (each namespace imports by its ARN). |
| `status.outputs.anomaly_detector_ids` | `map<string, string>` | AWS-generated anomaly detector IDs keyed by detector alias (each detector imports as "detector_id,workspace_id"). |
| `status.outputs.anomaly_detector_arns` | `map<string, string>` | Anomaly detector ARNs keyed by detector alias. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.logging.logGroupArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |
| `spec.queryLogging.destinations[].logGroupArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsManagedPrometheusScraper | `spec.ampWorkspaceArn` | `status.outputs.workspace_arn` |

## See Also

- [Overview](../README.md)
