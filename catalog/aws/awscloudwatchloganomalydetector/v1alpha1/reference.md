# AwsCloudwatchLogAnomalyDetector

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCloudwatchLogAnomalyDetectorSpec defines one CloudWatch Logs
anomaly detector: a machine-learning model that trains over a LIST
of log groups and surfaces unusual patterns (new error classes,
volume spikes, format drift) as anomalies.

One detector spans many log groups - it is multi-parent by design,
never a single group's satellite. Anomalies appear after the model's
initial training period (up to 24 hours on real traffic).

## Example

```yaml
# Canonical AwsCloudwatchLogAnomalyDetector example (hack/dev manifest
# and refgen Example source): a detector over one application log
# group with a five-minute cadence and a 30-day anomaly window.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogAnomalyDetector
metadata:
  name: api-anomalies
  id: api-anomalies
  org: test-org
  env: dev
spec:
  region: us-west-2
  detectorName: api-anomalies
  logGroupArns:
    - value: arn:aws:logs:us-west-2:123456789012:log-group:/app/api
  enabled: true
  evaluationFrequency: FIVE_MIN
  anomalyVisibilityTime: 30
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.detectorName` | `string` |  |  |  |
| `spec.logGroupArns` | `[]string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.enabled` | `bool` |  |  |  |
| `spec.evaluationFrequency` | `string` |  |  |  |
| `spec.filterPattern` | `string` |  |  |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.anomalyVisibilityTime` | `int64` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the detector runs in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.detectorName

`string`

The detector's display name in AWS. Unset lets AWS show the
detector unnamed (identity is the generated ARN either way).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.logGroupArns

`[]string | valueFrom` · required

The log groups the detector trains over. Reference
AwsCloudwatchLogGroup log_group_arn outputs or pass literal log
group ARNs (the bare group ARN, no ":*" suffix). NOTE: AWS's API
models this as a list but currently accepts exactly ONE entry -
additional entries fail at apply until AWS lifts its cap.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.enabled

`bool`

Whether the detector is actively analyzing. AWS requires an
explicit value; both engines always send it. Set false to pause
analysis without losing the trained model.

### spec.evaluationFrequency

`string`

How often the detector evaluates new log events. Unset keeps AWS's
default cadence. Longer frequencies cost less and lag more.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ONE_MIN","FIVE_MIN","TEN_MIN","FIFTEEN_MIN","THIRTY_MIN","ONE_HOUR"]}}

### spec.filterPattern

`string`

Limits training to events matching this CloudWatch Logs filter
pattern. Unset trains on everything.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key that encrypts the detector's findings.
Reference an AwsKmsKey key_arn output or pass a literal key ARN.
Changing it replaces the detector (AWS cannot re-encrypt a trained
model in place).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.anomalyVisibilityTime

`int64` · optional (explicit presence)

How many days a surfaced anomaly stays visible before aging out
(7-90). Unset keeps AWS's default (21). Presence-typed so the
boundary values are expressible.

- rule: {"int64":{"lte":"90","gte":"7"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudwatchLogAnomalyDetector, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.anomaly_detector_arn` | `string` | The detector's ARN (its identity and the provider's import ID). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.logGroupArns` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
