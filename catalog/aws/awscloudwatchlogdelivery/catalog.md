# AWS CloudWatch Logs Delivery

Deploys the two ways logs leave CloudWatch: the modern vended-log pipeline (a delivery source wrapping one AWS resource whose service vends logs, owned destinations in S3, CloudWatch Logs, Firehose, or X-Ray, and the deliveries joining them) and the legacy cross-account destination (a named Kinesis-backed endpoint other accounts' subscription filters send to). The two arms deploy independently — one instance can carry either or both. AWS allows at most one delivery per (source, destination-type) pair, and a destination shared by many pipelines is owned by one instance and referenced by ARN from the rest.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Delivery Source** — created only when `vended.source` is set: wraps one AWS resource (a CloudFront distribution, a Bedrock knowledge base, …) as a named source for one log type. AWS models sources per (resource, log type) — a second log type from the same resource is a second source.
- **Delivery Destinations** — one per `vended.destinations` entry: a named wrapper around the receiving S3 bucket, log group, Firehose stream, or the account's X-Ray trace store. Each destination's cross-account `policy` is attached only when set.
- **Deliveries** — one per `vended.deliveries` entry, joining this instance's source to an owned destination by name or an external destination by ARN, with record fields, delimiter, and S3 layout settings.
- **Cross-Account Log Destination and its access policy** — created only when `crossAccountDestination` is set: the legacy Kinesis-backed endpoint with the policy naming the producer accounts allowed to subscribe.
- **AWS Tags** — resource metadata tags applied automatically on the taggable objects (the cross-account destination is tagged in a separate call after create — tags on the Put call break AWS's test-message delivery check).

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, carrying CloudWatch Logs delivery permissions plus IAM pass-role for the cross-account arm. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **The producing resource** (vended arm) — must exist, and its service must support vended log delivery for the `logType` you name; each service's vended-logs documentation lists its types.
- **The receiving resource** (vended arm) — the S3 bucket, log group, or Firehose stream behind each owned destination.
- **A delivery role** (only for the cross-account arm) — an IAM role whose trust policy allows `logs.amazonaws.com`, with write access to the target Kinesis stream. The first create retries for up to two minutes while that trust propagates — a slow first apply is normal.

## Deploy

### Console

Open the deployment store, find **AWS CloudWatch Logs Delivery**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the two arms' spec fields. Start from the **CloudFront Access Logs to S3** or **Organization Log Sink** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogDelivery
metadata:
  name: cdn-access-logs
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  vended:
    source:
      name: cdn-access-logs
      logType: ACCESS_LOGS
      resourceArn:
        value: arn:aws:cloudfront::123456789012:distribution/E2EXAMPLE
    destinations:
      - name: cdn-log-archive
        destinationResourceArn:
          value: arn:aws:s3:::acme-cdn-log-archive
        outputFormat: parquet
    deliveries:
      - name: to-s3
        destinationName: cdn-log-archive
        s3Configuration:
          enableHiveCompatiblePath: true
          suffixPath: cdn-logs
```

```shell
planton apply -f log-delivery.yaml
```

This creates a vended pipeline delivering the CloudFront distribution's access logs to an S3 archive as Parquet with Hive-compatible partitioning — Athena queries it with no ETL. An Infra Job tracks the provisioning in real time.

### InfraChart

When the pipeline deploys alongside its receiving bucket in one chart, wire the destination reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  vended:
    source:
      name: cdn-access-logs
      logType: ACCESS_LOGS
      resourceArn:
        value: arn:aws:cloudfront::123456789012:distribution/E2EXAMPLE
    destinations:
      - name: cdn-log-archive
        destinationResourceArn:
          valueFrom:
            kind: AwsS3Bucket
            name: cdn-log-archive
            fieldPath: status.outputs.bucket_arn
    deliveries:
      - name: to-s3
        destinationName: cdn-log-archive
        s3Configuration:
          enableHiveCompatiblePath: true
```

The InfraPipeline resolves the dependency graph, deploys the bucket first, then provisions the source, destination, and delivery against it.

## Key Configuration

These are the most important decisions when configuring a log delivery. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Vended or cross-account (or both)** — the vended arm is the modern framework for AWS services that vend logs (CloudFront, Bedrock, SES, …); the cross-account arm is the legacy Kinesis-backed endpoint that other accounts' subscription filters target. They solve different problems and deploy independently — an org log sink is the legacy arm, a CloudFront-to-Athena pipeline is the vended arm.

**One delivery per (source, destination-type)** — AWS accepts at most one delivery from a source to each destination type: S3 plus Firehose is fine, two S3 destinations from one source is a ConflictException. Fan out to multiple buckets via Firehose or replicate downstream.

**Source identity is total** — the source's `name`, `logType`, and `resourceArn` all replace the source on change, and AWS models sources per (resource, log type). Shipping a second log type from the same resource means a second instance of this component.

**Own the shared destination once** — a destination shared by many pipelines lives in one owning instance; every other instance's deliveries reference it through `destinationArn` (fed by the owner's `destination_arns` output). The owner also carries the destination `policy` granting producer accounts `logs:CreateDelivery` — same-account pipelines never need that policy.

**S3 layout for Athena** — `enableHiveCompatiblePath: true` writes `year=.../month=...` partitions so Glue and Athena partition discovery work out of the box. For CloudFront sources AWS silently prepends `AWSLogs/{account-id}/CloudFront/` to the key path; configure only your own segment in `suffixPath` — both engines store the suffix without the AWS-added prefix, so state never fights it.

**The cross-account access policy has sharp edges** — write AWS principals as bare account IDs (`"AWS": "123456789012"`): the service rejects the usual IAM root-ARN spelling with "Principal section of policy contains ARN instead of account ID". In YAML, quote the ID — an unquoted account ID parses as a number and IAM rejects a numeric principal. And note the AWS contract: destroying the policy alone never removes it at AWS; only destroying the whole destination does. This component folds both, so a normal teardown is clean.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsS3Bucket** (S3 destination) | `vended.destinations[].destinationResourceArn` | `status.outputs.bucket_arn` |
| **AwsCloudwatchLogGroup** (CWL destination) | `vended.destinations[].destinationResourceArn` | `status.outputs.log_group_arn` |
| **AwsIamRole** (cross-account arm) | `crossAccountDestination.roleArn` | `status.outputs.role_arn` |
| **AwsKinesisStream** (cross-account arm) | `crossAccountDestination.targetArn` | `status.outputs.stream_arn` |

The source's `resourceArn` is also a reference field, but its kind is open by design — any vended-logs producer works (reference the producing kind's ARN output or pass a literal ARN). A delivery's `destinationArn` references another instance's owned destination through its `destination_arns` output.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `destination_arns` | Owned destination ARNs keyed by destination name | Other instances' `deliveries[].destinationArn` — how one shared destination serves many pipelines |
| `cross_account_destination_arn` | The legacy endpoint's ARN | What producer accounts' subscription filters target |

