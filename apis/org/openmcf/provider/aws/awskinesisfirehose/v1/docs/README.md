# AwsKinesisFirehose — Architecture Reference

## Overview

Amazon Kinesis Data Firehose is a fully managed service for loading streaming data into storage and analytics destinations. Unlike Kinesis Data Streams (which requires you to write and operate consumers), Firehose handles the entire delivery pipeline: buffering, transformation, format conversion, compression, and retry — with zero consumer code.

Firehose is the simplest path from streaming data to S3, OpenSearch, HTTP endpoints, and Redshift.

## How Firehose Works

### Data Flow

```
Source → Buffer → [Transform] → [Convert] → [Compress] → Destination
                                                    ↓ (failures)
                                               S3 Backup
```

1. **Ingest** — Data enters the delivery stream from a source (Direct PUT or Kinesis Data Stream)
2. **Buffer** — Records accumulate in an in-memory buffer until either the size or time threshold is reached
3. **Transform** (optional) — A Lambda function processes each batch, returning transformed records
4. **Convert** (optional) — JSON records are converted to columnar format (Parquet/ORC) using a Glue schema
5. **Compress** (optional) — Data is compressed before writing (GZIP, Snappy, etc.)
6. **Deliver** — The batch is written to the destination
7. **Backup** — Failed records (or all records, if configured) are written to S3

Each delivery stream has exactly **one source** and exactly **one destination**. Both are immutable after creation (ForceNew).

## Source Types

### Direct PUT (Default)

Applications call the `PutRecord` or `PutRecordBatch` API directly on the delivery stream. This is the simplest source — no intermediate streaming infrastructure required.

**Characteristics:**

- Up to 1,000 records/s or 1 MB/s per delivery stream (soft limit, can be increased)
- Records up to 1 MiB each
- Server-side encryption (SSE) can be enabled on the delivery stream buffer
- No ordering guarantees — records may arrive out of order
- No replay — once delivered, records cannot be re-read from the source

**When to use:**

- Application produces data at moderate throughput (<1 MB/s)
- No need for ordered processing or replay
- Want the simplest possible architecture (no Kinesis stream to manage)
- Log forwarding from CloudWatch Logs, IoT Core, or other AWS services

### Kinesis Data Stream Source

Firehose reads from an existing Kinesis Data Stream, acting as a managed consumer with automatic checkpointing and retry.

**Characteristics:**

- Firehose creates an internal consumer and reads all shards
- Inherits the stream's ordering guarantees (per-shard, per-partition-key)
- No SSE on the delivery stream — encryption is handled by the source stream
- Stream must exist before the delivery stream is created
- Source configuration is entirely ForceNew

**When to use:**

- Multiple consumers need the same data (Firehose + Lambda + custom app)
- Need replay capability (Kinesis retains data for 24h–365 days)
- Need ordering guarantees (per partition key)
- High throughput (>1 MB/s) with auto-scaling (ON_DEMAND stream)
- Already have a Kinesis stream in your architecture

### Source Comparison

| Feature | Direct PUT | Kinesis Stream Source |
|---------|------------|---------------------|
| Setup complexity | Lowest | Requires existing stream |
| Throughput | 1,000 rec/s or 1 MB/s (soft limit) | Stream capacity (unlimited with ON_DEMAND) |
| Ordering | None | Per-shard (partition key) |
| Replay | No | Yes (stream retention) |
| Multiple consumers | No | Yes (stream supports many readers) |
| SSE | On delivery stream buffer | On source stream |
| Cost | Firehose per-GB only | Stream cost + Firehose per-GB |

## Destination Types

### Destination Comparison

