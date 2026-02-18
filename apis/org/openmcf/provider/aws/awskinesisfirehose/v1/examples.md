# AwsKinesisFirehose Examples

## 1. Minimal Extended S3 (Direct PUT Data Lake)

The simplest delivery stream. Direct PUT source, GZIP compression, delivers to an S3 bucket for data lake storage. No transformation, no format conversion.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisFirehose
metadata:
  name: events-to-s3
  org: acme
  env: dev
  id: events-to-s3-dev
spec:
  region: us-east-1
  extendedS3:
    bucketArn:
      value: arn:aws:s3:::acme-data-lake-dev
    roleArn:
      value: arn:aws:iam::123456789012:role/firehose-s3-role
    prefix: "raw/events/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/"
    errorOutputPrefix: "errors/events/"
    compressionFormat: GZIP
    fileExtension: ".json.gz"
    buffering:
      intervalInSeconds: 300
      sizeInMbs: 5
```

## 2. Extended S3 with Parquet Conversion (Glue Catalog)

Converts incoming JSON records to Apache Parquet using an AWS Glue Data Catalog schema. SNAPPY compression for optimal analytics query performance with Athena, Spark, or Presto.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisFirehose
metadata:
  name: analytics-parquet
  org: acme
  env: production
  id: analytics-parquet-prod
spec:
  region: us-east-1
  extendedS3:
    bucketArn:
      valueFrom:
        kind: AwsS3Bucket
        name: analytics-data-lake
        fieldPath: status.outputs.bucket_arn
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: firehose-analytics-role
        fieldPath: status.outputs.role_arn
    prefix: "parquet/events/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/"
    errorOutputPrefix: "errors/parquet-conversion/"
    fileExtension: ".parquet"
    buffering:
      intervalInSeconds: 900
      sizeInMbs: 128
    dataFormatConversion:
      enabled: true
      inputFormat: OPENX_JSON
      outputFormat: PARQUET
      parquetCompression: SNAPPY
      schema:
        databaseName: analytics_db
        tableName: clickstream_events
        roleArn:
          valueFrom:
            kind: AwsIamRole
            name: firehose-glue-role
            fieldPath: status.outputs.role_arn
```

## 3. Extended S3 with Kinesis Source, Lambda Processing, and Dynamic Partitioning

Reads from a Kinesis Data Stream, transforms records with Lambda, and dynamically partitions into S3 by a field extracted from the record (e.g., `customer_id`). Ideal for multi-tenant data lake pipelines.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisFirehose
metadata:
  name: multi-tenant-pipeline
  org: acme
  env: production
  id: multi-tenant-pipeline-prod
spec:
  region: us-east-1
  kinesisStreamSource:
    streamArn:
      valueFrom:
        kind: AwsKinesisStream
        name: order-events
        fieldPath: status.outputs.stream_arn
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: firehose-kinesis-reader-role
        fieldPath: status.outputs.role_arn
  extendedS3:
    bucketArn:
      valueFrom:
        kind: AwsS3Bucket
        name: multi-tenant-lake
        fieldPath: status.outputs.bucket_arn
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: firehose-s3-writer-role
        fieldPath: status.outputs.role_arn
    prefix: "data/customer=!{partitionKeyFromLambda:customer_id}/year=!{timestamp:yyyy}/month=!{timestamp:MM}/"
    errorOutputPrefix: "errors/multi-tenant/"
    compressionFormat: GZIP
    fileExtension: ".json.gz"
    buffering:
      intervalInSeconds: 60
      sizeInMbs: 64
    processing:
      enabled: true
      lambdaArn:
        valueFrom:
          kind: AwsLambda
          name: firehose-enrichment
          fieldPath: status.outputs.function_arn
      bufferSizeInMbs: 3
      bufferIntervalInSeconds: 60
      numberOfRetries: 3
    dynamicPartitioning:
      enabled: true
      retryDurationInSeconds: 300
    logging:
      enabled: true
      logGroupName: /aws/kinesisfirehose/multi-tenant-pipeline
      logStreamName: S3Delivery
