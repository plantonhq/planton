# AwsKinesisFirehose

Deploys an [Amazon Kinesis Data Firehose](https://docs.aws.amazon.com/firehose/latest/dev/what-is-this-service.html) delivery stream — a fully managed service for loading streaming data into storage and analytics destinations without writing any custom consumer code.

Firehose captures data from Direct PUT calls, an existing Kinesis Data Stream, or an Amazon MSK topic; buffers it; optionally transforms records and converts formats; then delivers it to exactly one destination. It handles retries, back-pressure, and error routing automatically.

## When to Use

Use Kinesis Data Firehose when you need:

- **Zero-code delivery to S3** — Stream data into a data lake with automatic batching, compression, and optional Parquet/ORC conversion
- **Managed ETL pipeline** — Transform records with Lambda, extract partition keys with JQ, convert formats via Glue, and partition by record fields — all without writing consumer infrastructure
- **Third-party integrations** — Deliver to Splunk, Datadog, New Relic, Sumo Logic, or any HTTPS endpoint with built-in retry and S3 backup
- **Warehouse and lakehouse loading** — COPY into Redshift, stream into Snowflake via Snowpipe Streaming, or commit directly into Glue-cataloged Apache Iceberg tables
- **Log analytics** — Index directly into OpenSearch domains or OpenSearch Serverless collections with configurable rotation and VPC delivery
- **Kafka offload** — Turn an MSK topic into a delivery pipeline with zero consumer code

### Firehose vs Alternatives

| Approach | Best For | Trade-offs |
|----------|----------|------------|
| **Firehose** | Zero-ops delivery to the eight supported destinations | Max 1 MiB/record, higher per-GB cost |
| **Kinesis Data Streams + custom consumer** | Complex processing, multiple outputs, replay | You manage the consumer (scaling, checkpointing, error handling) |
| **S3 batch uploads** | Periodic bulk loads, non-real-time | No streaming, higher latency (minutes to hours) |
| **Direct API calls** | Low-volume, synchronous writes | No batching, no retry, no transformation |

**Choose Firehose** when you want fully managed delivery with zero consumer code. **Choose Kinesis Data Streams** when you need replay, multiple independent consumers, or custom processing logic. **Choose batch uploads** when latency doesn't matter.

## Prerequisites

- AWS account with permissions to create Firehose delivery streams
- **Destination resources must exist** before the delivery stream:
  - Extended S3: S3 bucket (see `AwsS3Bucket`)
  - OpenSearch: OpenSearch domain (see `AwsOpenSearchDomain`)
  - OpenSearch Serverless: collection endpoint with a data access policy admitting the delivery role
  - HTTP endpoint: HTTPS endpoint accepting POST requests
  - Redshift: Redshift cluster with target database and table
  - Splunk: HEC endpoint with a minted token
  - Snowflake: account, database/schema/table, and a key-pair-enabled user
  - Iceberg: Glue Data Catalog tables in Iceberg format
- **IAM roles** granting Firehose access to destinations, S3, Lambda, KMS, and Glue (see `AwsIamRole`)
- (Optional) Kinesis Data Stream (see `AwsKinesisStream`) or MSK cluster (see `AwsMskCluster`) as source
- (Optional) Lambda function for transformation (see `AwsLambda`)
- (Optional) AWS Glue catalog table for format conversion
- (Optional) KMS key for encryption (see `AwsKmsKey`)
- (Optional) Secrets Manager secret for destination credentials

## Spec Reference

### Source Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `kinesis_stream_source` | object | No | — | Kinesis Data Stream source. Mutually exclusive with `msk_source`. When both absent, Direct PUT is used. **ForceNew** — entire source config. |
| `kinesis_stream_source.stream_arn` | StringValueOrRef | **Yes** (if source set) | — | ARN of the Kinesis stream to read from |
| `kinesis_stream_source.role_arn` | StringValueOrRef | **Yes** (if source set) | — | IAM role for Firehose to read from the stream |
| `msk_source` | object | No | — | Amazon MSK source. Mutually exclusive with `kinesis_stream_source`. **ForceNew** — entire source config. |
| `msk_source.msk_cluster_arn` | StringValueOrRef | **Yes** (if source set) | — | ARN of the MSK cluster (IAM access control must be enabled) |
| `msk_source.topic_name` | string | **Yes** (if source set) | — | Kafka topic to read from |
| `msk_source.connectivity` | string | **Yes** (if source set) | — | `"PRIVATE"` (in-VPC brokers) or `"PUBLIC"` |
| `msk_source.role_arn` | StringValueOrRef | **Yes** (if source set) | — | IAM role with kafka + kafka-cluster data-plane permissions |
| `msk_source.read_from_timestamp` | string | No | latest offset | RFC 3339 point in time to start reading from |

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
| `processing` | object | No | — | Record-transformation pipeline (see below) |
| `logging` | object | No | — | CloudWatch error logging (see below) |
| `dynamic_partitioning` | object | No | — | Dynamic partitioning config. **ForceNew**. Define partition keys with a `metadata_extraction` processor. |
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
| `default_document_id_format` | string | No | `"FIREHOSE_DEFAULT"` | `"FIREHOSE_DEFAULT"` (dedupe-safe retries) or `"NO_DOCUMENT_ID"` (higher throughput) |
| `buffering` | object | No | 300s / 5 MiB | Buffering hints (max 100 MiB for OpenSearch) |
| `retry_duration_in_seconds` | int32 | No | 300 | Retry duration: 0–7200s |
| `s3_backup_mode` | string | No | `"FailedDocumentsOnly"` | `"FailedDocumentsOnly"` or `"AllDocuments"`. **ForceNew**. |
| `s3_config` | object | **Yes** | — | S3 backup for failed/all documents |
| `processing` | object | No | — | Record-transformation pipeline |
| `logging` | object | No | — | CloudWatch error logging |
| `vpc_config` | object | No | — | VPC delivery config. **ForceNew**. |

#### OpenSearch Serverless Destination (`opensearch_serverless`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `collection_endpoint` | string | **Yes** | — | Collection endpoint (`https://<id>.<region>.aoss.amazonaws.com`) |
| `index_name` | string | **Yes** | — | Target index (must be admitted by the collection's data access policy) |
| `role_arn` | StringValueOrRef | **Yes** | — | IAM role with `aoss:APIAccessAll` + data-access-policy write grant |
| `buffering` | object | No | 300s / 5 MiB | Buffering hints (max 100 MiB) |
| `retry_duration_in_seconds` | int32 | No | 300 | Retry duration: 0–7200s |
| `s3_backup_mode` | string | No | `"FailedDocumentsOnly"` | `"FailedDocumentsOnly"` or `"AllDocuments"`. **ForceNew**. |
| `s3_config` | object | **Yes** | — | S3 backup for failed/all documents |
| `processing` | object | No | — | Record-transformation pipeline |
| `logging` | object | No | — | CloudWatch error logging |
| `vpc_config` | object | No | — | VPC delivery config. **ForceNew**. |

#### HTTP Endpoint Destination (`http_endpoint`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `url` | string | **Yes** | — | HTTPS endpoint URL (must start with `https://`) |
| `name` | string | No | — | Human-readable endpoint name (AWS Console / metrics) |
| `access_key` | string | No | — | Authentication key (sensitive; sent in `X-Amz-Firehose-Access-Key` header). Mutually exclusive with `secrets_manager`. |
| `secrets_manager` | object | No | — | Source the access key from Secrets Manager (secret shape `{"api_key": "..."}`) — the recommended production mode |
| `role_arn` | StringValueOrRef | No | — | IAM role for endpoint delivery and S3 backup (the S3 config carries its own role) |
| `buffering` | object | No | 300s / 5 MiB | Buffering hints (max 100 MiB) |
| `retry_duration_in_seconds` | int32 | No | 300 | Retry duration: 0–7200s |
| `s3_backup_mode` | string | No | `"FailedDataOnly"` | `"FailedDataOnly"` or `"AllData"` |
| `s3_config` | object | **Yes** | — | S3 backup for failed/all records |
| `processing` | object | No | — | Record-transformation pipeline |
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
| `username` | string | Conditional | — | Redshift database username (with `password`; exactly one auth mode) |
| `password` | string | Conditional | — | Redshift database password (sensitive). Prefer `secrets_manager`. |
| `secrets_manager` | object | Conditional | — | Source credentials from Secrets Manager (secret shape `{"username": "...", "password": "..."}`) — the recommended production mode |
| `s3_config` | object | **Yes** | — | S3 intermediate staging bucket |
| `retry_duration_in_seconds` | int32 | No | 3600 | Retry duration: 0–7200s |
| `s3_backup_mode` | string | No | `"Disabled"` | `"Disabled"` or `"Enabled"` |
| `s3_backup` | object | No | — | S3 backup for source records |
| `processing` | object | No | — | Record-transformation pipeline |
| `logging` | object | No | — | CloudWatch error logging |

#### Splunk Destination (`splunk`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `hec_endpoint` | string | **Yes** | — | HEC endpoint URL including port (must start with `https://`) |
| `hec_endpoint_type` | string | No | `"Raw"` | `"Raw"` (preformatted events) or `"Event"` (Splunk event JSON) |
| `hec_token` | string | Conditional | — | HEC token (sensitive). Exactly one auth mode with `secrets_manager`. |
| `secrets_manager` | object | Conditional | — | Source the token from Secrets Manager (secret shape `{"hec_token": "..."}`) — the recommended production mode |
| `hec_acknowledgment_timeout_in_seconds` | int32 | No | 180 | Indexer ack wait: 180–600s |
| `buffering` | object | No | 60s / 5 MiB | Buffering hints — Splunk caps: 0–60s / 1–5 MiB |
| `retry_duration_in_seconds` | int32 | No | 3600 | Retry duration: 0–7200s |
| `s3_backup_mode` | string | No | `"FailedEventsOnly"` | `"FailedEventsOnly"` or `"AllEvents"` |
| `s3_config` | object | **Yes** | — | S3 backup for failed/all events |
| `processing` | object | No | — | Record-transformation pipeline |
| `logging` | object | No | — | CloudWatch error logging |

Note: Splunk has no destination-level `role_arn` — HEC authorization is the token, and the S3 configuration's role carries the backup permissions.

#### Snowflake Destination (`snowflake`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `account_url` | string | **Yes** | — | `https://<account-identifier>.snowflakecomputing.com` |
| `database` / `schema` / `table` | string | **Yes** | — | Target table coordinates |
| `role_arn` | StringValueOrRef | **Yes** | — | IAM role for S3 backup + Secrets Manager read |
| `user` | string | Conditional | — | Snowflake user (with `private_key`; exactly one auth mode) |
| `private_key` | string | Conditional | — | RSA private key, PEM body without header/footer lines (sensitive). Prefer `secrets_manager`. |
| `key_passphrase` | string | No | — | Passphrase for an encrypted private key (sensitive; 7–255 chars) |
| `secrets_manager` | object | Conditional | — | Source credentials from Secrets Manager (secret shape `{"user": ..., "private_key": ..., "key_passphrase": ...}`) — the recommended production mode |
| `data_loading_option` | string | No | `"JSON_MAPPING"` | `"JSON_MAPPING"`, `"VARIANT_CONTENT_MAPPING"`, `"VARIANT_CONTENT_AND_METADATA_MAPPING"` |
| `content_column_name` | string | Conditional | — | VARIANT column for record content (required for VARIANT modes) |
| `metadata_column_name` | string | Conditional | — | VARIANT column for Firehose metadata (required for content+metadata mode) |
| `snowflake_role` | string | No | user's default | Snowflake role to assume (least-privilege ingestion role recommended) |
| `private_link_vpce_id` | string | No | — | PrivateLink VPCE ID for private connectivity |
| `buffering` | object | No | 0s / 1 MiB | Buffering hints — Snowpipe Streaming defaults to near-real-time |
| `retry_duration_in_seconds` | int32 | No | 60 | Retry duration: 0–7200s |
| `s3_backup_mode` | string | No | `"FailedDataOnly"` | `"FailedDataOnly"` or `"AllData"` |
| `s3_config` | object | **Yes** | — | S3 backup for failed/all records |
| `processing` | object | No | — | Record-transformation pipeline |
| `logging` | object | No | — | CloudWatch error logging |

#### Iceberg Destination (`iceberg`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `catalog_arn` | StringValueOrRef | **Yes** | — | Glue Data Catalog ARN (`arn:aws:glue:<region>:<account>:catalog`). **ForceNew**. |
| `role_arn` | StringValueOrRef | **Yes** | — | IAM role for Glue table read/update + warehouse S3 access |
| `destination_tables` | repeated object | No | — | Target Iceberg tables. Multiple tables require per-record routing metadata from a processor. **ForceNew**. |
| `destination_tables[].database_name` | string | **Yes** | — | Glue database containing the table |
| `destination_tables[].table_name` | string | **Yes** | — | Iceberg table name |
| `destination_tables[].s3_error_output_prefix` | string | No | — | Per-table error prefix |
| `destination_tables[].unique_keys` | repeated string | No | — | Row-identity columns enabling update/delete semantics |
| `append_only` | bool | No | false | Append-only mode (disables upserts). **ForceNew**. |
| `buffering` | object | No | 300s / 5 MiB | Buffering hints |
| `retry_duration_in_seconds` | int32 | No | 300 | Retry duration: 0–7200s |
| `s3_backup_mode` | string | No | `"FailedDataOnly"` | `"FailedDataOnly"` or `"AllData"` |
| `s3_config` | object | **Yes** | — | S3 backup for failed/all records |
| `processing` | object | No | — | Record-transformation pipeline |
| `logging` | object | No | — | CloudWatch error logging |

### Shared Sub-Messages

#### Buffering Hints (`buffering`)

| Field | Type | Range | Default | Description |
|-------|------|-------|---------|-------------|
| `interval_in_seconds` | int32 | 0–900 | destination-specific | Flush interval. Lower = less latency; higher = fewer objects. Splunk caps at 60. |
| `size_in_mbs` | int32 | 1–128 | destination-specific | Flush threshold in MiB. OpenSearch/HTTP cap at 100; Splunk at 5. |

Firehose delivers when **either** threshold is reached — whichever comes first.

#### Record-Transformation Pipeline (`processing`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | **Yes** | false | Enable the pipeline (requires at least one processor) |
| `processors` | repeated object | Conditional | — | Ordered processors; each entry sets exactly ONE typed arm |

Processor arms (exactly one per entry):

| Arm | Purpose | Fields |
|-----|---------|--------|
| `lambda` | Transform records with a Lambda function | `lambda_arn` (ref, required), `buffer_size_in_mbs` (0.2–3, fractional legal), `buffer_interval_in_seconds` (60–900), `number_of_retries` (0–300), `role_arn` (ref; a dedicated invocation role — set only when it differs from the delivery role, or the provider reports perpetual diffs) |
| `metadata_extraction` | Extract partition keys with a JQ expression (drives dynamic partitioning) | `query` (required), `json_parsing_engine` (`"JQ-1.6"`) |
| `decompression` | Decompress GZIP records (CloudWatch Logs subscriptions) | `compression_format` (`"GZIP"`) |
| `cloudwatch_log_processing` | Unwrap CloudWatch Logs subscription envelopes | `data_message_extraction` (bool) |
| `append_delimiter` | Append a delimiter per record (JSON lines) | `delimiter` (required, e.g. `"\\n"`). Extended S3 only. |
| `record_deaggregation` | Split aggregated payloads (KPL / delimited) | `sub_record_type` (`"JSON"`/`"DELIMITED"`), `delimiter` (base64, required for DELIMITED). Extended S3 only. |

AWS restricts `append_delimiter` and `record_deaggregation` to the extended_s3 destination (rejected at creation elsewhere) — enforced as a validation rule.

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

#### Secrets Manager Credentials (`secrets_manager`)

Available on the Redshift, Splunk, HTTP endpoint, and Snowflake destinations. Setting the block IS the enable switch (**ForceNew**); the destination's plaintext credential fields must then be empty.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `secret_arn` | StringValueOrRef | **Yes** | — | ARN of the secret holding the credential (shape depends on the destination) |
| `role_arn` | StringValueOrRef | No | delivery role | IAM role with `secretsmanager:GetSecretValue` |

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

Define the partition keys with a `metadata_extraction` processor and reference them in the S3 `prefix` as `!{partitionKeyFromQuery:<key>}`.

#### Data Format Conversion (`data_format_conversion`)

The input format is one deserializer arm (`open_x_json` or `hive_json`), the output format one serializer arm (`parquet` or `orc`) — exactly one of each is required when conversion is enabled, and an empty arm (`openXJson: {}`) accepts that format's defaults.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | **Yes** | false | Enable JSON → columnar conversion |
| `open_x_json` | object | Conditional | — | OpenX JSON deserializer arm (general-purpose JSON) |
| `hive_json` | object | Conditional | — | Apache Hive JSON deserializer arm (custom timestamp parsing) |
| `parquet` | object | Conditional | — | Apache Parquet serializer arm |
| `orc` | object | Conditional | — | Apache ORC serializer arm |
| `schema` | object | Conditional | — | Glue catalog schema (required when enabled) |

#### OpenX JSON Deserializer (`data_format_conversion.open_x_json`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `case_insensitive` | optional bool | No | true | Lowercase JSON keys before matching Glue columns |
| `column_to_json_key_mappings` | map | No | — | Column → JSON key overrides (e.g. `ts: timestamp` for reserved words) |
| `convert_dots_in_json_keys_to_underscores` | bool | No | false | Read `a.b` keys into `a_b` columns |

#### Hive JSON Deserializer (`data_format_conversion.hive_json`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `timestamp_formats` | repeated string | No | — | Joda-Time patterns for timestamp fields (`"millis"` for epoch ms) |

#### Parquet Serializer (`data_format_conversion.parquet`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `compression` | string | No | `"SNAPPY"` | `"SNAPPY"`, `"GZIP"`, `"UNCOMPRESSED"` |
| `block_size_bytes` | int64 | No | 268435456 | Row-group size; min 67108864 (64 MiB) |
| `page_size_bytes` | int64 | No | 1048576 | Page size; min 65536 (64 KiB) |
| `max_padding_bytes` | int64 | No | 0 | Max row-group padding |
| `enable_dictionary_compression` | bool | No | false | Dictionary-encode repeated values |
| `writer_version` | string | No | `"V1"` | `"V1"` or `"V2"` (confirm reader support for V2) |

#### ORC Serializer (`data_format_conversion.orc`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `compression` | string | No | `"SNAPPY"` | `"SNAPPY"`, `"ZLIB"`, `"NONE"` |
| `block_size_bytes` | int64 | No | 268435456 | Block size; min 67108864 (64 MiB) |
| `stripe_size_bytes` | int64 | No | 67108864 | Stripe size; min 8388608 (8 MiB) |
| `bloom_filter_columns` | repeated string | No | — | Columns to build bloom filters for (equality-filtered columns) |
| `bloom_filter_false_positive_probability` | optional double | No | 0.05 | 0–1; explicit 0 is legal and distinct from unset |
| `dictionary_key_threshold` | double | No | 0 | 0–1; distinct-key fraction above which dictionary encoding is abandoned |
| `enable_padding` | bool | No | false | Pad stripes to HDFS-style block boundaries |
| `padding_tolerance` | optional double | No | 0.05 | 0–1; max stripe fraction wasted as padding |
| `format_version` | string | No | `"V0_12"` | `"V0_11"` (legacy Hive 0.11) or `"V0_12"` |
| `row_index_stride` | int32 | No | 10000 | Rows between index entries; min 1000 |

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

## Outputs

| Output | Description |
|--------|-------------|
| `delivery_stream_arn` | ARN of the delivery stream (used by IAM policies, CloudWatch alarms, event sources) |
| `delivery_stream_name` | Name of the delivery stream (used for PutRecord/PutRecordBatch API calls) |
| `destination_id` | Identifier of the destination configuration (UpdateDestination coordinate) |
| `version_id` | Configuration version, incremented by AWS on every update |

## ForceNew Fields

The following fields require replacing the entire delivery stream if changed:

| Field | Reason |
|-------|--------|
| `metadata.name` | Delivery stream name is immutable |
| Destination type (which oneof arm is set) | Cannot switch destination types |
| `kinesis_stream_source` / `msk_source` (all fields) | Source type and configuration are immutable |
| `secrets_manager` presence (on any destination) | Enabling/disabling Secrets Manager auth replaces the stream |
| `dynamic_partitioning.enabled` | Cannot enable/disable after creation |
| OpenSearch-family `s3_backup_mode` | Backup mode is create-time on OpenSearch destinations |
| `vpc_config` (all fields) | VPC ENI configuration is immutable |
| Iceberg `catalog_arn`, `destination_tables`, `append_only` | Catalog binding and table routing are create-time |

## Scope

### Supported Destinations

1. **Extended S3** — Data lake storage with compression, record transformation, dynamic partitioning, Parquet/ORC format conversion
2. **OpenSearch** — Direct indexing with rotation, document-ID control, VPC delivery, S3 backup
3. **OpenSearch Serverless** — Indexing into serverless collections
4. **HTTP Endpoint** — Generic HTTPS delivery (Datadog, New Relic, Sumo Logic, custom APIs)
5. **Redshift** — Data warehouse loading via S3 staging + COPY command
6. **Splunk** — HEC delivery with indexer acknowledgment
7. **Snowflake** — Snowpipe Streaming inserts with key-pair auth and PrivateLink
8. **Iceberg** — Direct commits into Glue-cataloged Apache Iceberg tables with multi-table routing and upserts

### Supported Sources

1. **Direct PUT** (default) — Applications call PutRecord/PutRecordBatch APIs
2. **Kinesis Data Stream** — Firehose reads from an existing stream with automatic checkpointing
3. **Amazon MSK** — Firehose reads a Kafka topic (provisioned or serverless cluster; IAM auth)

### Deliberate Omissions

| Feature | Reason |
|---------|--------|
| Legacy Elasticsearch destination | A superseded API arm for the same domain fleet; `AwsOpenSearchDomain` (which also runs Elasticsearch engine versions) composes through the `opensearch` arm. |
| Custom prefix expressions via spec | Prefix expressions are set directly in the `prefix` string field. |

## Related Resources

- **AwsS3Bucket** — Destination bucket, backup bucket, or Redshift staging bucket
- **AwsKinesisStream** — Source stream when using Kinesis source mode
- **AwsMskCluster** — Source cluster when using MSK source mode
- **AwsIamRole** — Roles for Firehose to access destinations, S3, KMS, Lambda, Glue, Secrets Manager
- **AwsKmsKey** — Encryption key for SSE or S3 SSE-KMS
- **AwsLambda** — Transformation function for record processing
- **AwsOpenSearchDomain** — Target domain for the OpenSearch destination
- **AwsCloudwatchAlarm** — Monitor delivery metrics (DeliveryToS3.Success, IncomingBytes)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
