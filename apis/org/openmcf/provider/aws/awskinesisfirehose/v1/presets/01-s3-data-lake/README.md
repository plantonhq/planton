# S3 Data Lake Preset

Minimal Extended S3 destination for general-purpose log and event storage. Uses Direct PUT source with GZIP compression and time-based partitioning.

## When to Use

- **General-purpose log storage** — Application logs, audit trails, clickstream data
- **Event archival** — Event sourcing, analytics pipelines, data lake ingestion
- **Cost-effective retention** — Long-term storage with S3 lifecycle policies
- **Downstream processing** — Data ready for Athena, Glue, Spark, or custom ETL

## Key Configuration

- **Direct PUT source** — No Kinesis stream required; applications call PutRecord/PutRecordBatch directly
- **GZIP compression** — Reduces storage cost and transfer size
- **120s buffering interval, 64 MiB buffer** — Balances latency with batching efficiency
- **Year/month/day partitioning** — `events/year=YYYY/month=MM/day=DD/` for efficient querying
- **File extension** — `.json.gz` for compressed JSON objects

## Prerequisites

| Resource | Description |
|----------|-------------|
| **S3 bucket** | Target bucket for delivered objects. Create with appropriate lifecycle and encryption. |
| **IAM role** | Role with `s3:PutObject`, `s3:AbortMultipartUpload`, `s3:GetBucketLocation`, `s3:ListBucket` on the bucket. |

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `my-data-lake-bucket` | Name of your S3 data lake bucket |
| `123456789012` | Your AWS account ID |
| `firehose-s3-delivery-role` | IAM role name for Firehose delivery |
