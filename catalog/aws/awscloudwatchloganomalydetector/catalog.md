# AWS CloudWatch Logs Anomaly Detector

Deploys a CloudWatch Logs anomaly detector — a machine-learning model that studies a log group's normal patterns and flags what breaks them: a new exception class, a volume spike, a format change — without a single hand-written filter pattern. The model trains on live traffic and needs up to 24 hours to build a stable baseline, so a freshly deployed detector reporting nothing is training, not broken. Pausing with `enabled: false` preserves the trained model; deleting the detector discards it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CloudWatch Logs Anomaly Detector** — trained over the log groups in `logGroupArns`, evaluating on the `evaluationFrequency` cadence, surfacing anomalies for `anomalyVisibilityTime` days. Findings are encrypted with the customer-managed key in `kmsKeyId` when set.
- **AWS Tags** — resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, carrying the CloudWatch Logs anomaly-detection permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **The log group to train on** — reference an AwsCloudwatchLogGroup's `log_group_arn` output or pass a literal group ARN (the bare ARN, no `:*` suffix).
- **A KMS key** (only for encrypted findings) — a customer-managed key for `kmsKeyId`; reference an AwsKmsKey or pass a key ARN.

## Deploy

### Console

Open the deployment store, find **AWS CloudWatch Logs Anomaly Detector**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec fields — the log group, cadence, visibility window, and optional training filter. Start from the **Application Log Anomalies** or **Error-Focused Hourly Detector** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogAnomalyDetector
metadata:
  name: api-anomalies
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  detectorName: api-anomalies
  logGroupArns:
    - value: arn:aws:logs:us-east-1:123456789012:log-group:/app/api
  enabled: true
  evaluationFrequency: FIVE_MIN
  anomalyVisibilityTime: 30
```

```shell
planton apply -f log-anomaly-detector.yaml
```

This creates an active detector over the `/app/api` log group, evaluating every five minutes and keeping surfaced anomalies visible for 30 days. An Infra Job tracks the provisioning in real time.

### InfraChart

When the detector deploys alongside its log group in one chart, wire the group reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  logGroupArns:
    - valueFrom:
        kind: AwsCloudwatchLogGroup
        name: api-logs
        fieldPath: status.outputs.log_group_arn
  enabled: true
  evaluationFrequency: FIVE_MIN
```

The InfraPipeline resolves the dependency graph, deploys the log group first, then attaches the detector to it.

## Key Configuration

These are the most important decisions when configuring an anomaly detector. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The list that takes one entry** — AWS models `logGroupArns` as a list but currently rejects more than one entry; a second group fails at apply, not at validation. The spec stays list-shaped because that is AWS's own forward-compatible contract — give it exactly one log group until AWS lifts the cap.

**Pause, don't delete** — `enabled: false` stops evaluation but keeps the trained baseline; deleting the detector discards it, and retraining takes up to a day of live traffic. Pause through planned incident noise (load tests, migrations); delete only on decommission.

**Evaluation cadence is the cost-versus-lag trade** — `evaluationFrequency` runs from ONE_MIN to ONE_HOUR. Longer frequencies analyze less often, cost less, and surface anomalies later. FIVE_MIN suits interactive services; hourly suits batchy or noisy workloads.

**Train on everything, or only on what matters** — `filterPattern` limits training to events matching a CloudWatch Logs filter pattern. Unset trains on the full stream, which finds format drift and volume anomalies; a pattern like ERROR/WARN-only produces a calmer detector focused on failure classes.

**The KMS key is a one-way door** — changing `kmsKeyId` replaces the detector, because AWS cannot re-encrypt a trained model in place. A replacement means losing the baseline and retraining, so decide on customer-managed encryption before the model has history worth keeping.

**Anomaly visibility window** — `anomalyVisibilityTime` (7–90 days, AWS default 21) is how long a surfaced anomaly stays inspectable before aging out. Long windows suit slow-burn investigations; short windows keep the finding list current.

**AccessDenied looks like deletion** — the provider drops the detector from state when reads return AccessDenied, so an IAM regression shows up in plans as "detector vanished". Check permissions before concluding someone deleted it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCloudwatchLogGroup** | `logGroupArns[]` | `status.outputs.log_group_arn` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains one value: `anomaly_detector_arn`, the detector's identity and the provider's import ID. Nothing downstream composes on a detector via ValueFromRef — the output exists for auditing and import.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Default posture for a service's log group** — five-minute evaluation, 30-day visibility, no training filter, so the model learns the whole stream and catches format drift alongside error spikes. Let it baseline for a day before trusting the quiet. Start from the **Application Log Anomalies** preset.

**Error-focused hourly detector** — for noisy or batchy workloads: train only on ERROR/WARN lines via `filterPattern`, evaluate hourly, and keep anomalies visible the full 90 days for slow-burn investigations. Trades detection latency for cost and calm. Start from the **Error-Focused Hourly Detector** preset.

## Works With

- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — the log group the detector trains on, wired via `logGroupArns`
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — the customer-managed key encrypting the detector's findings