The remaining outputs are identity records rather than composition inputs: `source_arn`, `source_name`, and `source_service` describe the vended source, `delivery_ids` and `delivery_arns` are the deliveries' import handles, and `cross_account_destination_name` is the legacy arm's import ID.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CloudFront access logs to S3** — the modern access-log pipeline: a CloudFront source delivering Parquet to an S3 archive with Hive-compatible partitioning, queryable by Athena with zero ETL. Start from the **CloudFront Access Logs to S3** preset.

**Organization log sink** — the central-account half of cross-account aggregation: a Kinesis-backed destination whose access policy lists the producer accounts, each of which points subscription filters at the destination's ARN. Every account's logs land in one stream. Start from the **Organization Log Sink** preset.

**Shared central destination for vended logs** — one instance owns an S3 destination with a `policy` granting the producer accounts `logs:CreateDelivery`; each producer account deploys its own source and a delivery referencing the shared destination by ARN. The vended-framework equivalent of the org sink, without Kinesis in the middle.

## Works With

- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the archive behind S3 destinations, referenced by bucket ARN
- [**AWS CloudFront**](/infra-catalog/aws-cloud-front) — a common vended source: its distribution ARN feeds `resourceArn` for ACCESS_LOGS
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — the receiving group behind CWL destinations
- [**AWS Kinesis Data Stream**](/infra-catalog/aws-kinesis-stream) — the stream behind the legacy cross-account destination
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the role CloudWatch Logs assumes to write into the stream
