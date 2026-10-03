# AWS Kinesis Firehose

Deploys a Kinesis Data Firehose delivery stream that captures, optionally transforms, and delivers streaming data to exactly one of eight destinations: Extended S3 (with Parquet/ORC conversion and dynamic partitioning), OpenSearch, OpenSearch Serverless, HTTP endpoint, Redshift, Splunk, Snowflake, or Apache Iceberg tables in the AWS Glue Data Catalog. Data enters from Direct PUT API calls, an existing Kinesis Data Stream, or an Amazon MSK topic. Supports server-side encryption (Direct PUT only), an ordered record-transformation pipeline (Lambda, JQ metadata extraction, decompression, CloudWatch Logs unwrapping, delimiting, de-aggregation), S3 backup for failed deliveries on every non-S3 destination, and AWS Secrets Manager credential delivery for the credentialed destinations. Every source, destination, role, bucket, key, and network reference accepts ValueFromRef wiring, so the whole delivery pipeline — S3 buckets, KMS keys, IAM roles, Kinesis streams, MSK clusters, Lambda transformers, subnets, security groups, and OpenSearch domains — composes in one InfraChart. The destination type and the source decision are both immutable after creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Firehose Delivery Stream** -- a managed delivery pipeline named from your manifest's `metadata.name`, configured with exactly one destination (Extended S3, OpenSearch, OpenSearch Serverless, HTTP endpoint, Redshift, Splunk, Snowflake, or Iceberg)
- **Source Configuration** -- Direct PUT by default; optionally reads from a Kinesis Data Stream (`kinesisStreamSource`) or an Amazon MSK topic (`mskSource`)
- **Server-Side Encryption** -- configured only when `sseEnabled` is `true` (Direct PUT sources only); uses the AWS-owned CMK by default, or a customer-managed KMS key when `sseKmsKeyArn` is provided
- **Destination-Specific Resources** -- buffering configuration, the required S3 path for failed (or all) records on non-S3 destinations, the optional transformation pipeline, CloudWatch error logging, and destination-specific settings (dynamic partitioning, Glue Data Catalog schema conversion, VPC delivery, Redshift COPY staging, Splunk HEC acknowledgment, Snowpipe Streaming, Iceberg table routing with upsert keys)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An S3 bucket** -- required for every destination. Extended S3 uses it as the primary target; every other destination requires it for failed (or all) record backup, and Redshift additionally stages COPY data through it. Provide the bucket ARN directly or reference an AwsS3Bucket Infra Component via ValueFromRef.
- **An IAM role** -- required per destination. The role must grant Firehose write access to the target service and, when used, to Lambda (transformation), Glue (format conversion or Iceberg), and Secrets Manager (credential resolution). Provide the ARN directly or reference an AwsIamRole Infra Component.
- **A source stream or cluster** (optional) -- an existing Kinesis Data Stream or an MSK cluster with IAM access control enabled, when not using Direct PUT. Both must exist before the delivery stream is created.
- **Destination infrastructure** -- the OpenSearch domain/collection, Redshift table, Splunk HEC input, Snowflake table, or Glue-cataloged Iceberg tables must exist first; Firehose delivers into existing infrastructure.
- **A KMS key** (optional) -- for server-side encryption of the delivery stream buffer (Direct PUT only) or S3 object encryption.

## Deploy

### Console

Open the deployment store, find **AWS Kinesis Firehose**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the source decision, the eight-way destination choice, and the destination's own target, delivery, transformation, backup, and monitoring steps. Six presets cover the common shapes -- start from one and adjust.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsKinesisFirehose
metadata:
  name: logs-to-s3
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  extendedS3:
    bucketArn:
      value: "arn:aws:s3:::my-data-lake"
    roleArn:
      value: "arn:aws:iam::123456789012:role/FirehoseDeliveryRole"
    compressionFormat: GZIP
    prefix: "logs/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/"
```

```shell
planton apply -f kinesis-firehose.yaml
```

This creates a Direct PUT delivery stream that compresses records with GZIP and delivers them to S3 with date-partitioned prefixes. No encryption, transformation, or format conversion is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the delivery stream to an S3 bucket and IAM role deployed in the same InfraPipeline:

```yaml
spec:
  extendedS3:
    bucketArn:
      valueFrom:
        kind: AwsS3Bucket
        name: data-lake
        fieldPath: status.outputs.bucket_arn
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: firehose-delivery-role
        fieldPath: status.outputs.role_arn
  sseKmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: streaming-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the S3 bucket, IAM role, and KMS key first, then provisions the delivery stream with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Firehose delivery stream. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Destination** -- Exactly one of the eight destination blocks: `extendedS3` for data lakes and log archives (the only arm with dynamic partitioning, Parquet/ORC conversion, and the delimiter/de-aggregation processors), `opensearch` / `opensearchServerless` for search indexing, `httpEndpoint` for third-party integrations (Datadog, New Relic, Sumo Logic), `redshift` for warehouse loading via S3 staging + COPY, `splunk` for HEC delivery with indexer acknowledgment, `snowflake` for near-real-time Snowpipe Streaming inserts, or `iceberg` for Glue-cataloged lakehouse tables with unique-key upserts. The destination type is immutable after creation.

