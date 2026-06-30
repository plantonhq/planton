---
title: "Kinesis Firehose"
description: "Kinesis Firehose deployment documentation"
icon: "package"
order: 100
componentName: "awskinesisfirehose"
---

# AWS Kinesis Firehose

Deploys an Amazon Kinesis Data Firehose delivery stream that captures, transforms, and delivers streaming data to S3, OpenSearch, HTTP endpoints, or Redshift. The component supports Direct PUT and Kinesis Data Stream sources, optional Lambda transformation, dynamic partitioning, and Parquet/ORC format conversion.

## What Gets Created

When you deploy an AwsKinesisFirehose resource, Planton provisions:

- **Kinesis Firehose Delivery Stream** — the core `aws_kinesis_firehose_delivery_stream` resource configured with the selected destination type
- **Kinesis source configuration** — created only when `kinesisStreamSource` is set, configures Firehose to consume from an existing Kinesis Data Stream with automatic checkpointing and retry
- **Server-side encryption** — created only when `sseEnabled` is `true`, encrypts data at rest in the delivery stream buffer using AWS-owned or customer-managed KMS keys (Direct PUT sources only)
- **Extended S3 destination** — primary S3 delivery with optional GZIP/Snappy compression, Lambda processing, dynamic partitioning, and Parquet/ORC format conversion via AWS Glue Data Catalog
- **OpenSearch destination** — direct indexing into an OpenSearch domain with configurable index rotation, VPC delivery, and S3 backup for failed documents
- **HTTP endpoint destination** — HTTPS delivery to any endpoint (Datadog, New Relic, Sumo Logic, custom APIs) with S3 backup for failed records
- **Redshift destination** — S3 staging followed by a Redshift COPY command for bulk data warehouse loading

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **A destination resource** — an S3 bucket, OpenSearch domain, HTTPS endpoint, or Redshift cluster depending on the chosen destination type
- **An IAM role** with permissions appropriate for the destination (S3 write, OpenSearch index, Redshift COPY, etc.)
- **A Kinesis Data Stream** if using Kinesis source mode instead of Direct PUT
- **An AWS Glue Data Catalog database and table** if enabling Parquet/ORC data format conversion
- **VPC subnets and security groups** if delivering to a VPC-deployed OpenSearch domain

## Quick Start

Create a file `firehose.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisFirehose
metadata:
  name: my-firehose
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsKinesisFirehose.my-firehose
spec:
  region: us-east-1
  extendedS3:
    bucketArn: arn:aws:s3:::my-data-bucket
    roleArn: arn:aws:iam::123456789012:role/firehose-s3-role
```

Deploy:

```shell
planton apply -f firehose.yaml
```

This creates a Direct PUT delivery stream that writes raw records to S3 with no compression or transformation.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the Firehose delivery stream will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |

Exactly one destination must be configured. The destination type is ForceNew — changing it requires replacing the delivery stream.

| Field | Type | Description |
|-------|------|-------------|
| `extendedS3` | `object` | Extended S3 destination for data lake storage with compression, partitioning, and format conversion |
| `opensearch` | `object` | OpenSearch destination for direct indexing with S3 backup |
| `httpEndpoint` | `object` | HTTP endpoint destination for HTTPS delivery with S3 backup |
| `redshift` | `object` | Redshift destination for S3 staging + COPY command |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `kinesisStreamSource` | `object` | — | Kinesis Data Stream source configuration. When absent, the delivery stream uses Direct PUT. ForceNew. |
| `kinesisStreamSource.streamArn` | `string` | — | ARN of the Kinesis Data Stream to consume from. Required when `kinesisStreamSource` is set. Can reference an AwsKinesisStream resource via `valueFrom`. |
| `kinesisStreamSource.roleArn` | `string` | — | IAM role ARN granting Firehose read access to the Kinesis stream. Required when `kinesisStreamSource` is set. Can reference an AwsIamRole resource via `valueFrom`. |
| `sseEnabled` | `bool` | `false` | Enables server-side encryption for data at rest in the delivery stream buffer. Only valid for Direct PUT sources. |
| `sseKmsKeyArn` | `string` | — | Customer-managed KMS key ARN for SSE. When absent, uses the AWS-owned CMK. Requires `sseEnabled` to be `true`. Can reference an AwsKmsKey resource via `valueFrom`. |