| Feature | Extended S3 | OpenSearch | HTTP Endpoint | Redshift |
|---------|-------------|-----------|---------------|----------|
| **Use case** | Data lake, archive, analytics | Log analytics, search | Third-party SaaS, custom APIs | Data warehouse |
| **Delivery path** | Direct to S3 | Direct to OpenSearch index | HTTPS POST | S3 staging → COPY |
| **Format conversion** | Parquet/ORC via Glue | No | No | No |
| **Dynamic partitioning** | Yes | No | No | No |
| **Lambda transformation** | Yes | Yes | Yes | Yes |
| **S3 backup** | Optional (source records) | Required (failed or all docs) | Required (failed or all data) | Required (staging) + optional (source) |
| **VPC delivery** | N/A (S3 is a regional service) | Yes | No | No (public JDBC) |
| **Max buffer size** | 128 MiB | 100 MiB | 64 MiB | N/A (S3 staging) |
| **Index/table rotation** | Via prefix expressions | Built-in rotation periods | N/A | N/A |
| **Usage share** | ~60% | ~15% | ~15% | ~10% |

### Extended S3

The most feature-rich destination and the most common. Data lands in S3 as objects, optionally compressed, converted to columnar format, and dynamically partitioned by record fields.

**Best for:** Data lakes, log archives, analytics pipelines (Athena, Spark, Presto), long-term storage, compliance archives.

### OpenSearch

Indexes records directly into an Amazon OpenSearch Service domain. Supports index rotation (hourly, daily, weekly, monthly) and VPC delivery for private clusters. Failed documents are always backed up to S3.

**Best for:** Log analytics, full-text search, real-time dashboards (OpenSearch Dashboards/Kibana).

### HTTP Endpoint

Delivers to any HTTPS endpoint that accepts POST requests and returns HTTP 200 on success. The endpoint receives JSON arrays of records. Authentication is via a configurable access key in the `X-Amz-Firehose-Access-Key` header.

**Best for:** Third-party integrations (Datadog, New Relic, Sumo Logic, Splunk via HEC, Honeycomb), custom APIs, webhook-based pipelines.

### Redshift

A two-stage destination: Firehose writes data to an S3 staging bucket, then issues a Redshift `COPY` command to bulk-load the data. This is the standard Redshift ingestion pattern for streaming data.

**Best for:** Data warehouse loading, business intelligence, reporting pipelines.

## Buffering Model

Firehose buffers incoming records and flushes to the destination when **either** threshold is reached — whichever comes first:

- **Buffer interval** — Time since the last flush (0–900 seconds, default 300)
- **Buffer size** — Accumulated data size (1–128 MiB, default 5 MiB)

### How Buffering Works

```
Records arrive → Buffer fills
                   ├── Size threshold reached? → Flush
                   └── Time threshold reached? → Flush
```

**Example:** With `intervalInSeconds: 300` and `sizeInMbs: 5`:

- If 5 MiB accumulates in 30 seconds → flushes after 30 seconds (size trigger)
- If only 100 KiB arrives in 300 seconds → flushes after 300 seconds (time trigger)

### Tuning Guidelines

| Goal | Interval | Size | Effect |
|------|----------|------|--------|
| Low latency | 60s | 1 MiB | More frequent, smaller deliveries |
| Optimal for analytics | 900s | 128 MiB | Fewer, larger files (better for Athena/Spark) |
| Balanced | 300s | 5 MiB | Default — good for most use cases |
| Parquet conversion | 900s | 128 MiB | Larger batches produce better Parquet files |

**Destination-specific limits:**

- Extended S3: 1–128 MiB
- OpenSearch: 1–100 MiB
- HTTP Endpoint: 1–64 MiB
- Redshift: Uses S3 staging buffering

## Lambda Transformation Pipeline

### How It Works

1. Firehose accumulates records in a processing buffer (1–3 MiB, 60–900s)
2. When the buffer threshold is reached, Firehose invokes the Lambda function with a batch of records
3. Lambda processes each record and returns a result with a **status code** per record
4. Firehose routes records based on their status code

### Lambda Input/Output

**Input event structure:**

```json
{
  "invocationId": "...",
  "deliveryStreamArn": "arn:aws:firehose:...",
  "region": "us-east-1",
  "records": [
    {
      "recordId": "...",
      "approximateArrivalTimestamp": 1234567890,
      "data": "<base64-encoded record>"
    }
  ]
}
```