```

## 4. OpenSearch Log Analytics

Indexes log records into an Amazon OpenSearch domain with daily index rotation. VPC delivery for a private OpenSearch cluster. Failed documents are backed up to S3.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisFirehose
metadata:
  name: app-logs-to-opensearch
  org: acme
  env: production
  id: app-logs-to-opensearch-prod
spec:
  region: us-east-1
  sseEnabled: true
  opensearch:
    domainArn:
      valueFrom:
        kind: AwsOpenSearchDomain
        name: log-analytics
        fieldPath: status.outputs.domain_arn
    indexName: app-logs
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: firehose-opensearch-role
        fieldPath: status.outputs.role_arn
    indexRotationPeriod: OneDay
    buffering:
      intervalInSeconds: 60
      sizeInMbs: 5
    retryDurationInSeconds: 300
    s3BackupMode: FailedDocumentsOnly
    s3Config:
      bucketArn:
        valueFrom:
          kind: AwsS3Bucket
          name: opensearch-backup
          fieldPath: status.outputs.bucket_arn
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: firehose-s3-backup-role
          fieldPath: status.outputs.role_arn
      prefix: "opensearch-failures/"
      compressionFormat: GZIP
    vpcConfig:
      subnetIds:
        - value: subnet-0a1b2c3d4e5f60001
        - value: subnet-0a1b2c3d4e5f60002
      securityGroupIds:
        - value: sg-0a1b2c3d4e5f60001
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: firehose-vpc-role
          fieldPath: status.outputs.role_arn
    logging:
      enabled: true
      logGroupName: /aws/kinesisfirehose/app-logs-to-opensearch
      logStreamName: OpenSearchDelivery
```

## 5. HTTP Endpoint to Datadog

Delivers records to Datadog's HTTP intake API with GZIP content encoding, an access key for authentication, and custom attributes for tagging. Failed deliveries are backed up to S3.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisFirehose
metadata:
  name: metrics-to-datadog
  org: acme
  env: production
  id: metrics-to-datadog-prod
spec:
  region: us-east-1
  sseEnabled: true
  httpEndpoint:
    url: "https://aws-kinesis-http-intake.logs.datadoghq.com/v1/input"
    name: Datadog
    accessKey: "dd-api-key-abc123def456"
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: firehose-datadog-role
        fieldPath: status.outputs.role_arn
    buffering:
      intervalInSeconds: 60
      sizeInMbs: 4
    retryDurationInSeconds: 300
    s3BackupMode: FailedDataOnly
    s3Config:
      bucketArn:
        valueFrom:
          kind: AwsS3Bucket
          name: datadog-backup
          fieldPath: status.outputs.bucket_arn
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: firehose-s3-backup-role
          fieldPath: status.outputs.role_arn
      prefix: "datadog-failures/"
      compressionFormat: GZIP
    requestConfig:
      contentEncoding: GZIP
      commonAttributes:
        - name: env
          value: production
        - name: service
          value: platform-metrics
        - name: source
          value: kinesis-firehose
    logging:
      enabled: true
      logGroupName: /aws/kinesisfirehose/metrics-to-datadog
      logStreamName: HttpEndpointDelivery
```

## 6. Redshift Warehouse Loading

Stages records in S3 and issues a COPY command to load data into a Redshift table. Uses JSON auto-detection for schema mapping.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisFirehose
metadata:
  name: orders-to-redshift
  org: acme
  env: production
  id: orders-to-redshift-prod
spec:
  region: us-east-1
  sseEnabled: true
  sseKmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: firehose-encryption-key
      fieldPath: status.outputs.key_arn
  redshift:
    clusterJdbcurl: "jdbc:redshift://acme-warehouse.abcdef123456.us-east-1.redshift.amazonaws.com:5439/analytics"
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: firehose-redshift-role
        fieldPath: status.outputs.role_arn
    dataTableName: public.order_events
    copyOptions: "JSON 'auto' GZIP TIMEFORMAT 'auto' TRUNCATECOLUMNS"
    username: firehose_loader
    password:
      value: "change-me-use-secrets-manager-in-production"
    s3Config:
      bucketArn:
        valueFrom:
          kind: AwsS3Bucket
          name: redshift-staging
          fieldPath: status.outputs.bucket_arn
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: firehose-s3-staging-role
          fieldPath: status.outputs.role_arn
      prefix: "redshift-staging/orders/"
      compressionFormat: GZIP
      buffering:
        intervalInSeconds: 300
        sizeInMbs: 10
    retryDurationInSeconds: 3600
    s3BackupMode: Enabled
    s3Backup:
      bucketArn:
        valueFrom:
          kind: AwsS3Bucket
          name: redshift-source-backup
          fieldPath: status.outputs.bucket_arn
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: firehose-s3-backup-role
          fieldPath: status.outputs.role_arn
      prefix: "source-backup/orders/"
      compressionFormat: GZIP
    logging:
      enabled: true
      logGroupName: /aws/kinesisfirehose/orders-to-redshift
      logStreamName: RedshiftDelivery
```