### Extended S3 Destination Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `extendedS3.bucketArn` | `string` | — | **(Required)** S3 bucket ARN for delivery. Can reference an AwsS3Bucket resource via `valueFrom`. |
| `extendedS3.roleArn` | `string` | — | **(Required)** IAM role ARN granting Firehose write access to S3, KMS, Lambda, and Glue as needed. Can reference an AwsIamRole resource via `valueFrom`. |
| `extendedS3.prefix` | `string` | — | S3 key prefix. Supports Firehose expression syntax (e.g., `year=!{timestamp:yyyy}/`). |
| `extendedS3.errorOutputPrefix` | `string` | — | S3 key prefix for records that fail transformation or delivery. |
| `extendedS3.compressionFormat` | `string` | `UNCOMPRESSED` | Compression applied before writing. Valid: `UNCOMPRESSED`, `GZIP`, `ZIP`, `Snappy`, `HADOOP_SNAPPY`. |
| `extendedS3.kmsKeyArn` | `string` | — | KMS key ARN for S3 server-side encryption (SSE-KMS). Can reference an AwsKmsKey resource via `valueFrom`. |
| `extendedS3.buffering` | `object` | `300s / 5 MiB` | Buffering hints: `intervalInSeconds` (0–900) and `sizeInMbs` (1–128). |
| `extendedS3.customTimeZone` | `string` | `UTC` | IANA time zone for S3 prefix timestamp expressions. |
| `extendedS3.fileExtension` | `string` | — | File extension appended to delivered objects (e.g., `.json`, `.parquet`). Must start with a period. |
| `extendedS3.s3BackupMode` | `string` | `Disabled` | When `Enabled`, a copy of pre-transformation records is written to `s3Backup`. |
| `extendedS3.s3Backup` | `object` | — | S3 configuration for source record backup. Required when `s3BackupMode` is `Enabled`. |
| `extendedS3.processing` | `object` | — | Lambda-based record transformation. Set `enabled`, `lambdaArn`, and optional buffer/retry settings. |
| `extendedS3.logging` | `object` | — | CloudWatch error logging. Set `enabled`, `logGroupName`, and `logStreamName`. |
| `extendedS3.dynamicPartitioning` | `object` | — | Dynamic partitioning by record fields for efficient querying with Athena/Spark. ForceNew. |
| `extendedS3.dataFormatConversion` | `object` | — | JSON-to-Parquet/ORC conversion via AWS Glue Data Catalog schema. |

### OpenSearch Destination Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `opensearch.domainArn` | `string` | — | ARN of the OpenSearch domain. Mutually exclusive with `clusterEndpoint`. Can reference an AwsOpenSearchDomain resource via `valueFrom`. |
| `opensearch.clusterEndpoint` | `string` | — | OpenSearch cluster endpoint URL. Mutually exclusive with `domainArn`. |
| `opensearch.indexName` | `string` | — | **(Required)** Index name (or prefix when rotation is enabled). |
| `opensearch.roleArn` | `string` | — | **(Required)** IAM role ARN with `es:ESHttpPut` and `es:ESHttpGet` permissions. Can reference an AwsIamRole resource via `valueFrom`. |
| `opensearch.s3Config` | `object` | — | **(Required)** S3 configuration for backing up failed or all documents. |
| `opensearch.indexRotationPeriod` | `string` | `OneDay` | Index rotation period. Valid: `NoRotation`, `OneHour`, `OneDay`, `OneWeek`, `OneMonth`. |
| `opensearch.typeName` | `string` | — | Document type name. Only relevant for Elasticsearch 6.x and earlier. |
| `opensearch.buffering` | `object` | `300s / 5 MiB` | Buffering hints. Max size: 100 MiB for OpenSearch destinations. |
| `opensearch.retryDurationInSeconds` | `int` | `300` | Retry duration for failed index requests. Range: 0–7200. |
| `opensearch.s3BackupMode` | `string` | `FailedDocumentsOnly` | Valid: `FailedDocumentsOnly`, `AllDocuments`. |
| `opensearch.processing` | `object` | — | Lambda-based record transformation before indexing. |
| `opensearch.logging` | `object` | — | CloudWatch error logging for delivery failures. |
| `opensearch.vpcConfig` | `object` | — | VPC configuration for VPC-deployed OpenSearch domains. ForceNew. |

