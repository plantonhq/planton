# AWS DynamoDB

Deploys a DynamoDB table end to end: key schema and secondary indexes, capacity (on-demand, provisioned, or price-tuned with warm throughput), DynamoDB Streams and a Kinesis change-data destination, Global Tables v2 multi-region replicas, encryption custody, point-in-time recovery, TTL, contributor insights, a resource-based IAM policy, and creation by restore or S3 import. On PROVISIONED tables the module also owns Application Auto Scaling, so capacity targets and scheduled adjustments live in the same manifest as the table.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **DynamoDB Table** -- a managed NoSQL table in the specified AWS region with the configured primary key schema (partition key and optional sort key), billing mode, and table class
- **Global Secondary Indexes** -- created only when `globalSecondaryIndexes` entries are configured; each GSI has its own key schema (up to 4 HASH + 4 RANGE elements), projection, and capacity riding the table's billing mode
- **Local Secondary Indexes** -- created only when `localSecondaryIndexes` entries are configured; share the table's partition key with an alternate sort key. Create-time only -- they can never be added or removed later
- **Capacity Configuration** -- `provisionedThroughput` on PROVISIONED tables, optional `onDemandThroughput` spend ceilings on PAY_PER_REQUEST tables, and optional increase-only `warmThroughput` floor capacity
- **Server-Side Encryption** -- configured only when `serverSideEncryption` is enabled; switches from the AWS-owned key to the AWS-managed `aws/dynamodb` key or a customer-managed KMS key
- **DynamoDB Streams** -- enabled only when `streamEnabled` is `true`; captures item-level changes with the configured `streamViewType`. Required (with `NEW_AND_OLD_IMAGES`) when the table has replicas
- **Kinesis Streaming Destination** -- configured only when `kinesisStreamingDestination` is provided; fans item-level change data into a Kinesis Data Stream (exactly one destination per table)
- **Global Table Replicas** -- created only when `replicas` entries are configured; each is an active read/write copy in another region, with optional Multi-Region Strong Consistency (`consistencyMode: STRONG` plus either two replicas or one replica and a `globalTableWitness` region)
- **Point-in-Time Recovery** -- enabled only when `pointInTimeRecovery.enabled` is `true`; continuous backups with per-second granularity over the configured recovery window (1-35 days)
- **Time-to-Live** -- enabled only when `ttl.enabled` is `true`; automatically deletes expired items based on an epoch-seconds attribute, free of write cost
- **Contributor Insights** -- enabled only when `contributorInsights.enabled` is `true`; CloudWatch per-key access profiling on the table and optionally on named GSIs
- **Resource Policy** -- attached only when `resourcePolicy` is configured (the policy document as native YAML plus the `confirmRemoveSelfResourceAccess` lockout guard); resource-based IAM grants on the table itself
- **Capacity Auto Scaling** -- on PROVISIONED tables, Application Auto Scaling owns live read/write capacity: target tracking inside `autoscaling` bounds with optional named `scheduledAdjustments`, or pinned min = max targets from `provisionedThroughput` when no autoscaling is configured -- so the block can be added or removed in place
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **No VPC or subnet requirements** -- DynamoDB is a fully managed serverless service that does not require network configuration.
- **A KMS key** (optional) for customer-managed encryption. Reference an `AwsKmsKey` deployed through Planton (its `key_arn` output resolves automatically) or provide a literal key ARN in `serverSideEncryption.kmsKeyArn`. Replicas each need a key in their own region.
- **A Kinesis Data Stream** (optional) in the table's region when configuring the change-data destination.
- **An S3 bucket with source data** (optional) when creating the table by import.

## Deploy

### Console

