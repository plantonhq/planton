# AwsKinesisFirehose

Deploys an [Amazon Kinesis Data Firehose](https://docs.aws.amazon.com/firehose/latest/dev/what-is-this-service.html) delivery stream — a fully managed service for loading streaming data into storage and analytics destinations without writing any custom consumer code.

Firehose captures data from Direct PUT calls or an existing Kinesis Data Stream, buffers it, optionally transforms and converts the format, then delivers it to a destination (S3, OpenSearch, HTTP endpoint, or Redshift). It handles retries, back-pressure, and error routing automatically.

## When to Use

Use Kinesis Data Firehose when you need:

- **Zero-code delivery to S3** — Stream data into a data lake with automatic batching, compression, and optional Parquet/ORC conversion
- **Managed ETL pipeline** — Transform records with Lambda, convert formats via Glue, and partition by record fields — all without writing consumer infrastructure
- **Third-party integrations** — Deliver to Datadog, New Relic, Sumo Logic, Splunk, or any HTTPS endpoint with built-in retry and S3 backup
- **Redshift warehouse loading** — Stage data in S3 and issue COPY commands automatically
- **Log analytics** — Index directly into OpenSearch with configurable rotation and VPC delivery

### Firehose vs Alternatives

| Approach | Best For | Trade-offs |
|----------|----------|------------|
| **Firehose** | Zero-ops delivery to S3/OpenSearch/HTTP/Redshift | Limited to supported destinations, max 1 MiB/record, higher per-GB cost |
| **Kinesis Data Streams + custom consumer** | Complex processing, multiple outputs, replay | You manage the consumer (scaling, checkpointing, error handling) |
| **S3 batch uploads** | Periodic bulk loads, non-real-time | No streaming, higher latency (minutes to hours) |
| **Direct API calls** | Low-volume, synchronous writes | No batching, no retry, no transformation |

**Choose Firehose** when you want fully managed delivery with zero consumer code. **Choose Kinesis Data Streams** when you need replay, multiple independent consumers, or custom processing logic. **Choose batch uploads** when latency doesn't matter.

## Prerequisites

- AWS account with permissions to create Firehose delivery streams
- **Destination resources must exist** before the delivery stream:
  - Extended S3: S3 bucket (see `AwsS3Bucket`)
  - OpenSearch: OpenSearch domain (see `AwsOpenSearchDomain`)
  - HTTP endpoint: HTTPS endpoint accepting POST requests
  - Redshift: Redshift cluster with target database and table
- **IAM roles** granting Firehose access to destinations, S3, Lambda, KMS, and Glue (see `AwsIamRole`)
- (Optional) Kinesis Data Stream as source (see `AwsKinesisStream`)
- (Optional) Lambda function for transformation (see `AwsLambda`)
- (Optional) AWS Glue catalog table for format conversion
- (Optional) KMS key for encryption (see `AwsKmsKey`)

## Spec Reference

### Source Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `kinesis_stream_source` | object | No | — | Kinesis Data Stream source. When absent, Direct PUT is used. **ForceNew** — entire source config. |
| `kinesis_stream_source.stream_arn` | StringValueOrRef | **Yes** (if source set) | — | ARN of the Kinesis stream to read from |
| `kinesis_stream_source.role_arn` | StringValueOrRef | **Yes** (if source set) | — | IAM role for Firehose to read from the stream |

### Server-Side Encryption (Direct PUT only)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `sse_enabled` | bool | No | false | Enable SSE for data at rest. Only valid for Direct PUT sources. |
| `sse_kms_key_arn` | StringValueOrRef | No | — | Customer-managed KMS key ARN. When absent and SSE enabled, uses AWS-owned CMK. |

### Destination (exactly one required, ForceNew)

Exactly one destination must be configured. The destination type is **ForceNew** — changing it replaces the entire delivery stream.

#### Extended S3 Destination (`extended_s3`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `bucket_arn` | StringValueOrRef | **Yes** | — | S3 bucket ARN for record delivery |
| `role_arn` | StringValueOrRef | **Yes** | — | IAM role for S3, KMS, Lambda, Glue access |
| `prefix` | string | No | — | S3 key prefix (supports Firehose expressions) |
| `error_output_prefix` | string | No | — | S3 prefix for failed records |
| `compression_format` | string | No | `"UNCOMPRESSED"` | `"UNCOMPRESSED"`, `"GZIP"`, `"ZIP"`, `"Snappy"`, `"HADOOP_SNAPPY"` |
| `kms_key_arn` | StringValueOrRef | No | — | KMS key for S3 SSE-KMS encryption |
| `buffering` | object | No | 300s / 5 MiB | Buffering hints (see below) |
| `custom_time_zone` | string | No | `"UTC"` | IANA time zone for prefix timestamps |
| `file_extension` | string | No | — | File extension for objects (e.g., `".json"`, `".parquet"`) |
| `s3_backup_mode` | string | No | `"Disabled"` | `"Disabled"` or `"Enabled"` — backup original records |
| `s3_backup` | object | No | — | S3 config for source record backup (required when mode is `"Enabled"`) |
| `processing` | object | No | — | Lambda transformation (see below) |
| `logging` | object | No | — | CloudWatch error logging (see below) |
| `dynamic_partitioning` | object | No | — | Dynamic partitioning config. **ForceNew**. |
| `data_format_conversion` | object | No | — | JSON → Parquet/ORC conversion via Glue catalog |

#### OpenSearch Destination (`opensearch`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `domain_arn` | StringValueOrRef | Conditional | — | OpenSearch domain ARN (mutually exclusive with `cluster_endpoint`) |
| `cluster_endpoint` | string | Conditional | — | Cluster endpoint URL (mutually exclusive with `domain_arn`) |
| `index_name` | string | **Yes** | — | Target index name (becomes prefix with rotation) |
| `role_arn` | StringValueOrRef | **Yes** | — | IAM role for OpenSearch write access |
| `index_rotation_period` | string | No | `"OneDay"` | `"NoRotation"`, `"OneHour"`, `"OneDay"`, `"OneWeek"`, `"OneMonth"` |
| `type_name` | string | No | — | Document type (ES 6.x only, leave empty for OpenSearch) |
| `buffering` | object | No | 300s / 5 MiB | Buffering hints (max 100 MiB for OpenSearch) |
| `retry_duration_in_seconds` | int32 | No | 300 | Retry duration: 0–7200s |
| `s3_backup_mode` | string | No | `"FailedDocumentsOnly"` | `"FailedDocumentsOnly"` or `"AllDocuments"` |
| `s3_config` | object | **Yes** | — | S3 backup for failed/all documents |
| `processing` | object | No | — | Lambda transformation |
| `logging` | object | No | — | CloudWatch error logging |
| `vpc_config` | object | No | — | VPC delivery config. **ForceNew**. |

#### HTTP Endpoint Destination (`http_endpoint`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `url` | string | **Yes** | — | HTTPS endpoint URL (must start with `https://`) |
| `name` | string | No | — | Human-readable endpoint name (AWS Console / metrics) |
| `access_key` | string | No | — | Authentication key (sent in `X-Amz-Firehose-Access-Key` header) |
| `role_arn` | StringValueOrRef | No | — | IAM role for endpoint delivery and S3 backup |
| `buffering` | object | No | 300s / 5 MiB | Buffering hints (max 64 MiB for HTTP) |
| `retry_duration_in_seconds` | int32 | No | 300 | Retry duration: 0–7200s |
| `s3_backup_mode` | string | No | `"FailedDataOnly"` | `"FailedDataOnly"` or `"AllData"` |
| `s3_config` | object | **Yes** | — | S3 backup for failed/all records |
| `processing` | object | No | — | Lambda transformation |
| `logging` | object | No | — | CloudWatch error logging |
| `request_config` | object | No | — | Content encoding and custom attributes |

#### Redshift Destination (`redshift`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `cluster_jdbcurl` | string | **Yes** | — | JDBC URL (`jdbc:redshift://<endpoint>:<port>/<db>`) |
| `role_arn` | StringValueOrRef | **Yes** | — | IAM role for COPY and S3 staging access |
| `data_table_name` | string | **Yes** | — | Target table name for COPY command |
| `data_table_columns` | string | No | — | Comma-separated column list for COPY |
| `copy_options` | string | No | — | Additional COPY options (e.g., `"JSON 'auto'"`) |
| `username` | string | No | — | Redshift database username |
| `password` | StringValueOrRef | No | — | Redshift database password (sensitive) |
| `s3_config` | object | **Yes** | — | S3 intermediate staging bucket |
| `retry_duration_in_seconds` | int32 | No | 3600 | Retry duration: 0–7200s |
| `s3_backup_mode` | string | No | `"Disabled"` | `"Disabled"` or `"Enabled"` |
| `s3_backup` | object | No | — | S3 backup for source records |
| `processing` | object | No | — | Lambda transformation |
| `logging` | object | No | — | CloudWatch error logging |

### Shared Sub-Messages

#### Buffering Hints (`buffering`)

| Field | Type | Range | Default | Description |
|-------|------|-------|---------|-------------|
| `interval_in_seconds` | int32 | 0–900 | 300 | Flush interval. Lower = less latency; higher = fewer objects. |
| `size_in_mbs` | int32 | 1–128 | 5 | Flush threshold in MiB. Destination-specific max applies. |

Firehose delivers when **either** threshold is reached — whichever comes first.

#### Lambda Processing (`processing`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | **Yes** | false | Enable Lambda transformation |
| `lambda_arn` | StringValueOrRef | Conditional | — | Lambda function ARN (required when enabled) |
| `buffer_size_in_mbs` | int32 | No | 3 | Lambda invocation buffer: 1–3 MiB |
| `buffer_interval_in_seconds` | int32 | No | 60 | Lambda invocation interval: 60–900s |
| `number_of_retries` | int32 | No | 3 | Lambda retry count: 0–300 |

#### CloudWatch Logging (`logging`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | **Yes** | false | Enable error logging |
| `log_group_name` | string | Conditional | — | Log group (required when enabled) |
| `log_stream_name` | string | Conditional | — | Log stream (required when enabled) |

#### S3 Config (`s3_config` / `s3_backup`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `bucket_arn` | StringValueOrRef | **Yes** | — | S3 bucket ARN |
| `role_arn` | StringValueOrRef | **Yes** | — | IAM role for S3 write access |
| `prefix` | string | No | — | S3 key prefix |
| `error_output_prefix` | string | No | — | S3 prefix for errors |
| `compression_format` | string | No | `"UNCOMPRESSED"` | Compression: `"UNCOMPRESSED"`, `"GZIP"`, `"ZIP"`, `"Snappy"`, `"HADOOP_SNAPPY"` |
| `kms_key_arn` | StringValueOrRef | No | — | KMS key for SSE-KMS |
| `buffering` | object | No | — | Buffering hints |
| `logging` | object | No | — | CloudWatch logging |

#### VPC Config (`vpc_config`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `subnet_ids` | repeated StringValueOrRef | **Yes** | — | Subnet IDs for ENIs (multi-AZ recommended) |
| `security_group_ids` | repeated StringValueOrRef | **Yes** | — | Security group IDs (allow outbound HTTPS 443) |
| `role_arn` | StringValueOrRef | **Yes** | — | IAM role for ENI management |

#### Dynamic Partitioning (`dynamic_partitioning`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | **Yes** | false | Enable dynamic partitioning. **ForceNew**. |
| `retry_duration_in_seconds` | int32 | No | 300 | Retry duration: 0–7200s |

#### Data Format Conversion (`data_format_conversion`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | **Yes** | false | Enable JSON → columnar conversion |
| `input_format` | string | No | `"OPENX_JSON"` | `"OPENX_JSON"` or `"HIVE_JSON"` |
| `output_format` | string | Conditional | — | `"PARQUET"` or `"ORC"` (required when enabled) |
| `parquet_compression` | string | No | `"SNAPPY"` | `"SNAPPY"`, `"GZIP"`, `"UNCOMPRESSED"` (Parquet only) |
| `orc_compression` | string | No | `"SNAPPY"` | `"SNAPPY"`, `"ZLIB"`, `"NONE"` (ORC only) |
| `schema` | object | Conditional | — | Glue catalog schema (required when enabled) |

#### Glue Schema Config (`data_format_conversion.schema`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `database_name` | string | **Yes** | — | Glue catalog database name |
| `table_name` | string | **Yes** | — | Glue catalog table name |
| `role_arn` | StringValueOrRef | **Yes** | — | IAM role for Glue catalog access |
| `catalog_id` | string | No | — | Glue catalog ID (defaults to current AWS account) |
| `region` | string | No | — | Glue catalog region (defaults to stream region) |
| `version_id` | string | No | `"LATEST"` | Table version |

#### HTTP Request Config (`request_config`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `content_encoding` | string | No | `"NONE"` | `"NONE"` or `"GZIP"` |
| `common_attributes` | repeated object | No | — | Custom key-value pairs sent as HTTP headers |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `delivery_stream_arn` | ARN of the delivery stream (used by IAM policies, CloudWatch alarms, event sources) |
| `delivery_stream_name` | Name of the delivery stream (used for PutRecord/PutRecordBatch API calls) |

## ForceNew Fields

The following fields require replacing the entire delivery stream if changed:

| Field | Reason |
|-------|--------|
| `metadata.name` | Delivery stream name is immutable |
| Destination type (`extended_s3` vs `opensearch` vs `http_endpoint` vs `redshift`) | Cannot switch destination types |
| `kinesis_stream_source` (all fields) | Source type and configuration are immutable |
| `dynamic_partitioning.enabled` | Cannot enable/disable after creation |
| `vpc_config` (all fields) | VPC ENI configuration is immutable |

## v1 Scope

### Supported Destinations

1. **Extended S3** — Data lake storage with compression, Lambda transformation, dynamic partitioning, Parquet/ORC format conversion
2. **OpenSearch** — Direct indexing with rotation, VPC delivery, S3 backup
3. **HTTP Endpoint** — Generic HTTPS delivery (Datadog, New Relic, Sumo Logic, custom APIs)
4. **Redshift** — Data warehouse loading via S3 staging + COPY command

### Supported Sources

1. **Direct PUT** (default) — Applications call PutRecord/PutRecordBatch APIs
2. **Kinesis Data Stream** — Firehose reads from an existing stream with automatic checkpointing

### Deliberate v1 Omissions

| Feature | Reason |
|---------|--------|
| Splunk destination | Requires HEC token management and Splunk-specific config. <15% usage. |
| Amazon OpenSearch Serverless destination | Separate destination type in AWS API. Can add in v2. |
| MSK (Managed Kafka) source | Requires Kafka-specific consumer group config. Can add in v2. |
| Secrets Manager for Redshift credentials | Additional dependency; direct credentials cover most use cases in v1. |
| Multiple processors (AppendDelimiter, MetadataExtraction) | Lambda covers 95%+ of transformation needs. |
| Custom prefix expressions via spec | Prefix expressions are set directly in the `prefix` string field. |

## Related Resources

- **AwsS3Bucket** — Destination bucket, backup bucket, or Redshift staging bucket
- **AwsKinesisStream** — Source stream when using Kinesis source mode
- **AwsIamRole** — Roles for Firehose to access destinations, S3, KMS, Lambda, Glue
- **AwsKmsKey** — Encryption key for SSE or S3 SSE-KMS
- **AwsLambda** — Transformation function for record processing
- **AwsOpenSearchDomain** — Target domain for OpenSearch destination
- **AwsCloudwatchAlarm** — Monitor delivery metrics (DeliveryToS3.Success, IncomingBytes)