### HTTP Endpoint Destination Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `httpEndpoint.url` | `string` | — | **(Required)** HTTPS URL of the destination endpoint. Must start with `https://`. |
| `httpEndpoint.s3Config` | `object` | — | **(Required)** S3 configuration for backing up failed or all records. |
| `httpEndpoint.name` | `string` | — | Human-readable endpoint name for the AWS Console and CloudWatch metrics. |
| `httpEndpoint.accessKey` | `string` | — | Access key sent in the `X-Amz-Firehose-Access-Key` header. Sensitive. |
| `httpEndpoint.roleArn` | `string` | — | IAM role ARN for delivery and S3 backup. Can reference an AwsIamRole resource via `valueFrom`. |
| `httpEndpoint.buffering` | `object` | `300s / 5 MiB` | Buffering hints for HTTP delivery. |
| `httpEndpoint.retryDurationInSeconds` | `int` | `300` | Retry duration for non-2xx responses. Range: 0–7200. |
| `httpEndpoint.s3BackupMode` | `string` | `FailedDataOnly` | Valid: `FailedDataOnly`, `AllData`. |
| `httpEndpoint.processing` | `object` | — | Lambda-based record transformation before HTTP delivery. |
| `httpEndpoint.logging` | `object` | — | CloudWatch error logging for delivery failures. |
| `httpEndpoint.requestConfig` | `object` | — | Request customization: `contentEncoding` (`NONE`, `GZIP`) and `commonAttributes` (key-value headers). |

### Redshift Destination Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `redshift.clusterJdbcurl` | `string` | — | **(Required)** JDBC URL of the Redshift cluster (e.g., `jdbc:redshift://host:5439/db`). |
| `redshift.roleArn` | `string` | — | **(Required)** IAM role ARN for S3 read and Redshift COPY. Can reference an AwsIamRole resource via `valueFrom`. |
| `redshift.dataTableName` | `string` | — | **(Required)** Target Redshift table for the COPY command. |
| `redshift.s3Config` | `object` | — | **(Required)** S3 configuration for intermediate staging. Firehose writes here before issuing COPY. |
| `redshift.dataTableColumns` | `string` | — | Comma-separated column names for the COPY command. When absent, COPY loads all columns in table order. |
| `redshift.copyOptions` | `string` | — | Additional COPY command options (e.g., `JSON 'auto'`, `GZIP`, `DELIMITER ','`). |
| `redshift.username` | `string` | — | Redshift database username. |
| `redshift.password` | `string` | — | Redshift database password. Sensitive. |
| `redshift.retryDurationInSeconds` | `int` | `3600` | Retry duration for failed COPY commands. Range: 0–7200. |
| `redshift.s3BackupMode` | `string` | `Disabled` | When `Enabled`, original records are backed up to `s3Backup`. |
| `redshift.s3Backup` | `object` | — | S3 configuration for source record backup. Required when `s3BackupMode` is `Enabled`. |
| `redshift.processing` | `object` | — | Lambda-based record transformation before staging. |
| `redshift.logging` | `object` | — | CloudWatch error logging for COPY failures. |

## Examples

