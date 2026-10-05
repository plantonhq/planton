# AWS CloudTrail

Deploys a CloudTrail trail — the account's API audit log — recording AWS API calls and delivering them as log files to an S3 bucket, with optional fan-out to CloudWatch Logs and SNS, Insights anomaly detection, and organization-wide capture. Event scope is set by classic or advanced selectors (exactly one style per trail), log-file validation makes tampering detectable, and a multi-region trail captures activity in every region from one home region. This is the first checkbox in every security review.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CloudTrail Trail** — the trail with its delivery bucket and prefix, multi-region and global-service capture, log-file validation, SSE-KMS encryption, SNS delivery notices, CloudWatch Logs mirroring, event selectors, and Insights engines
- **Organization Delegated Admin Registration** — registers an account as the organization's delegated CloudTrail administrator, created only when `organizationDelegatedAdminAccountId` is set (an account-global act; deregistered on destroy)

Destroying the component deletes the trail — a real delete — but the delivered log files stay in the bucket under its lifecycle rules.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with CloudTrail permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- A delivery bucket whose POLICY grants `cloudtrail.amazonaws.com` `s3:GetBucketAcl` on the bucket and `s3:PutObject` under the trail's prefix. AWS validates the policy at trail creation and rejects without it ("Incorrect S3 bucket policy is detected") — the policy lives on the bucket, not on this component, so create the bucket with its policy first.
- (Only for CloudWatch mirroring) a log group and an IAM role trusting `cloudtrail.amazonaws.com` with `logs:CreateLogStream` and `logs:PutLogEvents` on it.
- (Only for SSE-KMS) a KMS key whose policy carries the "Allow CloudTrail to encrypt logs" grant for `cloudtrail.amazonaws.com` — creation fails without the key's consent.
- (Only for organization trails) the organization's management account or delegated administrator, with all-features Organizations enabled.

## Deploy

### Console

Open the deployment store, find **AWS CloudTrail**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the delivery bucket, and the event scope. Start from the **Compliance Audit Trail** preset in the [Presets](#presets) tab for the audit posture: multi-region, validated, with both Insights engines.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudTrail
metadata:
  name: audit-trail
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  s3BucketName:
    valueFrom:
      kind: AwsS3Bucket
      name: audit-logs
      fieldPath: status.outputs.bucket_id
  s3KeyPrefix: audit
  isMultiRegionTrail: true
  enableLogFileValidation: true
  advancedEventSelectors:
    - name: Management events
      fieldSelectors:
        - field: eventCategory
          equals: ["Management"]
```

```shell
planton apply -f cloudtrail.yaml
```

This creates a multi-region trail recording all management events with hourly digest files for tamper detection, delivering to the referenced bucket under `audit/AWSLogs/<account-id>/`. An Infra Job tracks the provisioning in real time.

### InfraChart

When the trail deploys alongside its delivery bucket in one chart, wire the bucket reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  s3BucketName:
    valueFrom:
      kind: AwsS3Bucket
      name: audit-logs
      fieldPath: status.outputs.bucket_id
  isMultiRegionTrail: true
  enableLogFileValidation: true
```

The InfraPipeline resolves the dependency graph, creates the bucket (with its CloudTrail bucket policy) first, then provisions the trail against it.

## Key Configuration

These are the most important decisions when configuring a trail. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Multi-region plus validation is the audit posture** — A single-region trail misses activity in every other region, and without `enableLogFileValidation`'s hourly digest chain, tampering with delivered logs is undetectable. Turn both on for any trail that exists for audit; a multi-region trail still has exactly one home region (`region`) and is managed from there.

**One selector style, and it should be advanced** — Classic `eventSelectors` and `advancedEventSelectors` are mutually exclusive (the spec rejects both at validate time). Prefer the advanced style for new trails: it is the one AWS extends, and it matches on `eventName`, `resources.ARN` prefixes, and `eventCategory`. Every advanced selector needs an `eventCategory` field selector or AWS rejects it.

**Data events and Insights are the cost levers** — Management events are the baseline; a second trail delivering the same management events adds cost for duplicate copies, and data events and Insights bill per event volume. Scope data-event selectors to the specific buckets, functions, or tables you must audit rather than `arn:aws:s3` (all buckets), and enable `insightTypes` knowing the anomaly engines bill on top.

**The bucket policy is a create-time gate** — AWS validates the delivery bucket's policy when the trail is created, not lazily. In a chart this ordering is handled by the reference; outside one, a bucket created without the CloudTrail service-principal policy fails the trail's first apply.

**CloudWatch mirroring travels as a pair** — AWS requires the log group and the delivery role together, so `cloudwatchLogs` wires both or neither; half-configured mirroring cannot be expressed. S3 delivery continues regardless — the mirror buys live querying and metric filters, at CloudWatch ingestion cost.

**Organization trails have a home** — `isOrganizationTrail` captures every member account into this one trail but only works from the management account or the delegated administrator. The delegated-admin registration (`organizationDelegatedAdminAccountId`) is one-per-organization and account-global — most deployments leave it unset, and the spec requires it to ride an organization trail.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsS3Bucket** | `s3BucketName` | `status.outputs.bucket_id` |
| **AwsKmsKey** | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsSnsTopic** | `snsTopicName` | `status.outputs.topic_name` |
| **AwsCloudwatchLogGroup** | `cloudwatchLogs.logGroupArn` | `status.outputs.log_group_arn` |
| **AwsIamRole** | `cloudwatchLogs.roleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `sns_topic_arn` | ARN of the delivery-notice topic (set only when `snsTopicName` is configured) | Subscribing queues or functions that process each delivered log file |
| `trail_arn` | The trail's ARN (also the provider's import ID) | IAM policies scoping trail administration |
| `home_region` | The region that manages a multi-region trail | Operational tooling that must address the trail in its home region |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The compliance audit trail** — multi-region, log-file validation on, management events via one advanced selector, both Insights engines. One per account (or one organization trail for the whole estate) into a locked-down bucket whose lifecycle rules ARE the audit retention — destroy deletes the trail, never the evidence. Start from the **Compliance Audit Trail** preset.

**Scoped S3 data-events trail** — a second trail recording only object-level writes on specific buckets, using `resources.ARN` prefix matching and `readOnly: ["false"]`. Data events on everything is a runaway bill; prefix-scoped selectors keep the audit signal without it. Start from the **S3 Data-Events Trail** preset.

**Live security monitoring** — add `cloudwatchLogs` mirroring to the audit trail and build metric filters and alarms on the log group (root logins, IAM policy changes, console sign-ins without MFA). S3 remains the durable record; CloudWatch is the tripwire.

## Works With

- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the delivery destination; its `spec.policy` carries the CloudTrail service-principal grant
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — SSE-KMS encryption of delivered log files
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) — per-file delivery notices for downstream processing
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — the live mirror for Logs Insights queries and metric filters
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the role CloudTrail assumes to write into the log group
- [**AWS CloudTrail Event Data Store**](/infra-catalog/aws-cloud-trail-event-data-store) — CloudTrail Lake: SQL-queryable event storage that complements file delivery
