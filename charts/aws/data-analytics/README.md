# AWS Data Analytics

Provisions a data lake analytics pipeline using Kinesis for real-time ingestion, Firehose for automatic S3 delivery with compression and time-based partitioning, Glue Data Catalog for schema management, and Athena for serverless SQL queries against the data lake.

This is the canonical AWS streaming analytics pattern. Data flows from producers through Kinesis, is delivered to S3 in compressed partitioned files by Firehose, cataloged in Glue for schema-on-read, and queried interactively with Athena SQL.

## Architecture

```
  Data Producers
       │
       ▼
┌──────────────────┐
│ AwsKinesisStream │
│ (real-time ingest)│
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│AwsKinesisFirehose│
│ (delivery stream)│
│  GZIP + partition│
└────────┬─────────┘
         │
         ▼
┌──────────────────┐      ┌──────────────────────┐
│   AwsS3Bucket    │◀────▶│AwsGlueCatalogDatabase│
│  (data lake)     │      │  (table schemas)     │
│  /raw/year=.../  │      └──────────────────────┘
│  /athena-results/│               │
└──────────────────┘               ▼
                          ┌──────────────────────┐
                          │ AwsAthenaWorkgroup   │
                          │ (SQL query engine)   │
                          └──────────────────────┘

┌──────────────────────┐
│  AwsIamRole          │
│  (Firehose delivery) │
└──────────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  AwsS3Bucket, AwsKinesisStream, AwsGlueCatalogDatabase
Layer 1 (dep S3):    AwsIamRole, AwsAthenaWorkgroup
Layer 2 (dep all):   AwsKinesisFirehose
```

## Included Cloud Resources

| Resource | Kind | Group | Condition | Purpose |
|----------|------|-------|-----------|---------|
| S3 Bucket | `AwsS3Bucket` | storage | Always | Data lake storage + Athena results |
| Kinesis Stream | `AwsKinesisStream` | messaging | Always | Real-time data ingestion |
| IAM Role | `AwsIamRole` | identity | Always | Firehose Kinesis read + S3 write |
| Kinesis Firehose | `AwsKinesisFirehose` | messaging | Always | Kinesis-to-S3 delivery with compression |
| Glue Catalog DB | `AwsGlueCatalogDatabase` | analytics | Always | Table schema metadata |
| Athena Workgroup | `AwsAthenaWorkgroup` | analytics | Always | Serverless SQL query engine |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `aws_region` | AWS region | `us-east-1` | Yes |
| `s3_bucket_name` | Data lake S3 bucket (globally unique) | `my-data-lake` | Yes |
| **Kinesis** | | | |
| `kinesis_stream_name` | Stream name | `data-ingest` | Yes |
| `kinesis_stream_mode` | `ON_DEMAND` or `PROVISIONED` | `ON_DEMAND` | Yes |
| `kinesis_shard_count` | Shard count (PROVISIONED only) | `0` | No |
| **Firehose** | | | |
| `firehose_name` | Delivery stream name | `data-delivery` | Yes |
| `firehose_s3_prefix` | S3 key prefix with partitioning | `raw/year=.../` | Yes |
| `firehose_buffer_interval` | Flush interval (60-900s) | `300` | No |
| `firehose_buffer_size` | Flush size (1-128 MiB) | `5` | No |
| `firehose_compression` | Compression format | `GZIP` | No |
| **Glue** | | | |
| `glue_database_name` | Catalog database name | `analytics_db` | Yes |
| `glue_database_description` | Database description | `Data lake analytics tables` | No |
| **Athena** | | | |
| `athena_workgroup_name` | Workgroup name | `analytics-workgroup` | Yes |
| `athena_bytes_scanned_cutoff` | Max bytes per query (cost control) | `10737418240` (10 GB) | No |

## Data Flow

1. **Ingest**: Applications put records into the Kinesis stream using PutRecord/PutRecordBatch APIs
2. **Deliver**: Firehose reads from Kinesis and buffers records, flushing to S3 every 5 minutes (or 5 MiB) in GZIP-compressed files with time-based partitioning
3. **Catalog**: Create Glue tables (via Glue crawlers or DDL) pointing to the S3 data location
4. **Query**: Run SQL queries in Athena against the Glue catalog tables

## Post-Deployment Steps

1. **Create Glue tables** for your data schema. Use a Glue Crawler or Athena DDL:
   ```sql
   CREATE EXTERNAL TABLE analytics_db.events (
     event_id STRING,
     event_type STRING,
     payload STRING,
     timestamp TIMESTAMP
   )
   PARTITIONED BY (year STRING, month STRING, day STRING)
   ROW FORMAT SERDE 'org.openx.data.jsonserde.JsonSerDe'
   LOCATION 's3://my-data-lake/raw/'
   TBLPROPERTIES ('has_encrypted_data'='false');
   ```

2. **Load partitions** after data arrives:
   ```sql
   MSCK REPAIR TABLE analytics_db.events;
   ```

3. **Query the data**:
   ```sql
   SELECT event_type, COUNT(*) as cnt
   FROM analytics_db.events
   WHERE year='2026' AND month='02'
   GROUP BY event_type
   ORDER BY cnt DESC;
   ```

## Important Notes

- The S3 bucket includes a **lifecycle rule** that transitions raw data to Standard-IA after 90 days. Modify or remove this rule based on your access patterns.
- Kinesis uses **ON_DEMAND** mode by default (no capacity planning needed). Switch to PROVISIONED with explicit shard count for cost optimization on steady workloads.
- Firehose delivers data in **GZIP** format by default. Athena supports querying GZIP files natively. For Parquet/ORC conversion, configure `dataFormatConversion` on the Firehose resource after deployment.
- Athena query cost control is set to **10 GB** per query. Adjust `athena_bytes_scanned_cutoff` based on your dataset sizes. Partitioned tables significantly reduce scan volume.
- The Glue database is created with a default `locationUri` pointing to the `raw/` prefix. Individual tables can override this with their own S3 locations.
