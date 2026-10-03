# AWS Bedrock Invocation Logging

Manages a region's Amazon Bedrock model invocation logging configuration — the audit trail of every model call in that region, delivered to CloudWatch Logs, S3, or both. This is a settings singleton: AWS keeps exactly one logging configuration per account and region, the region IS the resource identity, and deploying this component takes ownership of it — deploy at most one instance per region. Without invocation logging there is no record of what prompts were sent or what the models returned; with it, the logs contain your prompts and completions, so their destinations deserve the same access control as the models themselves.

## What Gets Created

When you deploy this Infra Component, the IaC module configures the region's one invocation logging object — it adopts the account+region settings singleton rather than creating a new named resource:

- **Data-type capture** — which invocation payloads are logged: text, image, embedding, and video, each an explicit toggle (AWS defaults all four to enabled; unset inherits that)
- **CloudWatch delivery** — configured only when `cloudwatch` is set: the log group and the IAM role Bedrock assumes to write to it, with optional S3 spillover for payloads too large for a log event
- **S3 delivery** — configured only when `s3` is set: the bucket and key prefix Bedrock's service principal writes invocation records to

Destroying this component deletes the configuration — logging stops region-wide.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock logging permissions (`bedrock:PutModelInvocationLoggingConfiguration` and its read/delete siblings, plus `iam:PassRole` for the CloudWatch arm). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **For CloudWatch delivery** — a log group in the same region (Bedrock does not create it) and an IAM role trusting `bedrock.amazonaws.com` with `logs:CreateLogStream` and `logs:PutLogEvents` on that group. AWS validates this permission chain at apply time.
- **For S3 delivery** — a bucket in the same region whose bucket policy allows `bedrock.amazonaws.com` to `s3:PutObject`, scoped with an `aws:SourceAccount` condition. Bedrock writes as its own service principal here — an IAM role with S3 permissions does nothing for this arm.

## Deploy

### Console

Open the deployment store, find **AWS Bedrock Invocation Logging**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region, data-type toggles, and the delivery destinations. Start from the **Full-Fidelity Audit** preset in the [Presets](#presets) tab for the CloudWatch-plus-S3 posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockInvocationLogging
metadata:
  name: bedrock-logging-us-west-2
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  cloudwatch:
    logGroupName:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: bedrock-invocations
        fieldPath: status.outputs.log_group_name
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: bedrock-logging-role
        fieldPath: status.outputs.role_arn
    largeDataDeliveryS3:
      bucketName:
        valueFrom:
          kind: AwsS3Bucket
          name: bedrock-invocation-archive
          fieldPath: status.outputs.bucket_id
      keyPrefix: bedrock/large-payloads
  s3:
    bucketName:
      valueFrom:
        kind: AwsS3Bucket
        name: bedrock-invocation-archive
        fieldPath: status.outputs.bucket_id
    keyPrefix: bedrock/invocations
```

```shell
planton apply -f bedrock-invocation-logging.yaml
```

This configures full-fidelity logging for us-west-2: CloudWatch for querying, S3 for retention, and oversized payloads spilled to S3 instead of being truncated. An Infra Job tracks the provisioning in real time.

### InfraChart

When the logging configuration deploys alongside its log group, role, and bucket in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  cloudwatch:
    logGroupName:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: bedrock-invocations
        fieldPath: status.outputs.log_group_name
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: bedrock-logging-role
        fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the log group and role first, then writes the logging configuration that references them.

## Key Configuration

These are the most important decisions when configuring invocation logging. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One instance per region, period** — the region is the singleton's identity; two instances targeting the same region fight over the same configuration object. `metadata.name` never reaches AWS — it is Planton-side identity only.

**CloudWatch, S3, or both** — CloudWatch is for querying and alerting on recent invocations; S3 is for full-fidelity, long-horizon retention. Both together are the canonical audit posture. A configuration must have at least one destination — the manifest validation refuses a destination-less shape because it would deliver nothing.

**Two authorization models, one configuration** — the CloudWatch arm authenticates through the ROLE (`roleArn` trusting `bedrock.amazonaws.com` with write access to the group); the S3 arm authenticates through the BUCKET POLICY (Bedrock's service principal, not a role). Granting a role S3 permissions does nothing for the S3 arm — this split is the most common setup mistake.

**Size the CloudWatch arm for the 256 KB event cap** — model inputs and outputs routinely exceed CloudWatch's per-event limit. Without `largeDataDeliveryS3`, oversized payloads are truncated out of the stream; with it, they land whole in S3.

**The data-type toggles are the cost lever** — AWS defaults all four (text, image, embedding, video) to enabled. Text-first workloads usually set `embeddingDataDeliveryEnabled` and `videoDataDeliveryEnabled` to false: embedding vectors are bulky and rarely audit-relevant.

**Bedrock leaves permission-check canaries in your buckets** — configuring any S3 destination writes zero-byte `amazon-bedrock-logs-permission-check` objects under every configured prefix (with zero invocations made), and they survive deleting the logging configuration. A bucket that has ever been a destination is never empty — give it force-destroy or a lifecycle rule if you expect to delete it later.

**Destroy is region-wide silence** — deleting this component deletes the configuration and the region reverts to no invocation logging. Treat it like disabling audit logging, because that is exactly what it is.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCloudwatchLogGroup** | `cloudwatch.logGroupName` | `status.outputs.log_group_name` |
| **AwsIamRole** | `cloudwatch.roleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** | `s3.bucketName`, `cloudwatch.largeDataDeliveryS3.bucketName` | `status.outputs.bucket_id` |

### What This Kind Provides

This kind has no consumable outputs: `status.outputs` carries only `configured_region`, which echoes the region the instance owns — the singleton's identity and import ID, not a value downstream resources compose on. The configuration's effect is the log streams and objects that appear in the destinations you referenced.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Full-fidelity audit** — CloudWatch delivery for querying plus S3 delivery for retention, with `largeDataDeliveryS3` so nothing is truncated. The right posture for regulated workloads and for any team that will eventually be asked "what did the model say". Start from the **Full-Fidelity Audit** preset.

**S3 archive only** — a single S3 destination for teams that need the audit trail but not interactive querying; cheaper to retain at volume, and queryable later with Athena-class tooling over the bucket. The trade: no CloudWatch metric filters or alarms on invocation content. Start from the **S3 Archive Only** preset.

## Works With

- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — the query-side destination, wired via `cloudwatch.logGroupName`
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the delivery role Bedrock assumes for CloudWatch writes, wired via `cloudwatch.roleArn`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the retention destination and the large-payload spillover, wired via `bucketName` references