**Output structure:**

```json
{
  "records": [
    {
      "recordId": "...",
      "result": "Ok",
      "data": "<base64-encoded transformed record>"
    }
  ]
}
```

### Record Status Codes

| Status | Behavior |
|--------|----------|
| `Ok` | Record is delivered to the destination |
| `Dropped` | Record is intentionally discarded (counted as successfully processed) |
| `ProcessingFailed` | Record failed transformation — written to the error output prefix |

### Retry Behavior

- If Lambda returns an error (exception, timeout, throttle), Firehose retries the entire batch
- Retries up to `numberOfRetries` times (default 3, max 300)
- After all retries are exhausted, the batch is written to the error output prefix in S3
- Lambda timeout should be ≤ 5 minutes (Firehose's invocation timeout is 5 minutes)

### Dynamic Partitioning with Lambda

Lambda can extract partition keys from records by adding metadata to the response:

```json
{
  "recordId": "...",
  "result": "Ok",
  "data": "<base64>",
  "metadata": {
    "partitionKeys": {
      "customer_id": "cust-123",
      "region": "us-east-1"
    }
  }
}
```

These keys are referenced in the S3 prefix: `data/customer=!{partitionKeyFromLambda:customer_id}/`

## Data Format Conversion

### JSON → Columnar Conversion

Firehose can convert incoming JSON records to Apache Parquet or Apache ORC format using an AWS Glue Data Catalog schema. This dramatically improves query performance and reduces storage costs for analytics workloads.

### How It Works

1. Firehose receives JSON records
2. Deserializes using the configured input format (OpenX JSON SerDe or Hive JSON SerDe)
3. Maps fields to the Glue table schema
4. Serializes to the output columnar format (Parquet or ORC)
5. Applies columnar-native compression (SNAPPY, GZIP, ZLIB)
6. Writes to S3

### Performance Impact

| Metric | JSON (GZIP) | Parquet (SNAPPY) | Improvement |
|--------|-------------|------------------|-------------|
| Storage size | 1x | 0.1–0.4x | 60–90% smaller |
| Athena scan cost | Full file scan | Columnar pruning | 10–100x cheaper |
| Query latency | Seconds to minutes | Sub-second to seconds | 10–100x faster |

### Parquet vs ORC

| Feature | Parquet | ORC |
|---------|---------|-----|
| Best for | Athena, Spark, Presto, general analytics | Hive workloads |
| Compression | SNAPPY (default), GZIP, UNCOMPRESSED | SNAPPY (default), ZLIB, NONE |
| Predicate pushdown | Yes | Yes |
| ACID support | No | Yes (Hive ACID) |
| Recommendation | **Use Parquet** for most analytics workloads | Use ORC for Hive-centric pipelines |

### Prerequisites

- **Glue Data Catalog** database and table with the schema definition
- Table must define column names, types, and (optionally) partition keys
- Firehose IAM role must have `glue:GetTable` and `glue:GetTableVersions` permissions
- Schema changes require updating the Glue table (Firehose reads the latest version by default)

## Dynamic Partitioning

### How Partition Keys Work

Dynamic partitioning extracts key-value pairs from each record and uses them to construct unique S3 prefixes, creating a partitioned directory layout for efficient querying.

**Two extraction methods:**

1. **From JQ expressions** — Inline JQ expressions applied to JSON records (configured via S3 prefix expressions)
2. **From Lambda metadata** — Partition keys returned by a Lambda transformation function

### Prefix Expressions

The S3 prefix uses `!{partitionKeyFromQuery:key}` or `!{partitionKeyFromLambda:key}` syntax:

```
data/region=!{partitionKeyFromLambda:region}/customer=!{partitionKeyFromLambda:customer_id}/year=!{timestamp:yyyy}/month=!{timestamp:MM}/
```

This produces a directory layout like:

```
data/
  region=us-east-1/
    customer=cust-123/
      year=2026/
        month=02/
          file001.parquet
          file002.parquet
    customer=cust-456/
      year=2026/
        month=02/
          file003.parquet
  region=eu-west-1/
    customer=cust-789/
      ...
```

### Important Considerations

- Dynamic partitioning is **ForceNew** — it cannot be enabled or disabled after the delivery stream is created
- Each unique partition key combination creates a separate S3 prefix (and separate buffer)
- High cardinality partition keys (e.g., user_id with millions of values) create many small files — use buffering hints to mitigate
- Retry duration controls how long Firehose retries when a partition write fails (0–7200s, default 300)

## Encryption Model

### Server-Side Encryption (SSE) for Direct PUT

When using Direct PUT as the source, Firehose can encrypt data at rest in its internal buffer:

| Configuration | Encryption | Cost |
|---------------|-----------|------|
| `sseEnabled: false` | No encryption (plaintext in buffer) | No additional cost |
| `sseEnabled: true`, no KMS key | AWS-owned CMK (Firehose manages the key) | No additional cost |
| `sseEnabled: true`, with `sseKmsKeyArn` | Customer-managed CMK | KMS API charges per encryption/decryption |

### Encryption with Kinesis Stream Source

When using a Kinesis Data Stream as the source, **do not enable SSE on the delivery stream**. The source stream handles encryption:

- The Kinesis stream has its own KMS encryption configuration
- Firehose reads already-encrypted records and the stream's KMS key handles decryption
- Enabling SSE on the delivery stream would be redundant and is rejected by the API

### Encryption at S3 Destination

S3 encryption is configured separately in the destination's `kmsKeyArn` field:

- When absent, S3 uses its default encryption (SSE-S3 or bucket default)
- When set, S3 uses SSE-KMS with the specified customer-managed key
- This is independent of the delivery stream's SSE setting

## S3 Backup Behavior

Every non-S3 destination requires S3 backup for error handling. The behavior varies by destination:

| Destination | S3 Role | Backup Modes | Default |
|-------------|---------|-------------|---------|
| **Extended S3** | Primary destination | `Disabled` / `Enabled` (source record backup) | `Disabled` |
| **OpenSearch** | Backup for failed/all documents | `FailedDocumentsOnly` / `AllDocuments` | `FailedDocumentsOnly` |
| **HTTP Endpoint** | Backup for failed/all records | `FailedDataOnly` / `AllData` | `FailedDataOnly` |
| **Redshift** | Staging for COPY + optional source backup | `Disabled` / `Enabled` (source record backup) | `Disabled` |

### Extended S3 Backup

For Extended S3 destinations, "backup" means keeping a copy of the **original, pre-transformation** records. This is separate from the primary S3 delivery:

- **Primary delivery** → transformed/converted records (e.g., Parquet with enrichments)
- **Source backup** → original JSON records as received (for auditing/reprocessing)

### Redshift S3 Staging

The Redshift destination uses S3 as an **intermediate staging area** (not backup). Firehose writes data to S3, then issues a `COPY` command. This S3 data is the primary delivery path. Optionally, you can also enable source record backup to a separate S3 location.

## Cost Model

Firehose pricing is straightforward — pay per GB of data ingested. No upfront costs, no provisioning, no idle charges.

### Base Pricing (US East)

| Component | Price |
|-----------|-------|
| Data ingested (first 500 TB/month) | ~$0.029/GB |
| Data ingested (next 1.5 PB/month) | ~$0.025/GB |
| Data ingested (over 2 PB/month) | ~$0.020/GB |
| Format conversion (Parquet/ORC) | ~$0.018/GB |
| Dynamic partitioning | ~$0.020/GB |
| VPC delivery (per hour per AZ) | ~$0.01/hour/AZ |

### What Counts as "Ingested"

- Firehose rounds each record up to the nearest **5 KB** for billing purposes
- A 100-byte record is billed as 5 KB
- A 6 KB record is billed as 10 KB
- Records under 5 KB should be batched client-side for cost efficiency

### Additional Costs

- **Lambda transformation**: Standard Lambda pricing (invocations + duration)
- **S3 storage**: Standard S3 pricing for delivered and backup objects
- **KMS encryption**: Per-API-call KMS charges when using customer-managed keys
- **CloudWatch Logs**: Standard CloudWatch Logs pricing for error logs
- **Glue Data Catalog**: Free for the first million objects; standard pricing after

### Cost Optimization Tips

- Batch records client-side to approach 1 MiB per record (minimize 5 KB rounding overhead)
- Use GZIP compression to reduce delivered data volume (destination storage cost)
- Set appropriate buffer sizes — larger buffers reduce the number of S3 PUT operations
- Use format conversion (Parquet) for analytics workloads — storage savings often outweigh the conversion cost

## Security

### IAM Roles

Firehose requires IAM roles for every interaction with other AWS services. A single role can cover multiple permissions, or you can use separate roles for fine-grained control:

| Permission | Actions | Used By |
|------------|---------|---------|
| S3 write | `s3:PutObject`, `s3:AbortMultipartUpload`, `s3:GetBucketLocation`, `s3:ListBucket` | All destinations |
| Kinesis read | `kinesis:GetRecords`, `kinesis:GetShardIterator`, `kinesis:DescribeStream`, `kinesis:ListShards` | Kinesis source |
| Lambda invoke | `lambda:InvokeFunction`, `lambda:GetFunctionConfiguration` | Processing |
| OpenSearch write | `es:ESHttpPut`, `es:ESHttpGet` | OpenSearch destination |
| KMS encrypt/decrypt | `kms:Encrypt`, `kms:Decrypt`, `kms:GenerateDataKey` | SSE, S3 SSE-KMS |
| Glue catalog | `glue:GetTable`, `glue:GetTableVersions` | Format conversion |
| VPC ENI management | `ec2:CreateNetworkInterface`, `ec2:DescribeNetworkInterfaces`, `ec2:DeleteNetworkInterface` | VPC delivery |
| CloudWatch Logs | `logs:PutLogEvents` | Error logging |

### VPC Delivery

For OpenSearch domains deployed in a VPC, Firehose creates Elastic Network Interfaces (ENIs) in the specified subnets:

- Provide subnets in multiple AZs for high availability
- Security groups must allow outbound HTTPS (port 443) to the OpenSearch domain
- The VPC configuration is **ForceNew** — changing subnets or security groups replaces the delivery stream
- VPC delivery adds ~$0.01/hour per AZ to the cost

### Encryption at Rest

Three layers of encryption are available:

1. **Delivery stream buffer** (SSE) — Encrypts data in Firehose's internal buffer (Direct PUT only)
2. **S3 destination** (SSE-KMS) — Encrypts objects written to S3
3. **Source stream** (KMS) — Encrypts data in the Kinesis Data Stream (Kinesis source only)

All three are independent and can use different KMS keys.

## Limits and Quotas

### Delivery Stream Limits

| Limit | Value | Adjustable |
|-------|-------|-----------|
| Delivery streams per region | 500 | Yes (request increase) |
| Record size | 1 MiB maximum | No |
| `PutRecordBatch` batch size | 500 records or 4 MiB | No |
| `PutRecord` throughput (Direct PUT) | 1,000 records/s or 1 MB/s | Yes |
| `PutRecordBatch` throughput (Direct PUT) | 4,000 records/s or 4 MB/s | Yes |
| Lambda processing buffer | 1–3 MiB | No |
| Lambda processing timeout | 5 minutes | No |
| Lambda retries | 0–300 | No |
| Buffer interval | 0–900 seconds | No |
| Buffer size | 1–128 MiB (destination-dependent) | No |
| Dynamic partitioning active partitions | 500 per delivery stream | Yes |
| Retry duration (delivery) | 0–7200 seconds | No |

### Destination-Specific Limits

| Destination | Max Buffer Size | Notes |
|-------------|----------------|-------|
| Extended S3 | 128 MiB | — |
| OpenSearch | 100 MiB | — |
| HTTP Endpoint | 64 MiB | Endpoint must respond within 3 minutes |
| Redshift | N/A | Uses S3 staging; COPY command has separate limits |

## Firehose vs Kinesis Data Streams vs SQS/SNS

### When to Use Each

| Service | Model | Best For |
|---------|-------|----------|
| **Firehose** | Managed ETL pipeline | Zero-code delivery to S3/OpenSearch/HTTP/Redshift |
| **Kinesis Data Streams** | Streaming log (pull) | Real-time processing, multiple consumers, replay |
| **SQS** | Message queue (pull) | Task distribution, decoupling, exactly-once processing |
| **SNS** | Pub/sub (push) | Fan-out notifications, multi-subscriber broadcasting |

### Detailed Comparison

| Feature | Firehose | Kinesis Streams | SQS | SNS |
|---------|----------|----------------|-----|-----|
| Consumer code | None | You write it | You write it | You configure subscribers |
| Destinations | S3, OpenSearch, HTTP, Redshift | Anything (custom code) | Anything (custom code) | Lambda, SQS, HTTP, email, SMS |
| Ordering | None (or per-shard via Kinesis source) | Per-shard | FIFO or best-effort | None |
| Replay | No | Yes (retention-based) | No (once consumed) | No |
| Throughput | Auto-scaling | Per-shard or ON_DEMAND | Auto-scaling | Auto-scaling |
| Latency | Seconds to minutes (buffer-dependent) | ~200ms (poll) / ~70ms (EFO) | ~1ms | ~1ms |
| Retention | None (delivers immediately) | 24h–365 days | 1min–14 days | None |
| Transformation | Lambda (built-in) | Custom consumer code | Custom consumer code | None |
| Format conversion | Parquet/ORC (built-in) | Custom consumer code | Custom consumer code | None |
| Cost model | Per-GB ingested | Per-shard-hour or per-GB | Per-request | Per-publish + per-delivery |

### Common Architectures

**Simple data lake:**
```
Application → Firehose (Direct PUT) → S3
```

**Real-time + archive:**
```
Application → Kinesis Stream → Firehose → S3 (archive)
                             → Lambda → DynamoDB (real-time)
                             → Custom app (analytics)
```

**Multi-destination fan-out:**
```
Application → Kinesis Stream → Firehose #1 → S3 (data lake)
                             → Firehose #2 → OpenSearch (logs)
                             → Firehose #3 → Datadog (monitoring)
```

**Warehouse pipeline:**
```
Application → Firehose (Direct PUT) → Redshift (via S3 staging)
```

## Operational Best Practices

### Monitoring

Key CloudWatch metrics to monitor:

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `DeliveryToS3.Success` | Percentage of successful S3 deliveries | < 99% |
| `DeliveryToS3.Records` | Number of records delivered | Anomaly detection |
| `IncomingBytes` | Data volume entering the stream | Capacity planning |
| `IncomingRecords` | Record count entering the stream | Throughput analysis |
| `ThrottledRecords` | Records rejected due to throttling | > 0 sustained |
| `FailedConversion.Bytes` | Data that failed format conversion | > 0 |
| `ExecuteProcessing.Duration` | Lambda processing time | Approaching 5min timeout |

### Error Handling

- Always configure CloudWatch logging for production delivery streams
- Monitor the error output prefix in S3 — records here indicate transformation or delivery failures
- For OpenSearch destinations, check the S3 backup bucket for rejected documents (index mapping conflicts, field type mismatches)
- For HTTP endpoints, check the S3 backup for non-2xx responses from the endpoint

### Naming Conventions

- Delivery stream names are unique per AWS account per region
- Names are **ForceNew** — choose carefully at creation time
- 1–64 characters, allowed: letters, digits, hyphens, underscores, periods
- Convention: `{service}-{purpose}-{env}` (e.g., `clickstream-s3-prod`)