**Source** -- Direct PUT (default) accepts PutRecord/PutRecordBatch API calls; `kinesisStreamSource` consumes an existing stream; `mskSource` reads a Kafka topic on an MSK cluster. The two source blocks are mutually exclusive and the whole source decision is immutable after creation. With a stream or MSK source, server-side encryption must stay disabled -- the source handles encryption.

**Credentials** -- Redshift, Splunk, HTTP endpoints, and Snowflake authenticate with exactly one of: the embedded credential fields (marked sensitive -- org-secret references only) or an AWS Secrets Manager secret (`secretsManager` -- the production posture: rotations apply without touching the stream, and enabling/disabling it after creation replaces the stream).

**Buffering** -- Each destination's `buffering` carries `intervalInSeconds` and `sizeInMbs`; Firehose flushes when either threshold is met, and 0/absent means the AWS default. Splunk caps at 60 s / 5 MiB; the OpenSearch pair and HTTP endpoints cap the size at 100 MiB; Redshift buffers through its S3 staging configuration.

**Data format conversion** -- For Extended S3, enable `dataFormatConversion` to convert JSON to Apache Parquet or ORC through an AWS Glue Data Catalog schema: pick one deserializer arm (`openXJson` or `hiveJson`) and one serializer arm (`parquet` or `orc`), each carrying its format's full tuning surface (compression codec, block/page/stripe sizes, bloom filters, dictionary encoding, writer/format versions). Pair it with `dynamicPartitioning` (immutable after creation) to lay objects out by record fields for query pruning.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `sseKmsKeyArn`, S3 `kmsKeyArn` | `status.outputs.key_arn` |
| **AwsKinesisStream** (optional) | `kinesisStreamSource.streamArn` | `status.outputs.stream_arn` |
| **AwsMskCluster** (optional) | `mskSource.mskClusterArn` | `status.outputs.cluster_arn` |
| **AwsIamRole** | source/destination/S3/VPC/Glue/Secrets `roleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** | destination and backup `bucketArn` | `status.outputs.bucket_arn` |
| **AwsLambda** (optional) | `processing.processors[].lambda.lambdaArn` | `status.outputs.function_arn` |
| **AwsSubnet** (optional) | OpenSearch `vpcConfig.subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | OpenSearch `vpcConfig.securityGroupIds` | `status.outputs.security_group_id` |
| **AwsOpenSearchDomain** (optional) | `opensearch.domainArn` | `status.outputs.domain_arn` |

The Secrets Manager `secretArn` and the Iceberg `catalogArn` are value-or-reference fields with no default catalog kind -- provide those ARNs directly.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `delivery_stream_arn` | Amazon Resource Name of the delivery stream | IAM policies, CloudWatch alarm dimensions, MSK/WAF/Cognito log delivery targets |
| `delivery_stream_name` | Delivery stream name | Firehose API calls (PutRecord, PutRecordBatch), application configuration |
| `destination_id` | Identifier of the configured destination | Destination update tracking |
| `version_id` | Configuration version of the delivery stream | Change auditing |

## Common Patterns

Six presets cover the common delivery shapes:

- **S3 Data Lake Preset** -- minimal Extended S3 with GZIP compression and time-partitioned prefixes
- **OpenSearch Log Analytics Preset** -- domain indexing with daily rotation and failed-document backup
- **HTTP Endpoint Webhook Preset** -- third-party intake (Datadog) with GZIP request encoding and custom headers
- **S3 Parquet Analytics Preset** -- a Kinesis-stream source feeding dynamic partitioning and Parquet conversion through a Glue schema
- **Snowflake Streaming Preset** -- Snowpipe Streaming with Secrets Manager key-pair credentials and a least-privilege Snowflake role
- **Iceberg Lakehouse Preset** -- Glue-cataloged Iceberg tables with unique-key upserts

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for delivery stream buffer encryption and S3 SSE-KMS
- [**AWS Kinesis Data Stream**](/infra-catalog/aws-kinesis-stream) -- provides a streaming source for the delivery stream
- [**AWS MSK Cluster**](/infra-catalog/aws-msk-cluster) -- provides a Kafka topic source for the delivery stream
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides service roles for Firehose to access S3, OpenSearch, Redshift, Lambda, Glue, and Secrets Manager
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- provides the destination bucket for Extended S3, the staging bucket for Redshift, and backup buckets for every destination
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- provides a transformation function for record processing before delivery
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides subnets for VPC delivery to OpenSearch domains and collections
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides network access control for VPC delivery ENIs
- [**AWS OpenSearch Domain**](/infra-catalog/aws-open-search-domain) -- provides the OpenSearch domain for direct indexing