Open the deployment store, find **AWS DynamoDB**, and click **Deploy**. The creation wizard walks you through the creation source (new, restore, or import), key schema, capacity, indexes, streams, global tables, protection, and operational settings. Start from the **On-Demand Simple Table** preset in the [Presets](#presets) tab for the shape most services need.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsDynamodb
metadata:
  name: orders-table
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  billingMode: PAY_PER_REQUEST
  attributeDefinitions:
    - name: pk
      type: S
    - name: sk
      type: S
  keySchema:
    - attributeName: pk
      keyType: HASH
    - attributeName: sk
      keyType: RANGE
  deletionProtectionEnabled: true
  pointInTimeRecovery:
    enabled: true
```

```shell
planton apply -f dynamodb.yaml
```

This creates an on-demand DynamoDB table with a composite primary key (partition key `pk` and sort key `sk`), deletion protection, and point-in-time recovery enabled. No secondary indexes, streams, or TTL are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When the table deploys alongside its encryption key and change-data stream in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  billingMode: PAY_PER_REQUEST
  attributeDefinitions:
    - name: pk
      type: S
  keySchema:
    - attributeName: pk
      keyType: HASH
  serverSideEncryption:
    enabled: true
    kmsKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: data-key
        fieldPath: status.outputs.key_arn
  kinesisStreamingDestination:
    streamArn:
      valueFrom:
        kind: AwsKinesisStream
        name: table-changes
        fieldPath: status.outputs.stream_arn
```

The InfraPipeline resolves the dependency graph, deploys the KMS key and Kinesis stream first, then provisions the table with the resolved ARNs.

## Key Configuration

These are the most important decisions when configuring a DynamoDB table. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Billing mode** -- Set `billingMode` to `PAY_PER_REQUEST` for on-demand capacity (the right default for almost every table -- zero capacity planning, optional `onDemandThroughput` spend ceilings), or `PROVISIONED` with explicit `provisionedThroughput` read/write capacity units for sustained, predictable traffic where reserved-capacity pricing wins. The mode also decides how every GSI provisions. Switchable in place (one switch per 24 hours).

**Key schema and indexes** -- Define `attributeDefinitions` for the attributes used by keys and indexes (DynamoDB is schemaless beyond that). The primary `keySchema` requires one HASH (partition) element and takes an optional RANGE (sort) element -- create-time immutable. Add `globalSecondaryIndexes` for alternative access patterns (added/changed in place later, one at a time), or `localSecondaryIndexes` for strongly consistent alternate sort orders -- create-time only, and they permanently cap item collections at 10 GB.

**Creation source** -- A table can start empty, restore another table's point-in-time state (`restoreSourceName` same-account, `restoreSourceTableArn` cross-region/cross-account, with `restoreDateTime` or `restoreToLatestTime`), restore a backup (`restoreBackupArn`), or import data from S3 (`importTable` with CSV, DynamoDB JSON, or Ion input). The sources are mutually exclusive; restored tables inherit the source's schema and indexes.

**Global tables** -- Add `replicas` for multi-region active/active serving; replication requires streams with `NEW_AND_OLD_IMAGES`. For Multi-Region Strong Consistency set `consistencyMode: STRONG` on every replica with exactly two replicas, or one replica plus `globalTableWitness`.

**Streams and event processing** -- Set `streamEnabled: true` with a `streamViewType` to capture item-level changes; the `stream_arn` output is what an `AwsLambdaEventSourceMapping` attaches to. Configure `kinesisStreamingDestination` to fan the same changes into a Kinesis Data Stream for analytics pipelines.

**Data protection** -- Enable `deletionProtectionEnabled` to refuse accidental deletion. Enable `pointInTimeRecovery` for continuous backups (tune `recoveryPeriodInDays`, 1-35). Configure `serverSideEncryption` to control key custody -- DynamoDB always encrypts; the choice is whose key.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `serverSideEncryption.kmsKeyArn`, `replicas[].kmsKeyArn` | `status.outputs.key_arn` |
| **AwsKinesisStream** (optional) | `kinesisStreamingDestination.streamArn` | `status.outputs.stream_arn` |
| **AwsS3Bucket** (optional, import) | `importTable.s3Bucket` | `status.outputs.bucket_id` |

Each field also accepts a literal value for resources not managed by Planton.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `table_name` | DynamoDB table name | Application configuration, Lambda function environment variables |
| `table_arn` | Amazon Resource Name of the table | IAM policies, CloudWatch alarms, cross-service integrations |
| `stream_arn` | DynamoDB Streams ARN (when streams are enabled) | `AwsLambdaEventSourceMapping` event sources |
| `stream_label` | Stream label (when streams are enabled) | Stream consumer identification |

`table_id` is also exported -- the provider-assigned table ID, useful for audit rather than downstream wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**On-demand service table** -- A single- or composite-key on-demand table with point-in-time recovery and deletion protection: zero capacity planning, and the right starting point for most services. Start from the **On-Demand Simple Table** preset.

**Provisioned production table** -- A composite-key table with reserved capacity, a GSI, customer-managed encryption, and contributor insights for sustained, predictable traffic where reserved-capacity pricing wins. Start from the **Provisioned Production Table** preset; when traffic is predictable in shape but not in level, the **Provisioned Table with Auto Scaling** preset adds target tracking inside capacity bounds.

**Multi-region active/active** -- A streams-enabled table (`NEW_AND_OLD_IMAGES`, as replication requires) with a replica in a second region. Start from the **Global Table (Multi-Region Active-Active)** preset.

## Works With

- [**AWS Lambda Event Source Mapping**](/infra-catalog/aws-lambda-event-source-mapping) -- consumes the table's `stream_arn` output for change-driven Lambda processing
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides customer-managed encryption keys for the table and its replicas
- [**AWS Kinesis Data Stream**](/infra-catalog/aws-kinesis-stream) -- receives the table's change data through the Kinesis streaming destination
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- provides the source data for table imports
- [**AWS IAM Policy**](/infra-catalog/aws-iam-policy) -- application policies reference the table ARN and name outputs
