# AWS Config Recorder

Manages a region's AWS Config recording posture -- what changed, when, and what it looked like before: the configuration recorder, its S3 delivery channel, the recorder's running state, and the configuration-item retention window, as one resource. This is a region singleton: AWS allows exactly one recorder and one delivery channel per region, both named `default` by AWS convention, so deploy at most one instance per region. Recording bills per configuration item, which makes the recording group's scope the cost lever of the whole component.

## What Gets Created

This component owns the region's Config recording singletons -- AWS permits one of each per region and fixes their names (`metadata.name` never reaches AWS):

- **Configuration Recorder** -- the region's one recorder (AWS-conventional name `default`), with the service role, recording group (all / inclusion / exclusion), and recording mode (continuous or daily, with per-type overrides)
- **Delivery Channel** -- created only when `deliveryChannel` is set (required whenever the recorder runs); the region's one channel delivering history and snapshots to the S3 bucket, optionally KMS-encrypted and SNS-notified
- **Recorder Status** -- the folded start/stop state from `recordingEnabled` (unset = running)
- **Retention Configuration** -- managed only when `retentionPeriodInDays` is set; how long recorded configuration items stay queryable (30-2557 days)

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with AWS Config permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An IAM role trusting `config.amazonaws.com` for `roleArn` -- the AWS-managed `AWS_ConfigRole` policy is the canonical grant, plus write access to the history bucket.
- An S3 bucket for the delivery channel carrying a bucket POLICY that grants the Config service principal `s3:PutObject` and `s3:GetBucketAcl` -- AWS rejects the channel without it ("insufficient delivery policy"). The bucket policy is the bucket's contract, not this module's.
- No recorder already present in the region -- a console-enabled or other-tool recorder collides with this singleton (see Key Configuration).

## Deploy

### Console

Open the deployment store, find **AWS Config Recorder**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: the service role, recording scope and mode, delivery channel, and retention. Start from the **Scoped Recording** preset in the [Presets](#presets) tab for the cost-deliberate posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsConfigRecorder
metadata:
  name: config-recording
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: config-recorder-role
      fieldPath: status.outputs.role_arn
  recordingGroup:
    allSupported: false
    resourceTypes:
      - AWS::S3::Bucket
      - AWS::EC2::SecurityGroup
      - AWS::IAM::Role
    recordingStrategy: INCLUSION_BY_RESOURCE_TYPES
  deliveryChannel:
    s3BucketName:
      valueFrom:
        kind: AwsS3Bucket
        name: config-history
        fieldPath: status.outputs.bucket_id
    snapshotDeliveryFrequency: TwentyFour_Hours
  retentionPeriodInDays: 365
```

```shell
planton apply -f aws-config-recorder.yaml
```

This starts continuous recording of exactly three resource types, delivering history and daily snapshots to the referenced bucket with one year of queryable retention. An Infra Job tracks the provisioning in real time.

### InfraChart

When the recorder deploys alongside its role and bucket in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: config-recorder-role
      fieldPath: status.outputs.role_arn
  deliveryChannel:
    s3BucketName:
      valueFrom:
        kind: AwsS3Bucket
        name: config-history
        fieldPath: status.outputs.bucket_id
```

The InfraPipeline resolves the dependency graph, creates the role and bucket first, then the recorder and channel on top of them.

## Key Configuration

These are the most important decisions when configuring a Config recorder. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope is the bill** -- AWS Config charges per recorded configuration item, and `allSupported` in a busy account is the expensive default. Start from an inclusion list (`resourceTypes` with the `INCLUSION_BY_RESOURCE_TYPES` strategy) of exactly the types your Config rules evaluate; the exclusion strategy is the middle ground when the list of what you don't want is shorter. The spec's validation mirrors the provider's shape rules (inclusion and exclusion never mix, strategies must agree with `allSupported`), so an invalid combination fails in the manifest, before any apply.

**Record global types in exactly one region** -- `includeGlobalResourceTypes` pulls IAM users, roles, and policies into the recording. Enable it in one designated region only; every additional region multiplies those items and their cost while adding no information.

**Daily recording tames noisy types** -- the recording-mode override records chatty types as daily snapshots while everything else stays continuous. EC2 instances and ENIs during autoscaling churn are the classic bill amplifier; one override block (AWS accepts at most one) covers them.

**The singleton collides** -- a recorder already present in the region (console-enabled, or another tool's) fails creation. Adopt it by import or remove it -- there is no second name to hide behind, since AWS fixes the name `default`.

**Stopping is not losing** -- `recordingEnabled: false` keeps the recorder, channel, and all recorded history; recording and its bill pause until re-enabled. The flip side: a recorder stopped manually in the console drifts against the spec, and the next apply restarts it.

**Teardown order is encoded** -- AWS refuses to delete a delivery channel while its recorder runs, so destroy stops the recorder first, then removes recorder, channel, and retention configuration. Already-recorded configuration items stay queryable until their retention lapses.

**Retention bounds the queryable window** -- unset keeps AWS's default of 2557 days (7 years). Set `retentionPeriodInDays` deliberately: audit regimes want the ceiling, cost-sensitive accounts want the floor of what their compliance window actually requires.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** | `deliveryChannel.s3BucketName` | `status.outputs.bucket_id` |
| **AwsKmsKey** | `deliveryChannel.s3KmsKeyArn` | `status.outputs.key_arn` |
| **AwsSnsTopic** | `deliveryChannel.snsTopicArn` | `status.outputs.topic_arn` |

### What This Kind Provides

`status.outputs` echoes the singleton's identity and state -- `recorder_name` and `delivery_channel_name` (both `default`, AWS's regional convention), `recording_enabled`, and `region` (Config singletons are addressed by region plus the literal name, so verifiers need both). These are audit and verification echoes rather than composition inputs: AWS Config Rule resources in the region depend on the recorder existing, but reference nothing from it.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Scoped recording** -- continuous recording of exactly the types your compliance rules evaluate, delivered with daily snapshots and a one-year retention. The cost-deliberate starting posture; the inclusion list IS the bill, so extend it only as rules demand. Start from the **Scoped Recording** preset.

**Full recording posture** -- everything AWS Config supports, including global types, with a daily-snapshot override on EC2 instances and ENIs. The compliance-first shape for audit regimes that require complete configuration history -- the bill follows the account's activity. Start from the **Full Recording Posture** preset.

**Pause without teardown** -- during a migration or a cost freeze, set `recordingEnabled: false`: the posture, channel, and history all survive while per-item charges stop, and re-enabling is a one-field apply.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the service role Config assumes, wired via `roleArn`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- the history bucket, wired via the delivery channel; it must carry the Config service-principal bucket policy
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- optional encryption for delivered files
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) -- optional delivery notifications
- [**AWS Config Rule**](/infra-catalog/aws-config-rule) -- the evaluations that run over what this recorder captures; rules only see recorded types
- [**AWS Config Aggregator**](/infra-catalog/aws-config-aggregator) -- the multi-account, multi-region rollup of recorded data