### Extended S3 Data Lake

GZIP-compressed delivery to S3 with timestamp-based prefixes and buffering tuned for throughput:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisFirehose
metadata:
  name: data-lake-firehose
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsKinesisFirehose.data-lake-firehose
spec:
  region: us-east-1
  extendedS3:
    bucketArn: arn:aws:s3:::my-data-lake-bucket
    roleArn: arn:aws:iam::123456789012:role/firehose-s3-delivery-role
    prefix: events/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/
    compressionFormat: GZIP
    fileExtension: .json.gz
    buffering:
      intervalInSeconds: 120
      sizeInMbs: 64
```

### OpenSearch Log Analytics

Indexes application logs into an OpenSearch domain with daily index rotation and S3 backup for failed documents. References the OpenSearch domain via `valueFrom`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisFirehose
metadata:
  name: log-analytics-firehose
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsKinesisFirehose.log-analytics-firehose
spec:
  region: us-east-1
  opensearch:
    domainArn:
      valueFrom:
        kind: AwsOpenSearchDomain
        name: my-log-domain
        fieldPath: status.outputs.domain_arn
    indexName: application-logs
    roleArn: arn:aws:iam::123456789012:role/firehose-opensearch-role
    indexRotationPeriod: OneDay
    s3BackupMode: FailedDocumentsOnly
    buffering:
      intervalInSeconds: 60
      sizeInMbs: 5
    s3Config:
      bucketArn: arn:aws:s3:::my-firehose-backup-bucket
      roleArn: arn:aws:iam::123456789012:role/firehose-s3-backup-role
      prefix: opensearch-backup/failed/
      compressionFormat: GZIP
```

### Production S3 with Kinesis Source and Parquet Conversion

Consumes from an existing Kinesis Data Stream, converts JSON to Parquet via AWS Glue Data Catalog, and writes columnar files to a partitioned S3 data lake:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisFirehose
metadata:
  name: analytics-parquet-firehose
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsKinesisFirehose.analytics-parquet-firehose
spec:
  region: us-east-1
  kinesisStreamSource:
    streamArn:
      valueFrom:
        kind: AwsKinesisStream
        name: my-events-stream
        fieldPath: status.outputs.stream_arn
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: firehose-kinesis-consumer-role
        fieldPath: status.outputs.role_arn
  extendedS3:
    bucketArn: arn:aws:s3:::my-analytics-data-lake
    roleArn: arn:aws:iam::123456789012:role/firehose-glue-s3-role
    prefix: analytics/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/
    compressionFormat: UNCOMPRESSED
    fileExtension: .parquet
    buffering:
      intervalInSeconds: 60
      sizeInMbs: 64
    dynamicPartitioning:
      enabled: true
      retryDurationInSeconds: 300
    dataFormatConversion:
      enabled: true
      inputFormat: OPENX_JSON
      outputFormat: PARQUET
      parquetCompression: SNAPPY
      schema:
        databaseName: analytics_db
        tableName: events
        roleArn: arn:aws:iam::123456789012:role/firehose-glue-access-role
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `delivery_stream_arn` | `string` | ARN of the Kinesis Data Firehose delivery stream |
| `delivery_stream_name` | `string` | Name of the delivery stream, unique within the AWS account and region |

## Related Components

- [AwsKinesisStream](/docs/catalog/aws/kinesis-data-stream) — provides a Kinesis Data Stream as the source for the delivery stream
- [AwsS3Bucket](/docs/catalog/aws/s3-bucket) — serves as the delivery destination, backup target, or Redshift staging area
- [AwsOpenSearchDomain](/docs/catalog/aws/opensearch-domain) — serves as the indexing destination for log and search workloads
- [AwsIamRole](/docs/catalog/aws/iam-role) — provides the permissions Firehose needs for source, destination, and transformation access
- [AwsLambda](/docs/catalog/aws/lambda) — provides the Lambda function for record transformation before delivery
