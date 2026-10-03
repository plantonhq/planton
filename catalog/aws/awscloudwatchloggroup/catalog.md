# AWS CloudWatch Log Group

Deploys a CloudWatch Logs log group with configurable retention, optional customer-managed KMS encryption, log group class selection, and deletion protection. Log groups are the centralized destination for application logs, service logs, and operational data across AWS services -- and this component also carries the processing attached to the group: metric filters that turn log patterns into alarmable metrics, subscription filters that stream events to Kinesis, Firehose, or Lambda, and an ingest-time transformer pipeline.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CloudWatch Log Group** -- a log group container for log streams, configured with the specified retention period, log group class, and optional KMS encryption
- **Retention Policy** -- created only when `retentionInDays` is set to a non-zero value; automatically deletes log events after the specified number of days
- **KMS Encryption Configuration** -- configured only when `kmsKeyId` is provided; encrypts log data at rest with a customer-managed key instead of the default SSE-CWL encryption
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A KMS key** (optional) -- required only when using customer-managed encryption. The key must be in the same region as the log group, and its policy must grant `logs.amazonaws.com` the required permissions. Provide the ARN directly or reference an AwsKmsKey Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **AWS CloudWatch Log Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard 30-Day Retention** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogGroup
metadata:
  name: app-logs
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  retentionInDays: 30
```

```shell
planton apply -f log-group.yaml
```

This creates a STANDARD class log group with 30-day retention and default AWS encryption. No KMS encryption or deletion protection is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When using customer-managed encryption, use ValueFromRef to wire the log group to a KMS key deployed in the same InfraPipeline:

```yaml
spec:
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: log-encryption-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the KMS key first, then provisions the log group with customer-managed encryption using the resolved key ARN.

## Key Configuration

These are the most important decisions when configuring a CloudWatch log group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Retention period** -- Set `retentionInDays` to control how long log events are kept. A value of 0 (the default) retains logs indefinitely, which accumulates storage costs over time. Common values: 30 days for dev/staging, 90 days for production, 365 days for compliance. AWS only accepts specific values (1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653).

**Log group class** -- STANDARD (default) provides the full feature set including metric filters, subscription filters, and Contributor Insights. INFREQUENT_ACCESS offers ~50% cheaper storage but does not support metric filters or subscription filters -- best for high-volume logs queried rarely (VPC flow logs, CDN access logs). DELIVERY is purpose-built for AWS service log delivery at the lowest cost.

**KMS encryption** -- By default, CloudWatch Logs uses SSE-CWL encryption. Provide `kmsKeyId` with a customer-managed KMS key ARN when you need key rotation control, cross-account access via key policy, or CloudTrail audit trails of log data access. Required for most compliance frameworks.

**Deletion protection** -- Set `deletionProtectionEnabled: true` to prevent accidental deletion of production log groups. Any destroy operation fails until the flag is cleared -- and clearing requires an EXPLICIT `false`: leaving the field unset keeps whatever protection state the group already has, so protection never turns off by omission.

**Metric filters** -- Turn log patterns into CloudWatch metrics without any processing code: each entry in `metricFilters` pairs a filter pattern with a metric transformation (name, namespace, value -- "1" counts occurrences, "$field" publishes an extracted value, up to 3 dimensions, an optional default for non-matches). The cheapest path from a log line to an alarm.

**Subscription filters** -- Stream matching log events in real time to a Kinesis data stream, Firehose delivery stream, or Lambda function (up to 2 per log group). Kinesis and Firehose destinations need a `roleArn` CloudWatch Logs assumes to write; cross-account destinations are referenced by literal ARN. `emitSystemFields` can stamp `@aws.account` / `@aws.region` / `@source.log` onto forwarded events for centralized multi-account processing.

**Transformer** -- An ingest-time pipeline of 1--20 processors that parses and reshapes every log event before anything else sees it: parse JSON/grok/CSV/key-value (or one of the vended AWS log formats -- CloudFront, PostgreSQL, Route 53, VPC flow, WAF, OCSF), then rename/add/copy/move/delete keys, convert types and datetimes, and normalize strings. The first processor must be a parser, and transformed events are what Logs Insights queries see; metric and subscription filters opt in per filter with `applyOnTransformedLogs`. STANDARD class only, one per group.

**Pre-created log streams** -- Most log streams are created at runtime by the writing agent, and that is the right default. List names in `logStreams` only when something depends on the stream existing before the first write: an agent configured with a fixed stream name, or IAM policies scoped to specific stream ARNs.

**Data policies** -- `dataProtectionPolicy` scans ingested events for sensitive data (AWS-managed identifiers for emails, credentials, cards) and masks it for everyone without the `logs:Unmask` permission. `fieldIndexPolicy` names up to 20 frequently-queried JSON fields that Logs Insights indexes for faster, cheaper queries.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsKinesisStream** (optional) | `subscriptionFilters[*].destinationArn` | `status.outputs.stream_arn` |
| **AwsKinesisFirehose** (optional) | `subscriptionFilters[*].destinationArn` | `status.outputs.delivery_stream_arn` |
| **AwsLambda** (optional) | `subscriptionFilters[*].destinationArn` | `status.outputs.function_arn` |
| **AwsIamRole** (optional) | `subscriptionFilters[*].roleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `log_group_arn` | Amazon Resource Name of the log group | Step Functions logging destination, API Gateway access logs, OpenSearch log publishing |
| `log_group_name` | Name of the log group | ECS awslogs driver, ElastiCache log delivery, Lambda log group references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard 30-day retention** -- STANDARD class with 30-day retention and default encryption. The general-purpose configuration for application logs in dev, staging, and most production workloads. Start from the **Standard 30-Day Retention** preset.

**Encrypted 90-day retention** -- STANDARD class with 90-day retention and customer-managed KMS encryption. Suitable for production workloads subject to SOC2, HIPAA, or PCI-DSS requirements. Start from the **Encrypted 90-Day Retention** preset.

**Infrequent access long retention** -- INFREQUENT_ACCESS class with 365-day retention and KMS encryption. ~50% cheaper storage for high-volume logs queried rarely, such as VPC flow logs and CDN access logs. Start from the **Infrequent Access Long Retention** preset.

**Error-count metric filter** -- A metric filter counting ERROR-level log lines into a custom metric, ready for an AwsCloudwatchAlarm to page on. The cheapest path from a log line to an alert -- no processing code, no subscription. Start from the **Error-Count Metric Filter** preset.

**Transformed ingest pipeline** -- A transformer that parses raw events at ingest (JSON, grok, or a vended AWS format) and reshapes keys before Logs Insights, metric filters, or subscribers see them. Normalize once at the group instead of in every query. Start from the **Transformed Ingest Pipeline** preset.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for encrypting log data at rest
- [**AWS Kinesis Data Stream**](/infra-catalog/aws-kinesis-stream) -- real-time subscription filter destination for custom consumers
- [**AWS Kinesis Firehose**](/infra-catalog/aws-kinesis-firehose) -- buffered subscription filter destination for delivery to S3, OpenSearch, and more
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- per-batch subscription filter processing with your own code
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the delivery role CloudWatch Logs assumes for Kinesis and Firehose destinations