## 7. Production Extended S3 with All Features

Full-featured production delivery stream combining: Kinesis source, SSE-KMS encryption on the source stream, Lambda transformation, dynamic partitioning, Parquet format conversion via Glue, source record backup, and CloudWatch error logging.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisFirehose
metadata:
  name: clickstream-pipeline
  org: acme
  env: production
  id: clickstream-pipeline-prod
  labels:
    team: data-engineering
    cost-center: analytics
    compliance: gdpr
spec:
  region: us-east-1
  # Kinesis source — reads from the clickstream events stream.
  # SSE is NOT set here because the source stream handles encryption.
  kinesisStreamSource:
    streamArn:
      valueFrom:
        kind: AwsKinesisStream
        name: clickstream-events
        fieldPath: status.outputs.stream_arn
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: firehose-kinesis-reader
        fieldPath: status.outputs.role_arn

  # Extended S3 destination with every feature enabled
  extendedS3:
    bucketArn:
      valueFrom:
        kind: AwsS3Bucket
        name: clickstream-data-lake
        fieldPath: status.outputs.bucket_arn
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: firehose-pipeline-role
        fieldPath: status.outputs.role_arn

    # Dynamic prefix with partition keys extracted by Lambda
    prefix: "data/region=!{partitionKeyFromLambda:region}/customer=!{partitionKeyFromLambda:customer_id}/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/hour=!{timestamp:HH}/"
    errorOutputPrefix: "errors/clickstream/!{firehose:error-output-type}/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/"
    fileExtension: ".parquet"
    customTimeZone: "UTC"

    # S3 SSE-KMS encryption
    kmsKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: data-lake-encryption-key
        fieldPath: status.outputs.key_arn

    # Large buffers for optimal Parquet file sizes
    buffering:
      intervalInSeconds: 900
      sizeInMbs: 128

    # Lambda transformation — enriches records and extracts partition keys
    processing:
      enabled: true
      lambdaArn:
        valueFrom:
          kind: AwsLambda
          name: clickstream-enrichment
          fieldPath: status.outputs.function_arn
      bufferSizeInMbs: 3
      bufferIntervalInSeconds: 60
      numberOfRetries: 3

    # Dynamic partitioning by region and customer_id
    dynamicPartitioning:
      enabled: true
      retryDurationInSeconds: 300

    # JSON → Parquet via Glue catalog
    dataFormatConversion:
      enabled: true
      inputFormat: OPENX_JSON
      outputFormat: PARQUET
      parquetCompression: SNAPPY
      schema:
        databaseName: clickstream_db
        tableName: events_schema
        roleArn:
          valueFrom:
            kind: AwsIamRole
            name: firehose-glue-role
            fieldPath: status.outputs.role_arn

    # Source record backup — keep original JSON for reprocessing
    s3BackupMode: Enabled
    s3Backup:
      bucketArn:
        valueFrom:
          kind: AwsS3Bucket
          name: clickstream-source-backup
          fieldPath: status.outputs.bucket_arn
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: firehose-backup-role
          fieldPath: status.outputs.role_arn
      prefix: "source-backup/clickstream/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/"
      compressionFormat: GZIP

    # CloudWatch error logging
    logging:
      enabled: true
      logGroupName: /aws/kinesisfirehose/clickstream-pipeline
      logStreamName: S3Delivery
```
