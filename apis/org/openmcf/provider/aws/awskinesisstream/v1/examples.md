# AwsKinesisStream Examples

## 1. Minimal ON_DEMAND Stream

The simplest possible Kinesis stream. AWS manages capacity automatically. No encryption, default 24-hour retention.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStream
metadata:
  name: clickstream-events
  org: acme
  env: dev
  id: clickstream-events-dev
spec:
  streamMode: ON_DEMAND
```

## 2. Minimal PROVISIONED Stream

A provisioned stream with 2 shards (2 MB/s write, 4 MB/s read capacity).

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStream
metadata:
  name: order-events
  org: acme
  env: staging
  id: order-events-staging
spec:
  streamMode: PROVISIONED
  shardCount: 2
```

## 3. Encrypted Stream with Extended Retention

ON_DEMAND stream with KMS encryption and 7-day retention for reprocessing scenarios.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStream
metadata:
  name: payment-events
  org: acme
  env: production
  id: payment-events-prod
spec:
  streamMode: ON_DEMAND
  retentionPeriodHours: 168
  kmsKeyId:
    value: arn:aws:kms:us-east-1:123456789012:key/mrk-abc123
```

## 4. Encrypted Stream with KMS Key Reference (valueFrom)

Uses `valueFrom` to reference a KMS key managed by another OpenMCF resource.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStream
metadata:
  name: audit-events
  org: acme
  env: production
  id: audit-events-prod
spec:
  streamMode: ON_DEMAND
  retentionPeriodHours: 8760  # 365 days
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-encryption-key
      fieldPath: status.outputs.key_arn
```

## 5. Stream with Enhanced Monitoring

PROVISIONED stream with shard-level metrics for production observability. Monitor write throttling, read throttling, and consumer lag per shard.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStream
metadata:
  name: telemetry-stream
  org: acme
  env: production
  id: telemetry-stream-prod
spec:
  streamMode: PROVISIONED
  shardCount: 8
  retentionPeriodHours: 48
  shardLevelMetrics:
    - WriteProvisionedThroughputExceeded
    - ReadProvisionedThroughputExceeded
    - IteratorAgeMilliseconds
```

## 6. Stream with Larger Record Size

ON_DEMAND stream allowing records up to 5 MiB for aggregated event payloads.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStream
metadata:
  name: aggregated-events
  org: acme
  env: production
  id: aggregated-events-prod
spec:
  streamMode: ON_DEMAND
  maxRecordSizeInKib: 5120
  retentionPeriodHours: 72
```

## 7. Production-Ready Analytics Stream

Full-featured production stream: ON_DEMAND for auto-scaling, KMS encryption, 7-day retention, all shard-level metrics, larger record size, and safe deletion behavior.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStream
metadata:
  name: analytics-pipeline
  org: acme
  env: production
  id: analytics-pipeline-prod
  labels:
    team: data-engineering
    cost-center: analytics
spec:
  streamMode: ON_DEMAND
  retentionPeriodHours: 168
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: analytics-kms-key
      fieldPath: status.outputs.key_arn
  maxRecordSizeInKib: 2048
  shardLevelMetrics:
    - IncomingBytes
    - IncomingRecords
    - OutgoingBytes
    - OutgoingRecords
    - WriteProvisionedThroughputExceeded
    - ReadProvisionedThroughputExceeded
    - IteratorAgeMilliseconds
  enforceConsumerDeletion: true
```
