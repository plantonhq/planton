# AwsKinesisFirehose

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsKinesisFirehoseSpec defines the desired configuration for an Amazon Kinesis
Data Firehose delivery stream.

Firehose is a fully managed service for loading streaming data into storage
and analytics destinations. It captures, transforms, and delivers data --
without writing any custom code.

Data enters a delivery stream from one of three sources:
- **Direct PUT** (default) -- applications call PutRecord/PutRecordBatch APIs
  directly to the delivery stream.
- **Kinesis Data Stream** -- Firehose reads from an existing Kinesis stream,
  acting as a consumer with automatic checkpointing and retry.
- **Amazon MSK** -- Firehose reads from a topic on an Amazon MSK cluster
  (provisioned or serverless), turning Kafka topics into delivery pipelines
  with zero consumer code.

The delivery stream transforms and routes data to exactly ONE destination.
The destination type is ForceNew -- once created, you cannot change it.

Eight destination types are supported:
- **Extended S3** -- data lake storage with optional Parquet/ORC conversion,
  dynamic partitioning, record transformation, and compression.
- **OpenSearch** -- direct indexing into an OpenSearch Service domain with
  index rotation, buffering, and VPC delivery.
- **OpenSearch Serverless** -- indexing into an OpenSearch Serverless
  collection endpoint.
- **HTTP Endpoint** -- generic HTTPS delivery to any endpoint (Datadog,
  New Relic, Sumo Logic, custom APIs).
- **Redshift** -- data warehouse loading via S3 intermediate + COPY command.
- **Splunk** -- delivery to a Splunk HTTP Event Collector (HEC) endpoint.
- **Snowflake** -- direct streaming into a Snowflake table via Snowpipe
  Streaming (key-pair authentication, optional PrivateLink).
- **Iceberg** -- delivery into Apache Iceberg tables managed by the AWS Glue
  Data Catalog, with optional per-record table routing and upserts.

Notes:
- The delivery stream name (from metadata.name) is ForceNew.
- Server-side encryption (SSE) is only for Direct PUT sources. When using
  a Kinesis stream or MSK source, the source handles encryption.
- Every non-S3 destination requires an S3 configuration for failed (or all)
  records. The Extended S3 destination IS the primary S3 target.
- Destinations that authenticate with a credential (Redshift, Splunk,
  HTTP endpoint, Snowflake) can source it from AWS Secrets Manager instead
  of embedding it in the manifest -- prefer that for production.
- Credentials, region, and deployment workflow live outside this spec in
  stack inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsKinesisFirehose
metadata:
  name: test-firehose
  org: test-org
  env: dev
  id: test-firehose-dev
spec:
  region: us-west-2
  # Extended S3 destination: data lake storage with Direct PUT source.
  # This manifest is the offline plan/preview fixture for the arms no live
  # scenario can carry (the Glue-backed ORC conversion, the Hive JSON
  # deserializer, the backup-leg CloudWatch logging, and the Lambda
  # processor's dedicated invocation role) -- the parquet/OpenX tuning arms
  # ride the 04-s3-parquet-analytics preset.
  extendedS3:
    bucketArn:
      value: arn:aws:s3:::my-data-lake-bucket
    roleArn:
      value: arn:aws:iam::123456789012:role/firehose-s3-role
    prefix: firehose/events/
    # UNCOMPRESSED because format conversion is enabled below: ORC carries
    # its own columnar compression, and S3-level compression must stay off.
    compressionFormat: UNCOMPRESSED
    fileExtension: .orc
    buffering:
      intervalInSeconds: 60
      sizeInMbs: 64
    # Lambda transformation with a fractional buffer (AWS-legal 0.2-3 MiB)
    # and a dedicated invocation role distinct from the delivery role.
    processing:
      enabled: true
      processors:
        - lambda:
            lambdaArn:
              value: arn:aws:lambda:us-west-2:123456789012:function:enrich-events
            bufferSizeInMbs: 0.5
            roleArn:
              value: arn:aws:iam::123456789012:role/firehose-lambda-invoke
    # JSON -> ORC conversion: Hive JSON deserializer with explicit timestamp
    # patterns, ORC serializer tuned with bloom filters for point lookups.
    dataFormatConversion:
      enabled: true
      hiveJson:
        timestampFormats:
          - "yyyy-MM-dd'T'HH:mm:ss"
          - millis
      orc:
        compression: ZLIB
        stripeSizeBytes: 16777216
        bloomFilterColumns:
          - customer_id
        bloomFilterFalsePositiveProbability: 0.01
        formatVersion: V0_12
        rowIndexStride: 5000
      schema:
        databaseName: analytics_db
        tableName: events_orc
        roleArn:
          value: arn:aws:iam::123456789012:role/firehose-glue-access-role
    # Source-record backup with CloudWatch logging on the backup S3 leg --
    # the s3-level logging block, distinct from destination-level logging.
    s3BackupMode: Enabled
    s3Backup:
      bucketArn:
        value: arn:aws:s3:::my-data-lake-backup-bucket
      roleArn:
        value: arn:aws:iam::123456789012:role/firehose-s3-role
      prefix: firehose/raw/
      compressionFormat: GZIP
      logging:
        enabled: true
        logGroupName: /aws/kinesisfirehose/test-firehose
        logStreamName: S3BackupDelivery
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.kinesisStreamSource` | `AwsKinesisFirehoseKinesisStreamSource` |  |  |  |
| `spec.kinesisStreamSource.streamArn` | `string \| valueFrom` | yes |  | AwsKinesisStream (`status.outputs.stream_arn`) |
| `spec.kinesisStreamSource.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.mskSource` | `AwsKinesisFirehoseMskSource` |  |  |  |
| `spec.mskSource.mskClusterArn` | `string \| valueFrom` | yes |  | AwsMskCluster (`status.outputs.cluster_arn`) |
| `spec.mskSource.topicName` | `string` | yes |  |  |
| `spec.mskSource.connectivity` | `string` | yes |  |  |
| `spec.mskSource.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.mskSource.readFromTimestamp` | `string` |  |  |  |
| `spec.sseEnabled` | `bool` |  |  |  |
| `spec.sseKmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.extendedS3` | `AwsKinesisFirehoseExtendedS3Destination` |  |  |  |
| `spec.extendedS3.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.extendedS3.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.extendedS3.prefix` | `string` |  |  |  |
| `spec.extendedS3.errorOutputPrefix` | `string` |  |  |  |
| `spec.extendedS3.compressionFormat` | `string` |  |  |  |
| `spec.extendedS3.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.extendedS3.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.extendedS3.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.extendedS3.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.extendedS3.customTimeZone` | `string` |  |  |  |
| `spec.extendedS3.fileExtension` | `string` |  |  |  |
| `spec.extendedS3.s3BackupMode` | `string` |  |  |  |
| `spec.extendedS3.s3Backup` | `AwsKinesisFirehoseS3Config` |  |  |  |
| `spec.extendedS3.s3Backup.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.extendedS3.s3Backup.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.extendedS3.s3Backup.prefix` | `string` |  |  |  |
| `spec.extendedS3.s3Backup.errorOutputPrefix` | `string` |  |  |  |
| `spec.extendedS3.s3Backup.compressionFormat` | `string` |  |  |  |
| `spec.extendedS3.s3Backup.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.extendedS3.s3Backup.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.extendedS3.s3Backup.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.extendedS3.s3Backup.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.extendedS3.s3Backup.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.extendedS3.s3Backup.logging.enabled` | `bool` |  |  |  |
| `spec.extendedS3.s3Backup.logging.logGroupName` | `string` |  |  |  |
| `spec.extendedS3.s3Backup.logging.logStreamName` | `string` |  |  |  |
| `spec.extendedS3.processing` | `AwsKinesisFirehoseProcessing` |  |  |  |
| `spec.extendedS3.processing.enabled` | `bool` |  |  |  |
| `spec.extendedS3.processing.processors` | `[]AwsKinesisFirehoseProcessor` |  |  |  |
| `spec.extendedS3.processing.processors[].lambda` | `AwsKinesisFirehoseLambdaProcessor` |  |  |  |
| `spec.extendedS3.processing.processors[].lambda.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.extendedS3.processing.processors[].lambda.bufferSizeInMbs` | `double` |  |  |  |
| `spec.extendedS3.processing.processors[].lambda.bufferIntervalInSeconds` | `int32` |  |  |  |
| `spec.extendedS3.processing.processors[].lambda.numberOfRetries` | `int32` |  |  |  |
| `spec.extendedS3.processing.processors[].lambda.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.extendedS3.processing.processors[].metadataExtraction` | `AwsKinesisFirehoseMetadataExtractionProcessor` |  |  |  |
| `spec.extendedS3.processing.processors[].metadataExtraction.query` | `string` | yes |  |  |
| `spec.extendedS3.processing.processors[].metadataExtraction.jsonParsingEngine` | `string` |  |  |  |
| `spec.extendedS3.processing.processors[].decompression` | `AwsKinesisFirehoseDecompressionProcessor` |  |  |  |
| `spec.extendedS3.processing.processors[].decompression.compressionFormat` | `string` | yes |  |  |
| `spec.extendedS3.processing.processors[].cloudwatchLogProcessing` | `AwsKinesisFirehoseCloudwatchLogProcessingProcessor` |  |  |  |
| `spec.extendedS3.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction` | `bool` |  |  |  |
| `spec.extendedS3.processing.processors[].appendDelimiter` | `AwsKinesisFirehoseAppendDelimiterProcessor` |  |  |  |
| `spec.extendedS3.processing.processors[].appendDelimiter.delimiter` | `string` | yes |  |  |
| `spec.extendedS3.processing.processors[].recordDeaggregation` | `AwsKinesisFirehoseRecordDeaggregationProcessor` |  |  |  |
| `spec.extendedS3.processing.processors[].recordDeaggregation.subRecordType` | `string` | yes |  |  |
| `spec.extendedS3.processing.processors[].recordDeaggregation.delimiter` | `string` |  |  |  |
| `spec.extendedS3.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.extendedS3.logging.enabled` | `bool` |  |  |  |
| `spec.extendedS3.logging.logGroupName` | `string` |  |  |  |
| `spec.extendedS3.logging.logStreamName` | `string` |  |  |  |
| `spec.extendedS3.dynamicPartitioning` | `AwsKinesisFirehoseDynamicPartitioning` |  |  |  |
| `spec.extendedS3.dynamicPartitioning.enabled` | `bool` |  |  |  |
| `spec.extendedS3.dynamicPartitioning.retryDurationInSeconds` | `int32` |  |  |  |
| `spec.extendedS3.dataFormatConversion` | `AwsKinesisFirehoseDataFormatConversion` |  |  |  |
| `spec.extendedS3.dataFormatConversion.enabled` | `bool` |  |  |  |
| `spec.extendedS3.dataFormatConversion.openXJson` | `AwsKinesisFirehoseOpenXJsonDeserializer` |  |  |  |
| `spec.extendedS3.dataFormatConversion.openXJson.caseInsensitive` | `bool` |  |  |  |
| `spec.extendedS3.dataFormatConversion.openXJson.columnToJsonKeyMappings` | `map<string, string>` |  |  |  |
| `spec.extendedS3.dataFormatConversion.openXJson.convertDotsInJsonKeysToUnderscores` | `bool` |  |  |  |
| `spec.extendedS3.dataFormatConversion.hiveJson` | `AwsKinesisFirehoseHiveJsonDeserializer` |  |  |  |
| `spec.extendedS3.dataFormatConversion.hiveJson.timestampFormats` | `[]string` |  |  |  |
| `spec.extendedS3.dataFormatConversion.parquet` | `AwsKinesisFirehoseParquetSerializer` |  |  |  |
| `spec.extendedS3.dataFormatConversion.parquet.compression` | `string` |  |  |  |
| `spec.extendedS3.dataFormatConversion.parquet.blockSizeBytes` | `int64` |  |  |  |
| `spec.extendedS3.dataFormatConversion.parquet.pageSizeBytes` | `int64` |  |  |  |
| `spec.extendedS3.dataFormatConversion.parquet.maxPaddingBytes` | `int64` |  |  |  |
| `spec.extendedS3.dataFormatConversion.parquet.enableDictionaryCompression` | `bool` |  |  |  |
| `spec.extendedS3.dataFormatConversion.parquet.writerVersion` | `string` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc` | `AwsKinesisFirehoseOrcSerializer` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc.compression` | `string` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc.blockSizeBytes` | `int64` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc.stripeSizeBytes` | `int64` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc.bloomFilterColumns` | `[]string` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc.bloomFilterFalsePositiveProbability` | `double` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc.dictionaryKeyThreshold` | `double` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc.enablePadding` | `bool` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc.paddingTolerance` | `double` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc.formatVersion` | `string` |  |  |  |
| `spec.extendedS3.dataFormatConversion.orc.rowIndexStride` | `int32` |  |  |  |
| `spec.extendedS3.dataFormatConversion.schema` | `AwsKinesisFirehoseGlueSchemaConfig` |  |  |  |
| `spec.extendedS3.dataFormatConversion.schema.databaseName` | `string` | yes |  |  |
| `spec.extendedS3.dataFormatConversion.schema.tableName` | `string` | yes |  |  |
| `spec.extendedS3.dataFormatConversion.schema.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.extendedS3.dataFormatConversion.schema.catalogId` | `string` |  |  |  |
| `spec.extendedS3.dataFormatConversion.schema.region` | `string` |  |  |  |
| `spec.extendedS3.dataFormatConversion.schema.versionId` | `string` |  |  |  |
| `spec.opensearch` | `AwsKinesisFirehoseOpenSearchDestination` |  |  |  |
| `spec.opensearch.domainArn` | `string \| valueFrom` |  |  | AwsOpenSearchDomain (`status.outputs.domain_arn`) |
| `spec.opensearch.clusterEndpoint` | `string` |  |  |  |
| `spec.opensearch.indexName` | `string` | yes |  |  |
| `spec.opensearch.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.opensearch.indexRotationPeriod` | `string` |  |  |  |
| `spec.opensearch.typeName` | `string` |  |  |  |
| `spec.opensearch.defaultDocumentIdFormat` | `string` |  |  |  |
| `spec.opensearch.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.opensearch.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.opensearch.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.opensearch.retryDurationInSeconds` | `int32` |  |  |  |
| `spec.opensearch.s3BackupMode` | `string` |  |  |  |
| `spec.opensearch.s3Config` | `AwsKinesisFirehoseS3Config` | yes |  |  |
| `spec.opensearch.s3Config.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.opensearch.s3Config.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.opensearch.s3Config.prefix` | `string` |  |  |  |
| `spec.opensearch.s3Config.errorOutputPrefix` | `string` |  |  |  |
| `spec.opensearch.s3Config.compressionFormat` | `string` |  |  |  |
| `spec.opensearch.s3Config.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.opensearch.s3Config.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.opensearch.s3Config.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.opensearch.s3Config.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.opensearch.s3Config.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.opensearch.s3Config.logging.enabled` | `bool` |  |  |  |
| `spec.opensearch.s3Config.logging.logGroupName` | `string` |  |  |  |
| `spec.opensearch.s3Config.logging.logStreamName` | `string` |  |  |  |
| `spec.opensearch.processing` | `AwsKinesisFirehoseProcessing` |  |  |  |
| `spec.opensearch.processing.enabled` | `bool` |  |  |  |
| `spec.opensearch.processing.processors` | `[]AwsKinesisFirehoseProcessor` |  |  |  |
| `spec.opensearch.processing.processors[].lambda` | `AwsKinesisFirehoseLambdaProcessor` |  |  |  |
| `spec.opensearch.processing.processors[].lambda.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.opensearch.processing.processors[].lambda.bufferSizeInMbs` | `double` |  |  |  |
| `spec.opensearch.processing.processors[].lambda.bufferIntervalInSeconds` | `int32` |  |  |  |
| `spec.opensearch.processing.processors[].lambda.numberOfRetries` | `int32` |  |  |  |
| `spec.opensearch.processing.processors[].lambda.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.opensearch.processing.processors[].metadataExtraction` | `AwsKinesisFirehoseMetadataExtractionProcessor` |  |  |  |
| `spec.opensearch.processing.processors[].metadataExtraction.query` | `string` | yes |  |  |
| `spec.opensearch.processing.processors[].metadataExtraction.jsonParsingEngine` | `string` |  |  |  |
| `spec.opensearch.processing.processors[].decompression` | `AwsKinesisFirehoseDecompressionProcessor` |  |  |  |
| `spec.opensearch.processing.processors[].decompression.compressionFormat` | `string` | yes |  |  |
| `spec.opensearch.processing.processors[].cloudwatchLogProcessing` | `AwsKinesisFirehoseCloudwatchLogProcessingProcessor` |  |  |  |
| `spec.opensearch.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction` | `bool` |  |  |  |
| `spec.opensearch.processing.processors[].appendDelimiter` | `AwsKinesisFirehoseAppendDelimiterProcessor` |  |  |  |
| `spec.opensearch.processing.processors[].appendDelimiter.delimiter` | `string` | yes |  |  |
| `spec.opensearch.processing.processors[].recordDeaggregation` | `AwsKinesisFirehoseRecordDeaggregationProcessor` |  |  |  |
| `spec.opensearch.processing.processors[].recordDeaggregation.subRecordType` | `string` | yes |  |  |
| `spec.opensearch.processing.processors[].recordDeaggregation.delimiter` | `string` |  |  |  |
| `spec.opensearch.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.opensearch.logging.enabled` | `bool` |  |  |  |
| `spec.opensearch.logging.logGroupName` | `string` |  |  |  |
| `spec.opensearch.logging.logStreamName` | `string` |  |  |  |
| `spec.opensearch.vpcConfig` | `AwsKinesisFirehoseVpcConfig` |  |  |  |
| `spec.opensearch.vpcConfig.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.opensearch.vpcConfig.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.opensearch.vpcConfig.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.opensearchServerless` | `AwsKinesisFirehoseOpenSearchServerlessDestination` |  |  |  |
| `spec.opensearchServerless.collectionEndpoint` | `string` | yes |  |  |
| `spec.opensearchServerless.indexName` | `string` | yes |  |  |
| `spec.opensearchServerless.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.opensearchServerless.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.opensearchServerless.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.opensearchServerless.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.opensearchServerless.retryDurationInSeconds` | `int32` |  |  |  |
| `spec.opensearchServerless.s3BackupMode` | `string` |  |  |  |
| `spec.opensearchServerless.s3Config` | `AwsKinesisFirehoseS3Config` | yes |  |  |
| `spec.opensearchServerless.s3Config.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.opensearchServerless.s3Config.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.opensearchServerless.s3Config.prefix` | `string` |  |  |  |
| `spec.opensearchServerless.s3Config.errorOutputPrefix` | `string` |  |  |  |
| `spec.opensearchServerless.s3Config.compressionFormat` | `string` |  |  |  |
| `spec.opensearchServerless.s3Config.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.opensearchServerless.s3Config.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.opensearchServerless.s3Config.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.opensearchServerless.s3Config.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.opensearchServerless.s3Config.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.opensearchServerless.s3Config.logging.enabled` | `bool` |  |  |  |
| `spec.opensearchServerless.s3Config.logging.logGroupName` | `string` |  |  |  |
| `spec.opensearchServerless.s3Config.logging.logStreamName` | `string` |  |  |  |
| `spec.opensearchServerless.processing` | `AwsKinesisFirehoseProcessing` |  |  |  |
| `spec.opensearchServerless.processing.enabled` | `bool` |  |  |  |
| `spec.opensearchServerless.processing.processors` | `[]AwsKinesisFirehoseProcessor` |  |  |  |
| `spec.opensearchServerless.processing.processors[].lambda` | `AwsKinesisFirehoseLambdaProcessor` |  |  |  |
| `spec.opensearchServerless.processing.processors[].lambda.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.opensearchServerless.processing.processors[].lambda.bufferSizeInMbs` | `double` |  |  |  |
| `spec.opensearchServerless.processing.processors[].lambda.bufferIntervalInSeconds` | `int32` |  |  |  |
| `spec.opensearchServerless.processing.processors[].lambda.numberOfRetries` | `int32` |  |  |  |
| `spec.opensearchServerless.processing.processors[].lambda.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.opensearchServerless.processing.processors[].metadataExtraction` | `AwsKinesisFirehoseMetadataExtractionProcessor` |  |  |  |
| `spec.opensearchServerless.processing.processors[].metadataExtraction.query` | `string` | yes |  |  |
| `spec.opensearchServerless.processing.processors[].metadataExtraction.jsonParsingEngine` | `string` |  |  |  |
| `spec.opensearchServerless.processing.processors[].decompression` | `AwsKinesisFirehoseDecompressionProcessor` |  |  |  |
| `spec.opensearchServerless.processing.processors[].decompression.compressionFormat` | `string` | yes |  |  |
| `spec.opensearchServerless.processing.processors[].cloudwatchLogProcessing` | `AwsKinesisFirehoseCloudwatchLogProcessingProcessor` |  |  |  |
| `spec.opensearchServerless.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction` | `bool` |  |  |  |
| `spec.opensearchServerless.processing.processors[].appendDelimiter` | `AwsKinesisFirehoseAppendDelimiterProcessor` |  |  |  |
| `spec.opensearchServerless.processing.processors[].appendDelimiter.delimiter` | `string` | yes |  |  |
| `spec.opensearchServerless.processing.processors[].recordDeaggregation` | `AwsKinesisFirehoseRecordDeaggregationProcessor` |  |  |  |
| `spec.opensearchServerless.processing.processors[].recordDeaggregation.subRecordType` | `string` | yes |  |  |
| `spec.opensearchServerless.processing.processors[].recordDeaggregation.delimiter` | `string` |  |  |  |
| `spec.opensearchServerless.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.opensearchServerless.logging.enabled` | `bool` |  |  |  |
| `spec.opensearchServerless.logging.logGroupName` | `string` |  |  |  |
| `spec.opensearchServerless.logging.logStreamName` | `string` |  |  |  |
| `spec.opensearchServerless.vpcConfig` | `AwsKinesisFirehoseVpcConfig` |  |  |  |
| `spec.opensearchServerless.vpcConfig.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.opensearchServerless.vpcConfig.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.opensearchServerless.vpcConfig.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.httpEndpoint` | `AwsKinesisFirehoseHttpEndpointDestination` |  |  |  |
| `spec.httpEndpoint.url` | `string` | yes |  |  |
| `spec.httpEndpoint.name` | `string` |  |  |  |
| `spec.httpEndpoint.accessKey` | `string` (sensitive) |  |  |  |
| `spec.httpEndpoint.secretsManager` | `AwsKinesisFirehoseSecretsManagerConfig` |  |  |  |
| `spec.httpEndpoint.secretsManager.secretArn` | `string \| valueFrom` | yes |  |  |
| `spec.httpEndpoint.secretsManager.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.httpEndpoint.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.httpEndpoint.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.httpEndpoint.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.httpEndpoint.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.httpEndpoint.retryDurationInSeconds` | `int32` |  |  |  |
| `spec.httpEndpoint.s3BackupMode` | `string` |  |  |  |
| `spec.httpEndpoint.s3Config` | `AwsKinesisFirehoseS3Config` | yes |  |  |
| `spec.httpEndpoint.s3Config.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.httpEndpoint.s3Config.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.httpEndpoint.s3Config.prefix` | `string` |  |  |  |
| `spec.httpEndpoint.s3Config.errorOutputPrefix` | `string` |  |  |  |
| `spec.httpEndpoint.s3Config.compressionFormat` | `string` |  |  |  |
| `spec.httpEndpoint.s3Config.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.httpEndpoint.s3Config.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.httpEndpoint.s3Config.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.httpEndpoint.s3Config.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.httpEndpoint.s3Config.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.httpEndpoint.s3Config.logging.enabled` | `bool` |  |  |  |
| `spec.httpEndpoint.s3Config.logging.logGroupName` | `string` |  |  |  |
| `spec.httpEndpoint.s3Config.logging.logStreamName` | `string` |  |  |  |
| `spec.httpEndpoint.processing` | `AwsKinesisFirehoseProcessing` |  |  |  |
| `spec.httpEndpoint.processing.enabled` | `bool` |  |  |  |
| `spec.httpEndpoint.processing.processors` | `[]AwsKinesisFirehoseProcessor` |  |  |  |
| `spec.httpEndpoint.processing.processors[].lambda` | `AwsKinesisFirehoseLambdaProcessor` |  |  |  |
| `spec.httpEndpoint.processing.processors[].lambda.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.httpEndpoint.processing.processors[].lambda.bufferSizeInMbs` | `double` |  |  |  |
| `spec.httpEndpoint.processing.processors[].lambda.bufferIntervalInSeconds` | `int32` |  |  |  |
| `spec.httpEndpoint.processing.processors[].lambda.numberOfRetries` | `int32` |  |  |  |
| `spec.httpEndpoint.processing.processors[].lambda.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.httpEndpoint.processing.processors[].metadataExtraction` | `AwsKinesisFirehoseMetadataExtractionProcessor` |  |  |  |
| `spec.httpEndpoint.processing.processors[].metadataExtraction.query` | `string` | yes |  |  |
| `spec.httpEndpoint.processing.processors[].metadataExtraction.jsonParsingEngine` | `string` |  |  |  |
| `spec.httpEndpoint.processing.processors[].decompression` | `AwsKinesisFirehoseDecompressionProcessor` |  |  |  |
| `spec.httpEndpoint.processing.processors[].decompression.compressionFormat` | `string` | yes |  |  |
| `spec.httpEndpoint.processing.processors[].cloudwatchLogProcessing` | `AwsKinesisFirehoseCloudwatchLogProcessingProcessor` |  |  |  |
| `spec.httpEndpoint.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction` | `bool` |  |  |  |
| `spec.httpEndpoint.processing.processors[].appendDelimiter` | `AwsKinesisFirehoseAppendDelimiterProcessor` |  |  |  |
| `spec.httpEndpoint.processing.processors[].appendDelimiter.delimiter` | `string` | yes |  |  |
| `spec.httpEndpoint.processing.processors[].recordDeaggregation` | `AwsKinesisFirehoseRecordDeaggregationProcessor` |  |  |  |
| `spec.httpEndpoint.processing.processors[].recordDeaggregation.subRecordType` | `string` | yes |  |  |
| `spec.httpEndpoint.processing.processors[].recordDeaggregation.delimiter` | `string` |  |  |  |
| `spec.httpEndpoint.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.httpEndpoint.logging.enabled` | `bool` |  |  |  |
| `spec.httpEndpoint.logging.logGroupName` | `string` |  |  |  |
| `spec.httpEndpoint.logging.logStreamName` | `string` |  |  |  |
| `spec.httpEndpoint.requestConfig` | `AwsKinesisFirehoseRequestConfig` |  |  |  |
| `spec.httpEndpoint.requestConfig.contentEncoding` | `string` |  |  |  |
| `spec.httpEndpoint.requestConfig.commonAttributes` | `[]AwsKinesisFirehoseRequestAttribute` |  |  |  |
| `spec.httpEndpoint.requestConfig.commonAttributes[].name` | `string` | yes |  |  |
| `spec.httpEndpoint.requestConfig.commonAttributes[].value` | `string` | yes |  |  |
| `spec.redshift` | `AwsKinesisFirehoseRedshiftDestination` |  |  |  |
| `spec.redshift.clusterJdbcurl` | `string` | yes |  |  |
| `spec.redshift.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.redshift.dataTableName` | `string` | yes |  |  |
| `spec.redshift.dataTableColumns` | `string` |  |  |  |
| `spec.redshift.copyOptions` | `string` |  |  |  |
| `spec.redshift.username` | `string` |  |  |  |
| `spec.redshift.password` | `string` (sensitive) |  |  |  |
| `spec.redshift.secretsManager` | `AwsKinesisFirehoseSecretsManagerConfig` |  |  |  |
| `spec.redshift.secretsManager.secretArn` | `string \| valueFrom` | yes |  |  |
| `spec.redshift.secretsManager.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.redshift.s3Config` | `AwsKinesisFirehoseS3Config` | yes |  |  |
| `spec.redshift.s3Config.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.redshift.s3Config.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.redshift.s3Config.prefix` | `string` |  |  |  |
| `spec.redshift.s3Config.errorOutputPrefix` | `string` |  |  |  |
| `spec.redshift.s3Config.compressionFormat` | `string` |  |  |  |
| `spec.redshift.s3Config.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.redshift.s3Config.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.redshift.s3Config.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.redshift.s3Config.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.redshift.s3Config.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.redshift.s3Config.logging.enabled` | `bool` |  |  |  |
| `spec.redshift.s3Config.logging.logGroupName` | `string` |  |  |  |
| `spec.redshift.s3Config.logging.logStreamName` | `string` |  |  |  |
| `spec.redshift.retryDurationInSeconds` | `int32` |  |  |  |
| `spec.redshift.s3BackupMode` | `string` |  |  |  |
| `spec.redshift.s3Backup` | `AwsKinesisFirehoseS3Config` |  |  |  |
| `spec.redshift.s3Backup.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.redshift.s3Backup.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.redshift.s3Backup.prefix` | `string` |  |  |  |
| `spec.redshift.s3Backup.errorOutputPrefix` | `string` |  |  |  |
| `spec.redshift.s3Backup.compressionFormat` | `string` |  |  |  |
| `spec.redshift.s3Backup.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.redshift.s3Backup.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.redshift.s3Backup.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.redshift.s3Backup.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.redshift.s3Backup.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.redshift.s3Backup.logging.enabled` | `bool` |  |  |  |
| `spec.redshift.s3Backup.logging.logGroupName` | `string` |  |  |  |
| `spec.redshift.s3Backup.logging.logStreamName` | `string` |  |  |  |
| `spec.redshift.processing` | `AwsKinesisFirehoseProcessing` |  |  |  |
| `spec.redshift.processing.enabled` | `bool` |  |  |  |
| `spec.redshift.processing.processors` | `[]AwsKinesisFirehoseProcessor` |  |  |  |
| `spec.redshift.processing.processors[].lambda` | `AwsKinesisFirehoseLambdaProcessor` |  |  |  |
| `spec.redshift.processing.processors[].lambda.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.redshift.processing.processors[].lambda.bufferSizeInMbs` | `double` |  |  |  |
| `spec.redshift.processing.processors[].lambda.bufferIntervalInSeconds` | `int32` |  |  |  |
| `spec.redshift.processing.processors[].lambda.numberOfRetries` | `int32` |  |  |  |
| `spec.redshift.processing.processors[].lambda.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.redshift.processing.processors[].metadataExtraction` | `AwsKinesisFirehoseMetadataExtractionProcessor` |  |  |  |
| `spec.redshift.processing.processors[].metadataExtraction.query` | `string` | yes |  |  |
| `spec.redshift.processing.processors[].metadataExtraction.jsonParsingEngine` | `string` |  |  |  |
| `spec.redshift.processing.processors[].decompression` | `AwsKinesisFirehoseDecompressionProcessor` |  |  |  |
| `spec.redshift.processing.processors[].decompression.compressionFormat` | `string` | yes |  |  |
| `spec.redshift.processing.processors[].cloudwatchLogProcessing` | `AwsKinesisFirehoseCloudwatchLogProcessingProcessor` |  |  |  |
| `spec.redshift.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction` | `bool` |  |  |  |
| `spec.redshift.processing.processors[].appendDelimiter` | `AwsKinesisFirehoseAppendDelimiterProcessor` |  |  |  |
| `spec.redshift.processing.processors[].appendDelimiter.delimiter` | `string` | yes |  |  |
| `spec.redshift.processing.processors[].recordDeaggregation` | `AwsKinesisFirehoseRecordDeaggregationProcessor` |  |  |  |
| `spec.redshift.processing.processors[].recordDeaggregation.subRecordType` | `string` | yes |  |  |
| `spec.redshift.processing.processors[].recordDeaggregation.delimiter` | `string` |  |  |  |
| `spec.redshift.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.redshift.logging.enabled` | `bool` |  |  |  |
| `spec.redshift.logging.logGroupName` | `string` |  |  |  |
| `spec.redshift.logging.logStreamName` | `string` |  |  |  |
| `spec.splunk` | `AwsKinesisFirehoseSplunkDestination` |  |  |  |
| `spec.splunk.hecEndpoint` | `string` | yes |  |  |
| `spec.splunk.hecEndpointType` | `string` |  |  |  |
| `spec.splunk.hecToken` | `string` (sensitive) |  |  |  |
| `spec.splunk.secretsManager` | `AwsKinesisFirehoseSecretsManagerConfig` |  |  |  |
| `spec.splunk.secretsManager.secretArn` | `string \| valueFrom` | yes |  |  |
| `spec.splunk.secretsManager.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.splunk.hecAcknowledgmentTimeoutInSeconds` | `int32` |  |  |  |
| `spec.splunk.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.splunk.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.splunk.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.splunk.retryDurationInSeconds` | `int32` |  |  |  |
| `spec.splunk.s3BackupMode` | `string` |  |  |  |
| `spec.splunk.s3Config` | `AwsKinesisFirehoseS3Config` | yes |  |  |
| `spec.splunk.s3Config.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.splunk.s3Config.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.splunk.s3Config.prefix` | `string` |  |  |  |
| `spec.splunk.s3Config.errorOutputPrefix` | `string` |  |  |  |
| `spec.splunk.s3Config.compressionFormat` | `string` |  |  |  |
| `spec.splunk.s3Config.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.splunk.s3Config.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.splunk.s3Config.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.splunk.s3Config.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.splunk.s3Config.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.splunk.s3Config.logging.enabled` | `bool` |  |  |  |
| `spec.splunk.s3Config.logging.logGroupName` | `string` |  |  |  |
| `spec.splunk.s3Config.logging.logStreamName` | `string` |  |  |  |
| `spec.splunk.processing` | `AwsKinesisFirehoseProcessing` |  |  |  |
| `spec.splunk.processing.enabled` | `bool` |  |  |  |
| `spec.splunk.processing.processors` | `[]AwsKinesisFirehoseProcessor` |  |  |  |
| `spec.splunk.processing.processors[].lambda` | `AwsKinesisFirehoseLambdaProcessor` |  |  |  |
| `spec.splunk.processing.processors[].lambda.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.splunk.processing.processors[].lambda.bufferSizeInMbs` | `double` |  |  |  |
| `spec.splunk.processing.processors[].lambda.bufferIntervalInSeconds` | `int32` |  |  |  |
| `spec.splunk.processing.processors[].lambda.numberOfRetries` | `int32` |  |  |  |
| `spec.splunk.processing.processors[].lambda.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.splunk.processing.processors[].metadataExtraction` | `AwsKinesisFirehoseMetadataExtractionProcessor` |  |  |  |
| `spec.splunk.processing.processors[].metadataExtraction.query` | `string` | yes |  |  |
| `spec.splunk.processing.processors[].metadataExtraction.jsonParsingEngine` | `string` |  |  |  |
| `spec.splunk.processing.processors[].decompression` | `AwsKinesisFirehoseDecompressionProcessor` |  |  |  |
| `spec.splunk.processing.processors[].decompression.compressionFormat` | `string` | yes |  |  |
| `spec.splunk.processing.processors[].cloudwatchLogProcessing` | `AwsKinesisFirehoseCloudwatchLogProcessingProcessor` |  |  |  |
| `spec.splunk.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction` | `bool` |  |  |  |
| `spec.splunk.processing.processors[].appendDelimiter` | `AwsKinesisFirehoseAppendDelimiterProcessor` |  |  |  |
| `spec.splunk.processing.processors[].appendDelimiter.delimiter` | `string` | yes |  |  |
| `spec.splunk.processing.processors[].recordDeaggregation` | `AwsKinesisFirehoseRecordDeaggregationProcessor` |  |  |  |
| `spec.splunk.processing.processors[].recordDeaggregation.subRecordType` | `string` | yes |  |  |
| `spec.splunk.processing.processors[].recordDeaggregation.delimiter` | `string` |  |  |  |
| `spec.splunk.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.splunk.logging.enabled` | `bool` |  |  |  |
| `spec.splunk.logging.logGroupName` | `string` |  |  |  |
| `spec.splunk.logging.logStreamName` | `string` |  |  |  |
| `spec.snowflake` | `AwsKinesisFirehoseSnowflakeDestination` |  |  |  |
| `spec.snowflake.accountUrl` | `string` | yes |  |  |
| `spec.snowflake.database` | `string` | yes |  |  |
| `spec.snowflake.schema` | `string` | yes |  |  |
| `spec.snowflake.table` | `string` | yes |  |  |
| `spec.snowflake.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.snowflake.user` | `string` |  |  |  |
| `spec.snowflake.privateKey` | `string` (sensitive) |  |  |  |
| `spec.snowflake.keyPassphrase` | `string` (sensitive) |  |  |  |
| `spec.snowflake.secretsManager` | `AwsKinesisFirehoseSecretsManagerConfig` |  |  |  |
| `spec.snowflake.secretsManager.secretArn` | `string \| valueFrom` | yes |  |  |
| `spec.snowflake.secretsManager.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.snowflake.dataLoadingOption` | `string` |  |  |  |
| `spec.snowflake.contentColumnName` | `string` |  |  |  |
| `spec.snowflake.metadataColumnName` | `string` |  |  |  |
| `spec.snowflake.snowflakeRole` | `string` |  |  |  |
| `spec.snowflake.privateLinkVpceId` | `string` |  |  |  |
| `spec.snowflake.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.snowflake.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.snowflake.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.snowflake.retryDurationInSeconds` | `int32` |  |  |  |
| `spec.snowflake.s3BackupMode` | `string` |  |  |  |
| `spec.snowflake.s3Config` | `AwsKinesisFirehoseS3Config` | yes |  |  |
| `spec.snowflake.s3Config.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.snowflake.s3Config.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.snowflake.s3Config.prefix` | `string` |  |  |  |
| `spec.snowflake.s3Config.errorOutputPrefix` | `string` |  |  |  |
| `spec.snowflake.s3Config.compressionFormat` | `string` |  |  |  |
| `spec.snowflake.s3Config.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.snowflake.s3Config.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.snowflake.s3Config.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.snowflake.s3Config.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.snowflake.s3Config.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.snowflake.s3Config.logging.enabled` | `bool` |  |  |  |
| `spec.snowflake.s3Config.logging.logGroupName` | `string` |  |  |  |
| `spec.snowflake.s3Config.logging.logStreamName` | `string` |  |  |  |
| `spec.snowflake.processing` | `AwsKinesisFirehoseProcessing` |  |  |  |
| `spec.snowflake.processing.enabled` | `bool` |  |  |  |
| `spec.snowflake.processing.processors` | `[]AwsKinesisFirehoseProcessor` |  |  |  |
| `spec.snowflake.processing.processors[].lambda` | `AwsKinesisFirehoseLambdaProcessor` |  |  |  |
| `spec.snowflake.processing.processors[].lambda.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.snowflake.processing.processors[].lambda.bufferSizeInMbs` | `double` |  |  |  |
| `spec.snowflake.processing.processors[].lambda.bufferIntervalInSeconds` | `int32` |  |  |  |
| `spec.snowflake.processing.processors[].lambda.numberOfRetries` | `int32` |  |  |  |
| `spec.snowflake.processing.processors[].lambda.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.snowflake.processing.processors[].metadataExtraction` | `AwsKinesisFirehoseMetadataExtractionProcessor` |  |  |  |
| `spec.snowflake.processing.processors[].metadataExtraction.query` | `string` | yes |  |  |
| `spec.snowflake.processing.processors[].metadataExtraction.jsonParsingEngine` | `string` |  |  |  |
| `spec.snowflake.processing.processors[].decompression` | `AwsKinesisFirehoseDecompressionProcessor` |  |  |  |
| `spec.snowflake.processing.processors[].decompression.compressionFormat` | `string` | yes |  |  |
| `spec.snowflake.processing.processors[].cloudwatchLogProcessing` | `AwsKinesisFirehoseCloudwatchLogProcessingProcessor` |  |  |  |
| `spec.snowflake.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction` | `bool` |  |  |  |
| `spec.snowflake.processing.processors[].appendDelimiter` | `AwsKinesisFirehoseAppendDelimiterProcessor` |  |  |  |
| `spec.snowflake.processing.processors[].appendDelimiter.delimiter` | `string` | yes |  |  |
| `spec.snowflake.processing.processors[].recordDeaggregation` | `AwsKinesisFirehoseRecordDeaggregationProcessor` |  |  |  |
| `spec.snowflake.processing.processors[].recordDeaggregation.subRecordType` | `string` | yes |  |  |
| `spec.snowflake.processing.processors[].recordDeaggregation.delimiter` | `string` |  |  |  |
| `spec.snowflake.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.snowflake.logging.enabled` | `bool` |  |  |  |
| `spec.snowflake.logging.logGroupName` | `string` |  |  |  |
| `spec.snowflake.logging.logStreamName` | `string` |  |  |  |
| `spec.iceberg` | `AwsKinesisFirehoseIcebergDestination` |  |  |  |
| `spec.iceberg.catalogArn` | `string \| valueFrom` | yes |  |  |
| `spec.iceberg.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.iceberg.destinationTables` | `[]AwsKinesisFirehoseIcebergDestinationTable` | yes |  |  |
| `spec.iceberg.destinationTables[].databaseName` | `string` | yes |  |  |
| `spec.iceberg.destinationTables[].tableName` | `string` | yes |  |  |
| `spec.iceberg.destinationTables[].s3ErrorOutputPrefix` | `string` |  |  |  |
| `spec.iceberg.destinationTables[].uniqueKeys` | `[]string` |  |  |  |
| `spec.iceberg.appendOnly` | `bool` |  |  |  |
| `spec.iceberg.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.iceberg.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.iceberg.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.iceberg.retryDurationInSeconds` | `int32` |  |  |  |
| `spec.iceberg.s3BackupMode` | `string` |  |  |  |
| `spec.iceberg.s3Config` | `AwsKinesisFirehoseS3Config` | yes |  |  |
| `spec.iceberg.s3Config.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.iceberg.s3Config.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.iceberg.s3Config.prefix` | `string` |  |  |  |
| `spec.iceberg.s3Config.errorOutputPrefix` | `string` |  |  |  |
| `spec.iceberg.s3Config.compressionFormat` | `string` |  |  |  |
| `spec.iceberg.s3Config.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.iceberg.s3Config.buffering` | `AwsKinesisFirehoseBufferingHints` |  |  |  |
| `spec.iceberg.s3Config.buffering.intervalInSeconds` | `int32` |  |  |  |
| `spec.iceberg.s3Config.buffering.sizeInMbs` | `int32` |  |  |  |
| `spec.iceberg.s3Config.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.iceberg.s3Config.logging.enabled` | `bool` |  |  |  |
| `spec.iceberg.s3Config.logging.logGroupName` | `string` |  |  |  |
| `spec.iceberg.s3Config.logging.logStreamName` | `string` |  |  |  |
| `spec.iceberg.processing` | `AwsKinesisFirehoseProcessing` |  |  |  |
| `spec.iceberg.processing.enabled` | `bool` |  |  |  |
| `spec.iceberg.processing.processors` | `[]AwsKinesisFirehoseProcessor` |  |  |  |
| `spec.iceberg.processing.processors[].lambda` | `AwsKinesisFirehoseLambdaProcessor` |  |  |  |
| `spec.iceberg.processing.processors[].lambda.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.iceberg.processing.processors[].lambda.bufferSizeInMbs` | `double` |  |  |  |
| `spec.iceberg.processing.processors[].lambda.bufferIntervalInSeconds` | `int32` |  |  |  |
| `spec.iceberg.processing.processors[].lambda.numberOfRetries` | `int32` |  |  |  |
| `spec.iceberg.processing.processors[].lambda.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.iceberg.processing.processors[].metadataExtraction` | `AwsKinesisFirehoseMetadataExtractionProcessor` |  |  |  |
| `spec.iceberg.processing.processors[].metadataExtraction.query` | `string` | yes |  |  |
| `spec.iceberg.processing.processors[].metadataExtraction.jsonParsingEngine` | `string` |  |  |  |
| `spec.iceberg.processing.processors[].decompression` | `AwsKinesisFirehoseDecompressionProcessor` |  |  |  |
| `spec.iceberg.processing.processors[].decompression.compressionFormat` | `string` | yes |  |  |
| `spec.iceberg.processing.processors[].cloudwatchLogProcessing` | `AwsKinesisFirehoseCloudwatchLogProcessingProcessor` |  |  |  |
| `spec.iceberg.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction` | `bool` |  |  |  |
| `spec.iceberg.processing.processors[].appendDelimiter` | `AwsKinesisFirehoseAppendDelimiterProcessor` |  |  |  |
| `spec.iceberg.processing.processors[].appendDelimiter.delimiter` | `string` | yes |  |  |
| `spec.iceberg.processing.processors[].recordDeaggregation` | `AwsKinesisFirehoseRecordDeaggregationProcessor` |  |  |  |
| `spec.iceberg.processing.processors[].recordDeaggregation.subRecordType` | `string` | yes |  |  |
| `spec.iceberg.processing.processors[].recordDeaggregation.delimiter` | `string` |  |  |  |
| `spec.iceberg.logging` | `AwsKinesisFirehoseCloudwatchLogging` |  |  |  |
| `spec.iceberg.logging.enabled` | `bool` |  |  |  |
| `spec.iceberg.logging.logGroupName` | `string` |  |  |  |
| `spec.iceberg.logging.logStreamName` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.kinesisStreamSource

`AwsKinesisFirehoseKinesisStreamSource`

Kinesis Data Stream source configuration. When set, Firehose reads from
the specified stream instead of accepting Direct PUT calls. The entire
source configuration is ForceNew -- it cannot be changed after creation.

Mutually exclusive with msk_source. When a stream source is configured,
server-side encryption (sse_enabled) must NOT be set -- the source stream
handles its own encryption.

### spec.kinesisStreamSource.streamArn

`string | valueFrom` · required

ARN of the Kinesis Data Stream to read from. Firehose creates an internal
consumer and reads all shards. The stream must exist before the delivery
stream is created.

- references: AwsKinesisStream (`status.outputs.stream_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKinesisStream, name: <that resource's name>, fieldPath: status.outputs.stream_arn}} -- a bare string does not parse

### spec.kinesisStreamSource.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to read from the Kinesis
stream. The role must have kinesis:GetRecords, kinesis:GetShardIterator,
kinesis:DescribeStream, and kinesis:ListShards permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.mskSource

`AwsKinesisFirehoseMskSource`

Amazon MSK source configuration. When set, Firehose reads from a topic
on the specified MSK cluster instead of accepting Direct PUT calls. The
entire source configuration is ForceNew -- it cannot be changed after
creation.

Mutually exclusive with kinesis_stream_source. When an MSK source is
configured, server-side encryption (sse_enabled) must NOT be set -- the
source cluster handles its own encryption.

- rule: connectivity must be 'PRIVATE' or 'PUBLIC'
- rule: read_from_timestamp must be an RFC 3339 timestamp (e.g., '2026-05-01T00:00:00Z') when set

### spec.mskSource.mskClusterArn

`string | valueFrom` · required

ARN of the MSK cluster to read from. The cluster must exist and have
IAM access control enabled before the delivery stream is created.

- references: AwsMskCluster (`status.outputs.cluster_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsMskCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_arn}} -- a bare string does not parse

### spec.mskSource.topicName

`string` · required

Name of the Kafka topic to read from. The topic must exist on the
cluster before the delivery stream is created.

- rule: {"string":{"minLen":"1"}}

### spec.mskSource.connectivity

`string` · required

How Firehose connects to the MSK cluster:
- "PRIVATE" -- connect through the cluster's private brokers inside its
  VPC (the common case for provisioned and serverless clusters).
- "PUBLIC" -- connect through the cluster's public endpoints (requires
  public access to be enabled on the cluster).

- rule: {"string":{"minLen":"1"}}

### spec.mskSource.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to connect to and read from
the MSK cluster. The role must have kafka:GetBootstrapBrokers,
kafka:DescribeCluster, kafka:DescribeClusterV2, and the kafka-cluster:*
data-plane permissions (Connect, DescribeTopic, ReadData,
DescribeGroup) on the cluster, topic, and consumer group.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.mskSource.readFromTimestamp

`string`

Start reading the topic from this point in time instead of the current
offset. RFC 3339 format (e.g., "2026-05-01T00:00:00Z"). When absent,
Firehose starts from the latest offset at creation time.

### spec.sseEnabled

`bool`

Enable server-side encryption for data at rest in the delivery stream
buffer. Only valid for Direct PUT sources -- when using a Kinesis stream
or MSK source, encryption is handled by the source.

When true and sse_kms_key_arn is absent, uses the AWS-owned CMK.
When true and sse_kms_key_arn is present, uses a customer-managed CMK.

### spec.sseKmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for server-side encryption. When set,
Firehose uses this key instead of the AWS-owned CMK. Requires
sse_enabled to be true.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.extendedS3

`AwsKinesisFirehoseExtendedS3Destination`

Extended S3 destination for data lake storage. Supports compression,
record transformation, dynamic partitioning, and Parquet/ORC format
conversion via AWS Glue Data Catalog. The most feature-rich destination.

- rule: compression_format must be 'UNCOMPRESSED', 'GZIP', 'ZIP', 'Snappy', or 'HADOOP_SNAPPY' when set
- rule: s3_backup_mode must be 'Disabled' or 'Enabled' when set
- rule: s3_backup requires s3_backup_mode to be 'Enabled'

### spec.extendedS3.bucketArn

`string | valueFrom` · required

S3 bucket ARN where records are delivered.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.extendedS3.roleArn

`string | valueFrom` · required

IAM role ARN granting Firehose write access to the S3 bucket, KMS key
(if encrypted), Lambda function (if processing), and Glue catalog (if
format conversion is enabled).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.extendedS3.prefix

`string`

S3 key prefix prepended to every delivered object. Supports Firehose
expression syntax for dynamic prefixes:
  "data/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/"

When dynamic partitioning is enabled, use partitioning keys:
  "data/customer=!{partitionKeyFromQuery:customer_id}/"

### spec.extendedS3.errorOutputPrefix

`string`

S3 key prefix for records that fail transformation or delivery.

### spec.extendedS3.compressionFormat

`string`

Compression format applied before writing to S3. When data format
conversion is enabled, compression is applied to the converted
(Parquet/ORC) output -- in that case, use the format-native compression
(configured in data_format_conversion) and leave this as UNCOMPRESSED.

Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
Default: "UNCOMPRESSED".

### spec.extendedS3.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.extendedS3.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for S3 delivery. Default: 300s interval, 5 MiB size.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.extendedS3.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.extendedS3.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.extendedS3.customTimeZone

`string`

IANA time zone for S3 prefix timestamp expressions.
Default: "UTC". Example: "US/Eastern", "Europe/London".

### spec.extendedS3.fileExtension

`string`

File extension appended to delivered S3 objects (e.g., ".json", ".parquet").
Must start with a period. When data format conversion is enabled, the
extension is typically set to match the output format.

### spec.extendedS3.s3BackupMode

`string`

S3 backup mode for source records. When "Enabled", a copy of the original
(pre-transformation) records is written to s3_backup in addition to the
primary destination. Useful for auditing and reprocessing.

Valid values: "Disabled" (default), "Enabled".

### spec.extendedS3.s3Backup

`AwsKinesisFirehoseS3Config`

S3 configuration for source record backup. Required when s3_backup_mode
is "Enabled".

- rule: compression_format must be 'UNCOMPRESSED', 'GZIP', 'ZIP', 'Snappy', or 'HADOOP_SNAPPY' when set

### spec.extendedS3.s3Backup.bucketArn

`string | valueFrom` · required

S3 bucket ARN where records are delivered.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.extendedS3.s3Backup.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to write to the S3 bucket.
The role must have s3:PutObject, s3:AbortMultipartUpload,
s3:GetBucketLocation, and s3:ListBucket permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.extendedS3.s3Backup.prefix

`string`

S3 key prefix prepended to delivered objects. Supports Firehose expression
syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").

### spec.extendedS3.s3Backup.errorOutputPrefix

`string`

S3 key prefix for error output. When Firehose cannot deliver or transform
a record, it writes to this prefix. Uses the same expression syntax as prefix.

### spec.extendedS3.s3Backup.compressionFormat

`string`

Compression format for delivered objects. Applied before writing to S3.
Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
Default: "UNCOMPRESSED".

### spec.extendedS3.s3Backup.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
When absent, S3 uses its default encryption settings (SSE-S3 or bucket
default encryption).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.extendedS3.s3Backup.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for S3 delivery.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.extendedS3.s3Backup.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.extendedS3.s3Backup.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.extendedS3.s3Backup.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch logging configuration for S3 delivery errors.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.extendedS3.s3Backup.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.extendedS3.s3Backup.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.extendedS3.s3Backup.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.extendedS3.processing

`AwsKinesisFirehoseProcessing`

Record-transformation pipeline. Applied before compression and format
conversion. For dynamic partitioning, include a metadata_extraction
processor to define the partition keys.

- rule: processors require enabled to be true
- rule: at least one processor is required when processing is enabled

### spec.extendedS3.processing.enabled

`bool`

Enable the processing pipeline. When true, at least one processor should
be configured.

### spec.extendedS3.processing.processors

`[]AwsKinesisFirehoseProcessor`

Ordered list of processors. Each entry configures exactly one processor
type; Firehose executes them in order.

- rule: exactly one processor must be set: lambda, metadata_extraction, decompression, cloudwatch_log_processing, append_delimiter, or record_deaggregation

### spec.extendedS3.processing.processors[].lambda

`AwsKinesisFirehoseLambdaProcessor`

Invoke an AWS Lambda function to transform records. The function
receives batches of records and returns transformed records with a
status (Ok, Dropped, ProcessingFailed) per record.

- rule: buffer_size_in_mbs must be between 0.2 and 3 when set
- rule: buffer_interval_in_seconds must be between 60 and 900 when set
- rule: number_of_retries must be between 0 and 300 when set

### spec.extendedS3.processing.processors[].lambda.lambdaArn

`string | valueFrom` · required

ARN of the Lambda function that transforms records. May include a
version or alias qualifier to pin the deployed transformation.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.extendedS3.processing.processors[].lambda.bufferSizeInMbs

`double`

Buffer size in MiB that Firehose accumulates before invoking Lambda.
Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
destination is Splunk).

Smaller buffers invoke Lambda more frequently with smaller batches.
Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
invocation costs.

### spec.extendedS3.processing.processors[].lambda.bufferIntervalInSeconds

`int32`

Buffer interval in seconds. Firehose invokes Lambda when this interval
elapses, even if the buffer size threshold has not been reached.
Range: 60-900 seconds. Default: 60 seconds.

### spec.extendedS3.processing.processors[].lambda.numberOfRetries

`int32`

Number of times Firehose retries a failed Lambda invocation before
writing the record to the error output prefix.
Range: 0-300. Default: 3.

### spec.extendedS3.processing.processors[].lambda.roleArn

`string | valueFrom`

IAM role ARN Firehose assumes to invoke the Lambda function. When
absent, Firehose uses the delivery stream's destination role -- the
right choice for almost every pipeline. Set this only when the
transformation function must be invoked with a DIFFERENT role than the
one that writes to the destination (e.g. the function lives in another
account). Note: AWS reports the delivery role back for unset values, and
the provider does not store default-valued processor parameters in
state -- so set this only to a non-default role, never to the delivery
role itself (that would cause perpetual plan diffs).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.extendedS3.processing.processors[].metadataExtraction

`AwsKinesisFirehoseMetadataExtractionProcessor`

Extract partition keys from JSON records with a JQ expression. Used
with Extended S3 dynamic partitioning -- extracted keys are referenced
in the S3 prefix as !{partitionKeyFromQuery:<key>}.

- rule: json_parsing_engine must be 'JQ-1.6' when set

### spec.extendedS3.processing.processors[].metadataExtraction.query

`string` · required

JQ expression that extracts partition keys from each JSON record.
The result must be an object whose keys become partition keys.

Example: "{customer_id: .customer_id, event_type: .type}" extracts two
keys, referenced in the prefix as
"data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".

- rule: {"string":{"minLen":"1"}}

### spec.extendedS3.processing.processors[].metadataExtraction.jsonParsingEngine

`string`

JSON parsing engine used to evaluate the query.
Valid value: "JQ-1.6" (default -- the only engine AWS supports today).

### spec.extendedS3.processing.processors[].decompression

`AwsKinesisFirehoseDecompressionProcessor`

Decompress GZIP-compressed records before delivery. Typically the first
processor when the source sends compressed payloads (e.g., CloudWatch
Logs subscription filters).

- rule: compression_format must be 'GZIP'

### spec.extendedS3.processing.processors[].decompression.compressionFormat

`string` · required

Compression format of the incoming records.
Valid value: "GZIP" (the only format AWS supports today).

- rule: {"string":{"minLen":"1"}}

### spec.extendedS3.processing.processors[].cloudwatchLogProcessing

`AwsKinesisFirehoseCloudwatchLogProcessingProcessor`

Unwrap CloudWatch Logs subscription envelopes into individual log
events. Use after a decompression processor when the source is a
CloudWatch Logs subscription filter.

### spec.extendedS3.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction

`bool`

When true, extract only the log event message field, discarding the
CloudWatch envelope metadata. When false, records pass through with the
envelope intact.

### spec.extendedS3.processing.processors[].appendDelimiter

`AwsKinesisFirehoseAppendDelimiterProcessor`

Append a delimiter to every record. Use to produce newline-delimited
JSON (JSON lines) output for query engines and log consumers.

Only supported on the extended_s3 destination -- delimiting is an
S3-object formatting concern; other destinations frame records
natively.

### spec.extendedS3.processing.processors[].appendDelimiter.delimiter

`string` · required

Delimiter appended to each record. Use "\\n" for newline-delimited
output -- the format Athena, Spark, and most log consumers expect.

- rule: {"string":{"minLen":"1"}}

### spec.extendedS3.processing.processors[].recordDeaggregation

`AwsKinesisFirehoseRecordDeaggregationProcessor`

Split multi-record aggregates (e.g., KPL-aggregated or delimited
payloads) into individual records before further processing. Required
before dynamic partitioning when producers aggregate records.

Only supported on the extended_s3 destination -- AWS rejects it for
every other destination type at creation.

- rule: sub_record_type must be 'JSON' or 'DELIMITED'
- rule: delimiter is required when sub_record_type is 'DELIMITED'

### spec.extendedS3.processing.processors[].recordDeaggregation.subRecordType

`string` · required

How records are aggregated in the payload:
- "JSON" -- concatenated JSON documents (no delimiter needed).
- "DELIMITED" -- records separated by a custom delimiter; requires
  the delimiter field.

- rule: {"string":{"minLen":"1"}}

### spec.extendedS3.processing.processors[].recordDeaggregation.delimiter

`string`

Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
newline). Required when sub_record_type is "DELIMITED".

### spec.extendedS3.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch error logging for S3 delivery failures.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.extendedS3.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.extendedS3.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.extendedS3.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.extendedS3.dynamicPartitioning

`AwsKinesisFirehoseDynamicPartitioning`

Dynamic partitioning configuration. Enables partitioning delivered data
by record fields (e.g., customer_id, event_type) for efficient querying
with Athena, Spark, or Presto. ForceNew -- cannot be enabled/disabled
after creation. Define the partition keys with a metadata_extraction
processor (or a Lambda processor emitting partition metadata) and
reference them in the prefix.

- rule: retry_duration_in_seconds must be between 0 and 7200 when set

### spec.extendedS3.dynamicPartitioning.enabled

`bool`

Enable dynamic partitioning. ForceNew -- cannot be changed after creation.
When enabled, configure partition key expressions in the S3 prefix using
!{partitionKeyFromQuery:...} or !{partitionKeyFromLambda:...} syntax,
and define the keys with a metadata_extraction (or Lambda) processor.

### spec.extendedS3.dynamicPartitioning.retryDurationInSeconds

`int32`

Duration in seconds that Firehose retries delivery when a partition key
expression fails or the S3 PutObject call is throttled.
Range: 0-7200. Default: 300 seconds.

### spec.extendedS3.dataFormatConversion

`AwsKinesisFirehoseDataFormatConversion`

Data format conversion from JSON to columnar formats (Parquet or ORC)
using an AWS Glue Data Catalog schema. Dramatically improves query
performance and reduces storage cost for analytics workloads.

- rule: a deserializer arm (open_x_json or hive_json) is required when data format conversion is enabled
- rule: a serializer arm (parquet or orc) is required when data format conversion is enabled
- rule: schema is required when data format conversion is enabled

### spec.extendedS3.dataFormatConversion.enabled

`bool`

Enable data format conversion. When true, a deserializer arm, a
serializer arm, and schema are required. When false with arms configured, the
conversion settings are retained but inactive (AWS permits disabling
conversion without discarding its configuration).

### spec.extendedS3.dataFormatConversion.openXJson

`AwsKinesisFirehoseOpenXJsonDeserializer`

OpenX JSON deserializer arm. Handles most JSON formats including nested
objects -- the right choice for general use.

### spec.extendedS3.dataFormatConversion.openXJson.caseInsensitive

`bool` · optional (explicit presence)

When true (the AWS default), JSON keys are lowercased before
deserialization, so keys match case-insensitively against the Glue
schema's lowercase column names. Set to false only when the source
JSON keys must be matched exactly as sent (combine with
column_to_json_key_mappings for columns whose keys differ only by
case).

### spec.extendedS3.dataFormatConversion.openXJson.columnToJsonKeyMappings

`map<string, string>`

Map of Glue schema column names to JSON keys that differ from them.
Use when a JSON key is not a valid column name -- e.g. a key that
collides with a Hive reserved word: {"ts": "timestamp"} reads the JSON
key "timestamp" into the column "ts".

### spec.extendedS3.dataFormatConversion.openXJson.convertDotsInJsonKeysToUnderscores

`bool`

When true, dots in JSON keys are converted to underscores before
matching against the schema ("a.b" reads into column "a_b"). Glue
column names cannot contain dots, so enable this when source keys do.
Default: false.

### spec.extendedS3.dataFormatConversion.hiveJson

`AwsKinesisFirehoseHiveJsonDeserializer`

Apache Hive JSON deserializer arm. Use for Hive-compatible JSON when
records carry non-standard timestamp encodings that need explicit
parsing patterns.

### spec.extendedS3.dataFormatConversion.hiveJson.timestampFormats

`[]string`

Joda-Time datetime format patterns for parsing timestamp fields from
the source JSON (e.g. "yyyy-MM-dd'T'HH:mm:ss"). Include the special
value "millis" to parse epoch-millisecond timestamps. When empty,
Firehose uses java.sql.Timestamp::valueOf.

### spec.extendedS3.dataFormatConversion.parquet

`AwsKinesisFirehoseParquetSerializer`

Apache Parquet serializer arm. Best for read-heavy analytical workloads
(Athena, Spark, Presto): excellent compression, predicate pushdown, and
columnar pruning.

- rule: compression must be 'SNAPPY', 'GZIP', or 'UNCOMPRESSED' when set
- rule: block_size_bytes must be at least 67108864 (64 MiB) when set
- rule: page_size_bytes must be at least 65536 (64 KiB) when set
- rule: max_padding_bytes must not be negative
- rule: writer_version must be 'V1' or 'V2' when set

### spec.extendedS3.dataFormatConversion.parquet.compression

`string`

Compression codec for Parquet pages. "SNAPPY" (the default) balances
speed and size; "GZIP" compresses harder at higher CPU cost --
preferable when S3 storage/scan cost outweighs write throughput;
"UNCOMPRESSED" only when downstream readers cannot decompress.

### spec.extendedS3.dataFormatConversion.parquet.blockSizeBytes

`int64`

Parquet row-group (block) size in bytes. Firehose and query engines use
it for padding and row-group sizing; larger blocks improve scan
efficiency at the cost of memory. Minimum: 67108864 (64 MiB).
Default: 268435456 (256 MiB).

### spec.extendedS3.dataFormatConversion.parquet.pageSizeBytes

`int64`

Parquet page size in bytes -- the smallest unit a read must fully
decompress. Minimum: 65536 (64 KiB). Default: 1048576 (1 MiB).

### spec.extendedS3.dataFormatConversion.parquet.maxPaddingBytes

`int64`

Maximum padding in bytes when writing row groups (used to align blocks
to HDFS-style boundaries; rarely needed on S3). Default: 0.

### spec.extendedS3.dataFormatConversion.parquet.enableDictionaryCompression

`bool`

Enable dictionary compression -- encodes repeated column values through
a dictionary. Effective when columns carry low-cardinality values.
Default: false.

### spec.extendedS3.dataFormatConversion.parquet.writerVersion

`string`

Parquet writer version. "V1" (the default) is readable by every
engine; "V2" enables newer encodings -- confirm downstream reader
support before switching. Valid values: "V1", "V2".

### spec.extendedS3.dataFormatConversion.orc

`AwsKinesisFirehoseOrcSerializer`

Apache ORC serializer arm. Best for Hive workloads: ACID support, bloom
filters, and built-in indexing.

- rule: compression must be 'SNAPPY', 'ZLIB', or 'NONE' when set
- rule: block_size_bytes must be at least 67108864 (64 MiB) when set
- rule: stripe_size_bytes must be at least 8388608 (8 MiB) when set
- rule: bloom_filter_false_positive_probability must be between 0 and 1
- rule: dictionary_key_threshold must be between 0 and 1
- rule: padding_tolerance must be between 0 and 1
- rule: format_version must be 'V0_11' or 'V0_12' when set
- rule: row_index_stride must be at least 1000 when set

### spec.extendedS3.dataFormatConversion.orc.compression

`string`

Compression codec for ORC stripes. "SNAPPY" (the default) balances
speed and size; "ZLIB" compresses harder at higher CPU cost; "NONE"
only when downstream readers cannot decompress.

### spec.extendedS3.dataFormatConversion.orc.blockSizeBytes

`int64`

ORC block size in bytes, used for padding calculations and copy-block
sizing. Minimum: 67108864 (64 MiB). Default: 268435456 (256 MiB).

### spec.extendedS3.dataFormatConversion.orc.stripeSizeBytes

`int64`

ORC stripe size in bytes -- the unit of independent reading. Minimum:
8388608 (8 MiB). Default: 67108864 (64 MiB).

### spec.extendedS3.dataFormatConversion.orc.bloomFilterColumns

`[]string`

Column names to build bloom filters for. Bloom filters let readers
skip stripes that cannot contain a searched value -- list the columns
your queries filter on by equality.

### spec.extendedS3.dataFormatConversion.orc.bloomFilterFalsePositiveProbability

`double` · optional (explicit presence)

Bloom filter false-positive probability, between 0 and 1. Lower values
make filters more selective but larger. AWS default: 0.05. Explicit 0
is AWS-legal, so absence (AWS default) and 0 are distinct states.

### spec.extendedS3.dataFormatConversion.orc.dictionaryKeyThreshold

`double`

Fraction of a column's total distinct keys above which dictionary
encoding is abandoned for that column, between 0 and 1. 0 (the AWS
default) always dictionary-encodes; 1 never abandons it.

### spec.extendedS3.dataFormatConversion.orc.enablePadding

`bool`

Pad stripes to HDFS-style block boundaries. Relevant for HDFS-backed
readers; rarely needed for S3. Default: false.

### spec.extendedS3.dataFormatConversion.orc.paddingTolerance

`double` · optional (explicit presence)

Maximum fraction of a stripe that may be wasted as padding, between 0
and 1. Only meaningful when enable_padding is true. AWS default: 0.05.
Explicit 0 is AWS-legal, so absence and 0 are distinct states.

### spec.extendedS3.dataFormatConversion.orc.formatVersion

`string`

ORC file format version. "V0_12" (the default) is current; "V0_11" is
the legacy Hive 0.11 format -- use only for readers frozen on it.
Valid values: "V0_11", "V0_12".

### spec.extendedS3.dataFormatConversion.orc.rowIndexStride

`int32`

Number of rows between index entries. Must be at least 1000 when set.
Default: 10000.

### spec.extendedS3.dataFormatConversion.schema

`AwsKinesisFirehoseGlueSchemaConfig`

AWS Glue Data Catalog schema reference. Defines the table schema used
for converting JSON records to the columnar format. Required when
data format conversion is enabled.

### spec.extendedS3.dataFormatConversion.schema.databaseName

`string` · required

Glue Data Catalog database name containing the table.

- rule: {"string":{"minLen":"1"}}

### spec.extendedS3.dataFormatConversion.schema.tableName

`string` · required

Glue Data Catalog table name defining the record schema.

- rule: {"string":{"minLen":"1"}}

### spec.extendedS3.dataFormatConversion.schema.roleArn

`string | valueFrom` · required

IAM role ARN granting Firehose permission to access the Glue catalog.
Must have glue:GetTable and glue:GetTableVersions permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.extendedS3.dataFormatConversion.schema.catalogId

`string`

Glue Data Catalog ID (AWS account ID). When omitted, defaults to the
current AWS account.

### spec.extendedS3.dataFormatConversion.schema.region

`string`

AWS region of the Glue catalog. When omitted, defaults to the delivery
stream's region.

### spec.extendedS3.dataFormatConversion.schema.versionId

`string`

Table version to use. Default: "LATEST".

### spec.opensearch

`AwsKinesisFirehoseOpenSearchDestination`

OpenSearch destination for direct indexing into an Amazon OpenSearch
Service domain. Supports index rotation, VPC delivery, and record
transformation. Failed documents are backed up to S3.

- rule: exactly one of domain_arn or cluster_endpoint must be set
- rule: index_rotation_period must be 'NoRotation', 'OneHour', 'OneDay', 'OneWeek', or 'OneMonth' when set
- rule: default_document_id_format must be 'FIREHOSE_DEFAULT' or 'NO_DOCUMENT_ID' when set
- rule: s3_backup_mode must be 'FailedDocumentsOnly' or 'AllDocuments' when set
- rule: retry_duration_in_seconds must be between 0 and 7200 when set
- rule: buffering.size_in_mbs must not exceed 100 for OpenSearch destinations
- rule: record_deaggregation and append_delimiter processors are only supported on the extended_s3 destination

### spec.opensearch.domainArn

`string | valueFrom`

ARN of the OpenSearch domain. Mutually exclusive with cluster_endpoint.
Use this for domains managed within the same AWS account.

- references: AwsOpenSearchDomain (`status.outputs.domain_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsOpenSearchDomain, name: <that resource's name>, fieldPath: status.outputs.domain_arn}} -- a bare string does not parse

### spec.opensearch.clusterEndpoint

`string`

OpenSearch cluster endpoint URL. Mutually exclusive with domain_arn.
Use this for cross-account domains or non-standard endpoints.
Format: "https://search-domain-xxxx.us-east-1.es.amazonaws.com"

### spec.opensearch.indexName

`string` · required

Name of the OpenSearch index to deliver records to. Required.
When index_rotation_period is set, this becomes the index prefix and
Firehose appends a timestamp suffix (e.g., "logs-2026-02-15").

- rule: {"string":{"minLen":"1"}}

### spec.opensearch.roleArn

`string | valueFrom` · required

IAM role ARN granting Firehose permission to write to OpenSearch.
Must have es:ESHttpPut and es:ESHttpGet permissions on the domain.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.opensearch.indexRotationPeriod

`string`

Index rotation period. Firehose appends a timestamp suffix to index_name
and creates a new index at each rotation boundary.

Valid values: "NoRotation", "OneHour", "OneDay" (default), "OneWeek", "OneMonth".

"NoRotation" writes all records to the same index (use for small, static datasets).
"OneDay" is recommended for most log and analytics use cases.

### spec.opensearch.typeName

`string`

OpenSearch document type name. Only relevant for Elasticsearch 6.x and
earlier (OpenSearch does not use document types). Leave empty for
OpenSearch domains.

### spec.opensearch.defaultDocumentIdFormat

`string`

How Firehose assigns document IDs:
- "FIREHOSE_DEFAULT" -- Firehose generates a unique document ID per
  record. Protects against duplicates on retry but disables OpenSearch
  ID-based deduplication.
- "NO_DOCUMENT_ID" -- no ID is sent; OpenSearch auto-generates one.
  Improves indexing throughput and lets retried deliveries deduplicate.

When empty, AWS applies its service default (FIREHOSE_DEFAULT).

### spec.opensearch.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for OpenSearch delivery. Default: 300s interval, 5 MiB.
Maximum size for OpenSearch destinations: 100 MiB.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.opensearch.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.opensearch.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.opensearch.retryDurationInSeconds

`int32`

Retry duration in seconds for failed OpenSearch index requests.
Range: 0-7200. Default: 300 seconds.
Set to 0 to disable retries (failed documents go directly to S3 backup).

### spec.opensearch.s3BackupMode

`string`

S3 backup mode for documents. Controls when records are written to S3.

Valid values:
- "FailedDocumentsOnly" (default) -- only documents that fail indexing
  are backed up to S3.
- "AllDocuments" -- all documents are backed up to S3 in addition to
  being indexed in OpenSearch.

ForceNew -- changing the backup mode replaces the delivery stream.

### spec.opensearch.s3Config

`AwsKinesisFirehoseS3Config` · required

S3 configuration for backing up failed (or all) documents. Required.

- rule: {"required":true}
- rule: compression_format must be 'UNCOMPRESSED', 'GZIP', 'ZIP', 'Snappy', or 'HADOOP_SNAPPY' when set

### spec.opensearch.s3Config.bucketArn

`string | valueFrom` · required

S3 bucket ARN where records are delivered.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.opensearch.s3Config.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to write to the S3 bucket.
The role must have s3:PutObject, s3:AbortMultipartUpload,
s3:GetBucketLocation, and s3:ListBucket permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.opensearch.s3Config.prefix

`string`

S3 key prefix prepended to delivered objects. Supports Firehose expression
syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").

### spec.opensearch.s3Config.errorOutputPrefix

`string`

S3 key prefix for error output. When Firehose cannot deliver or transform
a record, it writes to this prefix. Uses the same expression syntax as prefix.

### spec.opensearch.s3Config.compressionFormat

`string`

Compression format for delivered objects. Applied before writing to S3.
Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
Default: "UNCOMPRESSED".

### spec.opensearch.s3Config.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
When absent, S3 uses its default encryption settings (SSE-S3 or bucket
default encryption).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.opensearch.s3Config.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for S3 delivery.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.opensearch.s3Config.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.opensearch.s3Config.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.opensearch.s3Config.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch logging configuration for S3 delivery errors.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.opensearch.s3Config.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.opensearch.s3Config.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.opensearch.s3Config.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.opensearch.processing

`AwsKinesisFirehoseProcessing`

Record-transformation pipeline applied before indexing.

- rule: processors require enabled to be true
- rule: at least one processor is required when processing is enabled

### spec.opensearch.processing.enabled

`bool`

Enable the processing pipeline. When true, at least one processor should
be configured.

### spec.opensearch.processing.processors

`[]AwsKinesisFirehoseProcessor`

Ordered list of processors. Each entry configures exactly one processor
type; Firehose executes them in order.

- rule: exactly one processor must be set: lambda, metadata_extraction, decompression, cloudwatch_log_processing, append_delimiter, or record_deaggregation

### spec.opensearch.processing.processors[].lambda

`AwsKinesisFirehoseLambdaProcessor`

Invoke an AWS Lambda function to transform records. The function
receives batches of records and returns transformed records with a
status (Ok, Dropped, ProcessingFailed) per record.

- rule: buffer_size_in_mbs must be between 0.2 and 3 when set
- rule: buffer_interval_in_seconds must be between 60 and 900 when set
- rule: number_of_retries must be between 0 and 300 when set

### spec.opensearch.processing.processors[].lambda.lambdaArn

`string | valueFrom` · required

ARN of the Lambda function that transforms records. May include a
version or alias qualifier to pin the deployed transformation.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.opensearch.processing.processors[].lambda.bufferSizeInMbs

`double`

Buffer size in MiB that Firehose accumulates before invoking Lambda.
Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
destination is Splunk).

Smaller buffers invoke Lambda more frequently with smaller batches.
Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
invocation costs.

### spec.opensearch.processing.processors[].lambda.bufferIntervalInSeconds

`int32`

Buffer interval in seconds. Firehose invokes Lambda when this interval
elapses, even if the buffer size threshold has not been reached.
Range: 60-900 seconds. Default: 60 seconds.

### spec.opensearch.processing.processors[].lambda.numberOfRetries

`int32`

Number of times Firehose retries a failed Lambda invocation before
writing the record to the error output prefix.
Range: 0-300. Default: 3.

### spec.opensearch.processing.processors[].lambda.roleArn

`string | valueFrom`

IAM role ARN Firehose assumes to invoke the Lambda function. When
absent, Firehose uses the delivery stream's destination role -- the
right choice for almost every pipeline. Set this only when the
transformation function must be invoked with a DIFFERENT role than the
one that writes to the destination (e.g. the function lives in another
account). Note: AWS reports the delivery role back for unset values, and
the provider does not store default-valued processor parameters in
state -- so set this only to a non-default role, never to the delivery
role itself (that would cause perpetual plan diffs).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.opensearch.processing.processors[].metadataExtraction

`AwsKinesisFirehoseMetadataExtractionProcessor`

Extract partition keys from JSON records with a JQ expression. Used
with Extended S3 dynamic partitioning -- extracted keys are referenced
in the S3 prefix as !{partitionKeyFromQuery:<key>}.

- rule: json_parsing_engine must be 'JQ-1.6' when set

### spec.opensearch.processing.processors[].metadataExtraction.query

`string` · required

JQ expression that extracts partition keys from each JSON record.
The result must be an object whose keys become partition keys.

Example: "{customer_id: .customer_id, event_type: .type}" extracts two
keys, referenced in the prefix as
"data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".

- rule: {"string":{"minLen":"1"}}

### spec.opensearch.processing.processors[].metadataExtraction.jsonParsingEngine

`string`

JSON parsing engine used to evaluate the query.
Valid value: "JQ-1.6" (default -- the only engine AWS supports today).

### spec.opensearch.processing.processors[].decompression

`AwsKinesisFirehoseDecompressionProcessor`

Decompress GZIP-compressed records before delivery. Typically the first
processor when the source sends compressed payloads (e.g., CloudWatch
Logs subscription filters).

- rule: compression_format must be 'GZIP'

### spec.opensearch.processing.processors[].decompression.compressionFormat

`string` · required

Compression format of the incoming records.
Valid value: "GZIP" (the only format AWS supports today).

- rule: {"string":{"minLen":"1"}}

### spec.opensearch.processing.processors[].cloudwatchLogProcessing

`AwsKinesisFirehoseCloudwatchLogProcessingProcessor`

Unwrap CloudWatch Logs subscription envelopes into individual log
events. Use after a decompression processor when the source is a
CloudWatch Logs subscription filter.

### spec.opensearch.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction

`bool`

When true, extract only the log event message field, discarding the
CloudWatch envelope metadata. When false, records pass through with the
envelope intact.

### spec.opensearch.processing.processors[].appendDelimiter

`AwsKinesisFirehoseAppendDelimiterProcessor`

Append a delimiter to every record. Use to produce newline-delimited
JSON (JSON lines) output for query engines and log consumers.

Only supported on the extended_s3 destination -- delimiting is an
S3-object formatting concern; other destinations frame records
natively.

### spec.opensearch.processing.processors[].appendDelimiter.delimiter

`string` · required

Delimiter appended to each record. Use "\\n" for newline-delimited
output -- the format Athena, Spark, and most log consumers expect.

- rule: {"string":{"minLen":"1"}}

### spec.opensearch.processing.processors[].recordDeaggregation

`AwsKinesisFirehoseRecordDeaggregationProcessor`

Split multi-record aggregates (e.g., KPL-aggregated or delimited
payloads) into individual records before further processing. Required
before dynamic partitioning when producers aggregate records.

Only supported on the extended_s3 destination -- AWS rejects it for
every other destination type at creation.

- rule: sub_record_type must be 'JSON' or 'DELIMITED'
- rule: delimiter is required when sub_record_type is 'DELIMITED'

### spec.opensearch.processing.processors[].recordDeaggregation.subRecordType

`string` · required

How records are aggregated in the payload:
- "JSON" -- concatenated JSON documents (no delimiter needed).
- "DELIMITED" -- records separated by a custom delimiter; requires
  the delimiter field.

- rule: {"string":{"minLen":"1"}}

### spec.opensearch.processing.processors[].recordDeaggregation.delimiter

`string`

Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
newline). Required when sub_record_type is "DELIMITED".

### spec.opensearch.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch error logging for OpenSearch delivery failures.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.opensearch.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.opensearch.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.opensearch.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.opensearch.vpcConfig

`AwsKinesisFirehoseVpcConfig`

VPC configuration for delivering to VPC-deployed OpenSearch domains.
ForceNew -- the VPC config cannot be changed after creation.
When absent, Firehose delivers over the public internet.

### spec.opensearch.vpcConfig.subnetIds

`[]string | valueFrom`

Subnet IDs where Firehose creates ENIs for VPC delivery. Provide at
least one subnet. For high availability, use subnets in multiple AZs.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.opensearch.vpcConfig.securityGroupIds

`[]string | valueFrom`

Security group IDs applied to the ENIs. Must allow outbound HTTPS (443)
traffic to the destination.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.opensearch.vpcConfig.roleArn

`string | valueFrom` · required

IAM role ARN for Firehose to manage VPC ENIs. The role must have
ec2:CreateNetworkInterface, ec2:DescribeNetworkInterfaces,
ec2:DeleteNetworkInterface, and ec2:DescribeVpcs permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.opensearchServerless

`AwsKinesisFirehoseOpenSearchServerlessDestination`

OpenSearch Serverless destination for indexing into an OpenSearch
Serverless collection. Failed documents are backed up to S3.

- rule: collection_endpoint must start with 'https://'
- rule: s3_backup_mode must be 'FailedDocumentsOnly' or 'AllDocuments' when set
- rule: retry_duration_in_seconds must be between 0 and 7200 when set
- rule: buffering.size_in_mbs must not exceed 100 for OpenSearch Serverless destinations
- rule: record_deaggregation and append_delimiter processors are only supported on the extended_s3 destination

### spec.opensearchServerless.collectionEndpoint

`string` · required

Endpoint of the OpenSearch Serverless collection.
Format: "https://<collection-id>.<region>.aoss.amazonaws.com"

- rule: {"string":{"minLen":"1"}}

### spec.opensearchServerless.indexName

`string` · required

Name of the index to deliver records to. Required. The index must be
permitted by the collection's data access policy for the delivery role.

- rule: {"string":{"minLen":"1"}}

### spec.opensearchServerless.roleArn

`string | valueFrom` · required

IAM role ARN granting Firehose permission to write to the collection.
The role must be granted aoss:APIAccessAll on the collection and be
listed in the collection's data access policy with document-write
permission on the index.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.opensearchServerless.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for delivery. Default: 300s interval, 5 MiB.
Maximum size for OpenSearch Serverless destinations: 100 MiB.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.opensearchServerless.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.opensearchServerless.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.opensearchServerless.retryDurationInSeconds

`int32`

Retry duration in seconds for failed index requests.
Range: 0-7200. Default: 300 seconds.
Set to 0 to disable retries (failed documents go directly to S3 backup).

### spec.opensearchServerless.s3BackupMode

`string`

S3 backup mode for documents. Controls when records are written to S3.

Valid values:
- "FailedDocumentsOnly" (default) -- only documents that fail indexing
  are backed up to S3.
- "AllDocuments" -- all documents are backed up to S3 in addition to
  being indexed.

ForceNew -- changing the backup mode replaces the delivery stream.

### spec.opensearchServerless.s3Config

`AwsKinesisFirehoseS3Config` · required

S3 configuration for backing up failed (or all) documents. Required.

- rule: {"required":true}
- rule: compression_format must be 'UNCOMPRESSED', 'GZIP', 'ZIP', 'Snappy', or 'HADOOP_SNAPPY' when set

### spec.opensearchServerless.s3Config.bucketArn

`string | valueFrom` · required

S3 bucket ARN where records are delivered.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.opensearchServerless.s3Config.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to write to the S3 bucket.
The role must have s3:PutObject, s3:AbortMultipartUpload,
s3:GetBucketLocation, and s3:ListBucket permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.opensearchServerless.s3Config.prefix

`string`

S3 key prefix prepended to delivered objects. Supports Firehose expression
syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").

### spec.opensearchServerless.s3Config.errorOutputPrefix

`string`

S3 key prefix for error output. When Firehose cannot deliver or transform
a record, it writes to this prefix. Uses the same expression syntax as prefix.

### spec.opensearchServerless.s3Config.compressionFormat

`string`

Compression format for delivered objects. Applied before writing to S3.
Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
Default: "UNCOMPRESSED".

### spec.opensearchServerless.s3Config.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
When absent, S3 uses its default encryption settings (SSE-S3 or bucket
default encryption).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.opensearchServerless.s3Config.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for S3 delivery.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.opensearchServerless.s3Config.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.opensearchServerless.s3Config.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.opensearchServerless.s3Config.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch logging configuration for S3 delivery errors.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.opensearchServerless.s3Config.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.opensearchServerless.s3Config.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.opensearchServerless.s3Config.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.opensearchServerless.processing

`AwsKinesisFirehoseProcessing`

Record-transformation pipeline applied before indexing.

- rule: processors require enabled to be true
- rule: at least one processor is required when processing is enabled

### spec.opensearchServerless.processing.enabled

`bool`

Enable the processing pipeline. When true, at least one processor should
be configured.

### spec.opensearchServerless.processing.processors

`[]AwsKinesisFirehoseProcessor`

Ordered list of processors. Each entry configures exactly one processor
type; Firehose executes them in order.

- rule: exactly one processor must be set: lambda, metadata_extraction, decompression, cloudwatch_log_processing, append_delimiter, or record_deaggregation

### spec.opensearchServerless.processing.processors[].lambda

`AwsKinesisFirehoseLambdaProcessor`

Invoke an AWS Lambda function to transform records. The function
receives batches of records and returns transformed records with a
status (Ok, Dropped, ProcessingFailed) per record.

- rule: buffer_size_in_mbs must be between 0.2 and 3 when set
- rule: buffer_interval_in_seconds must be between 60 and 900 when set
- rule: number_of_retries must be between 0 and 300 when set

### spec.opensearchServerless.processing.processors[].lambda.lambdaArn

`string | valueFrom` · required

ARN of the Lambda function that transforms records. May include a
version or alias qualifier to pin the deployed transformation.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.opensearchServerless.processing.processors[].lambda.bufferSizeInMbs

`double`

Buffer size in MiB that Firehose accumulates before invoking Lambda.
Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
destination is Splunk).

Smaller buffers invoke Lambda more frequently with smaller batches.
Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
invocation costs.

### spec.opensearchServerless.processing.processors[].lambda.bufferIntervalInSeconds

`int32`

Buffer interval in seconds. Firehose invokes Lambda when this interval
elapses, even if the buffer size threshold has not been reached.
Range: 60-900 seconds. Default: 60 seconds.

### spec.opensearchServerless.processing.processors[].lambda.numberOfRetries

`int32`

Number of times Firehose retries a failed Lambda invocation before
writing the record to the error output prefix.
Range: 0-300. Default: 3.

### spec.opensearchServerless.processing.processors[].lambda.roleArn

`string | valueFrom`

IAM role ARN Firehose assumes to invoke the Lambda function. When
absent, Firehose uses the delivery stream's destination role -- the
right choice for almost every pipeline. Set this only when the
transformation function must be invoked with a DIFFERENT role than the
one that writes to the destination (e.g. the function lives in another
account). Note: AWS reports the delivery role back for unset values, and
the provider does not store default-valued processor parameters in
state -- so set this only to a non-default role, never to the delivery
role itself (that would cause perpetual plan diffs).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.opensearchServerless.processing.processors[].metadataExtraction

`AwsKinesisFirehoseMetadataExtractionProcessor`

Extract partition keys from JSON records with a JQ expression. Used
with Extended S3 dynamic partitioning -- extracted keys are referenced
in the S3 prefix as !{partitionKeyFromQuery:<key>}.

- rule: json_parsing_engine must be 'JQ-1.6' when set

### spec.opensearchServerless.processing.processors[].metadataExtraction.query

`string` · required

JQ expression that extracts partition keys from each JSON record.
The result must be an object whose keys become partition keys.

Example: "{customer_id: .customer_id, event_type: .type}" extracts two
keys, referenced in the prefix as
"data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".

- rule: {"string":{"minLen":"1"}}

### spec.opensearchServerless.processing.processors[].metadataExtraction.jsonParsingEngine

`string`

JSON parsing engine used to evaluate the query.
Valid value: "JQ-1.6" (default -- the only engine AWS supports today).

### spec.opensearchServerless.processing.processors[].decompression

`AwsKinesisFirehoseDecompressionProcessor`

Decompress GZIP-compressed records before delivery. Typically the first
processor when the source sends compressed payloads (e.g., CloudWatch
Logs subscription filters).

- rule: compression_format must be 'GZIP'

### spec.opensearchServerless.processing.processors[].decompression.compressionFormat

`string` · required

Compression format of the incoming records.
Valid value: "GZIP" (the only format AWS supports today).

- rule: {"string":{"minLen":"1"}}

### spec.opensearchServerless.processing.processors[].cloudwatchLogProcessing

`AwsKinesisFirehoseCloudwatchLogProcessingProcessor`

Unwrap CloudWatch Logs subscription envelopes into individual log
events. Use after a decompression processor when the source is a
CloudWatch Logs subscription filter.

### spec.opensearchServerless.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction

`bool`

When true, extract only the log event message field, discarding the
CloudWatch envelope metadata. When false, records pass through with the
envelope intact.

### spec.opensearchServerless.processing.processors[].appendDelimiter

`AwsKinesisFirehoseAppendDelimiterProcessor`

Append a delimiter to every record. Use to produce newline-delimited
JSON (JSON lines) output for query engines and log consumers.

Only supported on the extended_s3 destination -- delimiting is an
S3-object formatting concern; other destinations frame records
natively.

### spec.opensearchServerless.processing.processors[].appendDelimiter.delimiter

`string` · required

Delimiter appended to each record. Use "\\n" for newline-delimited
output -- the format Athena, Spark, and most log consumers expect.

- rule: {"string":{"minLen":"1"}}

### spec.opensearchServerless.processing.processors[].recordDeaggregation

`AwsKinesisFirehoseRecordDeaggregationProcessor`

Split multi-record aggregates (e.g., KPL-aggregated or delimited
payloads) into individual records before further processing. Required
before dynamic partitioning when producers aggregate records.

Only supported on the extended_s3 destination -- AWS rejects it for
every other destination type at creation.

- rule: sub_record_type must be 'JSON' or 'DELIMITED'
- rule: delimiter is required when sub_record_type is 'DELIMITED'

### spec.opensearchServerless.processing.processors[].recordDeaggregation.subRecordType

`string` · required

How records are aggregated in the payload:
- "JSON" -- concatenated JSON documents (no delimiter needed).
- "DELIMITED" -- records separated by a custom delimiter; requires
  the delimiter field.

- rule: {"string":{"minLen":"1"}}

### spec.opensearchServerless.processing.processors[].recordDeaggregation.delimiter

`string`

Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
newline). Required when sub_record_type is "DELIMITED".

### spec.opensearchServerless.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch error logging for delivery failures.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.opensearchServerless.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.opensearchServerless.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.opensearchServerless.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.opensearchServerless.vpcConfig

`AwsKinesisFirehoseVpcConfig`

VPC configuration for delivering to collections reached through a VPC
endpoint. ForceNew -- the VPC config cannot be changed after creation.
When absent, Firehose delivers over the public internet.

### spec.opensearchServerless.vpcConfig.subnetIds

`[]string | valueFrom`

Subnet IDs where Firehose creates ENIs for VPC delivery. Provide at
least one subnet. For high availability, use subnets in multiple AZs.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.opensearchServerless.vpcConfig.securityGroupIds

`[]string | valueFrom`

Security group IDs applied to the ENIs. Must allow outbound HTTPS (443)
traffic to the destination.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.opensearchServerless.vpcConfig.roleArn

`string | valueFrom` · required

IAM role ARN for Firehose to manage VPC ENIs. The role must have
ec2:CreateNetworkInterface, ec2:DescribeNetworkInterfaces,
ec2:DeleteNetworkInterface, and ec2:DescribeVpcs permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.httpEndpoint

`AwsKinesisFirehoseHttpEndpointDestination`

HTTP endpoint destination for delivery to any HTTPS endpoint. Supports
custom headers, content encoding, and record transformation. Commonly
used for third-party integrations (Datadog, New Relic, Sumo Logic).
Failed deliveries are backed up to S3.

- rule: url must start with 'https://'
- rule: access_key and secrets_manager are mutually exclusive
- rule: s3_backup_mode must be 'FailedDataOnly' or 'AllData' when set
- rule: retry_duration_in_seconds must be between 0 and 7200 when set
- rule: buffering.size_in_mbs must not exceed 100 for HTTP endpoint destinations
- rule: record_deaggregation and append_delimiter processors are only supported on the extended_s3 destination

### spec.httpEndpoint.url

`string` · required

HTTPS URL of the destination endpoint. Must start with "https://".
Maximum length: 1000 characters.

Examples:
- "https://http-intake.logs.datadoghq.com/v1/input"
- "https://api.honeycomb.io/1/kinesis_events/your-dataset"
- "https://my-api.example.com/firehose"

- rule: {"string":{"minLen":"1"}}

### spec.httpEndpoint.name

`string`

Human-readable name for the endpoint. Appears in the AWS Console and
CloudWatch metrics. Maximum 256 characters.

### spec.httpEndpoint.accessKey

`string` · sensitive

Access key for endpoint authentication, sent in the
X-Amz-Firehose-Access-Key header. Sensitive -- treated as a secret.
Maximum 4096 characters.

Mutually exclusive with secrets_manager -- prefer Secrets Manager for
production so the key never appears in manifests or IaC state. Endpoints
that do not require authentication may omit both.

### spec.httpEndpoint.secretsManager

`AwsKinesisFirehoseSecretsManagerConfig`

Source the access key from AWS Secrets Manager instead of access_key.
The secret shape is {"api_key": "..."}. Setting this block enables
Secrets Manager authentication (ForceNew).

### spec.httpEndpoint.secretsManager.secretArn

`string | valueFrom` · required

ARN of the Secrets Manager secret holding the destination credential.
An ARN reference resolved by Firehose at delivery time -- never the
secret material itself.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.httpEndpoint.secretsManager.roleArn

`string | valueFrom`

IAM role ARN granting Firehose permission to read the secret
(secretsmanager:GetSecretValue). When absent, the destination's
delivery role is used.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.httpEndpoint.roleArn

`string | valueFrom`

IAM role ARN granting Firehose permission to deliver to the endpoint
and write to the S3 backup bucket. Optional -- the S3 configuration
carries its own delivery role.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.httpEndpoint.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for HTTP delivery. Default: 300s interval, 5 MiB.
Maximum size for HTTP endpoint destinations: 100 MiB.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.httpEndpoint.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.httpEndpoint.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.httpEndpoint.retryDurationInSeconds

`int32`

Retry duration in seconds for failed HTTP deliveries (non-2xx responses
or timeouts). Range: 0-7200. Default: 300 seconds.

### spec.httpEndpoint.s3BackupMode

`string`

S3 backup mode. Controls when records are written to S3.

Valid values:
- "FailedDataOnly" (default) -- only records that fail HTTP delivery
  are backed up to S3.
- "AllData" -- all records are backed up to S3 in addition to being
  sent to the HTTP endpoint.

### spec.httpEndpoint.s3Config

`AwsKinesisFirehoseS3Config` · required

S3 configuration for backing up failed (or all) records. Required.

- rule: {"required":true}
- rule: compression_format must be 'UNCOMPRESSED', 'GZIP', 'ZIP', 'Snappy', or 'HADOOP_SNAPPY' when set

### spec.httpEndpoint.s3Config.bucketArn

`string | valueFrom` · required

S3 bucket ARN where records are delivered.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.httpEndpoint.s3Config.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to write to the S3 bucket.
The role must have s3:PutObject, s3:AbortMultipartUpload,
s3:GetBucketLocation, and s3:ListBucket permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.httpEndpoint.s3Config.prefix

`string`

S3 key prefix prepended to delivered objects. Supports Firehose expression
syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").

### spec.httpEndpoint.s3Config.errorOutputPrefix

`string`

S3 key prefix for error output. When Firehose cannot deliver or transform
a record, it writes to this prefix. Uses the same expression syntax as prefix.

### spec.httpEndpoint.s3Config.compressionFormat

`string`

Compression format for delivered objects. Applied before writing to S3.
Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
Default: "UNCOMPRESSED".

### spec.httpEndpoint.s3Config.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
When absent, S3 uses its default encryption settings (SSE-S3 or bucket
default encryption).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.httpEndpoint.s3Config.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for S3 delivery.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.httpEndpoint.s3Config.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.httpEndpoint.s3Config.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.httpEndpoint.s3Config.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch logging configuration for S3 delivery errors.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.httpEndpoint.s3Config.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.httpEndpoint.s3Config.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.httpEndpoint.s3Config.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.httpEndpoint.processing

`AwsKinesisFirehoseProcessing`

Record-transformation pipeline applied before HTTP delivery.

- rule: processors require enabled to be true
- rule: at least one processor is required when processing is enabled

### spec.httpEndpoint.processing.enabled

`bool`

Enable the processing pipeline. When true, at least one processor should
be configured.

### spec.httpEndpoint.processing.processors

`[]AwsKinesisFirehoseProcessor`

Ordered list of processors. Each entry configures exactly one processor
type; Firehose executes them in order.

- rule: exactly one processor must be set: lambda, metadata_extraction, decompression, cloudwatch_log_processing, append_delimiter, or record_deaggregation

### spec.httpEndpoint.processing.processors[].lambda

`AwsKinesisFirehoseLambdaProcessor`

Invoke an AWS Lambda function to transform records. The function
receives batches of records and returns transformed records with a
status (Ok, Dropped, ProcessingFailed) per record.

- rule: buffer_size_in_mbs must be between 0.2 and 3 when set
- rule: buffer_interval_in_seconds must be between 60 and 900 when set
- rule: number_of_retries must be between 0 and 300 when set

### spec.httpEndpoint.processing.processors[].lambda.lambdaArn

`string | valueFrom` · required

ARN of the Lambda function that transforms records. May include a
version or alias qualifier to pin the deployed transformation.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.httpEndpoint.processing.processors[].lambda.bufferSizeInMbs

`double`

Buffer size in MiB that Firehose accumulates before invoking Lambda.
Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
destination is Splunk).

Smaller buffers invoke Lambda more frequently with smaller batches.
Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
invocation costs.

### spec.httpEndpoint.processing.processors[].lambda.bufferIntervalInSeconds

`int32`

Buffer interval in seconds. Firehose invokes Lambda when this interval
elapses, even if the buffer size threshold has not been reached.
Range: 60-900 seconds. Default: 60 seconds.

### spec.httpEndpoint.processing.processors[].lambda.numberOfRetries

`int32`

Number of times Firehose retries a failed Lambda invocation before
writing the record to the error output prefix.
Range: 0-300. Default: 3.

### spec.httpEndpoint.processing.processors[].lambda.roleArn

`string | valueFrom`

IAM role ARN Firehose assumes to invoke the Lambda function. When
absent, Firehose uses the delivery stream's destination role -- the
right choice for almost every pipeline. Set this only when the
transformation function must be invoked with a DIFFERENT role than the
one that writes to the destination (e.g. the function lives in another
account). Note: AWS reports the delivery role back for unset values, and
the provider does not store default-valued processor parameters in
state -- so set this only to a non-default role, never to the delivery
role itself (that would cause perpetual plan diffs).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.httpEndpoint.processing.processors[].metadataExtraction

`AwsKinesisFirehoseMetadataExtractionProcessor`

Extract partition keys from JSON records with a JQ expression. Used
with Extended S3 dynamic partitioning -- extracted keys are referenced
in the S3 prefix as !{partitionKeyFromQuery:<key>}.

- rule: json_parsing_engine must be 'JQ-1.6' when set

### spec.httpEndpoint.processing.processors[].metadataExtraction.query

`string` · required

JQ expression that extracts partition keys from each JSON record.
The result must be an object whose keys become partition keys.

Example: "{customer_id: .customer_id, event_type: .type}" extracts two
keys, referenced in the prefix as
"data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".

- rule: {"string":{"minLen":"1"}}

### spec.httpEndpoint.processing.processors[].metadataExtraction.jsonParsingEngine

`string`

JSON parsing engine used to evaluate the query.
Valid value: "JQ-1.6" (default -- the only engine AWS supports today).

### spec.httpEndpoint.processing.processors[].decompression

`AwsKinesisFirehoseDecompressionProcessor`

Decompress GZIP-compressed records before delivery. Typically the first
processor when the source sends compressed payloads (e.g., CloudWatch
Logs subscription filters).

- rule: compression_format must be 'GZIP'

### spec.httpEndpoint.processing.processors[].decompression.compressionFormat

`string` · required

Compression format of the incoming records.
Valid value: "GZIP" (the only format AWS supports today).

- rule: {"string":{"minLen":"1"}}

### spec.httpEndpoint.processing.processors[].cloudwatchLogProcessing

`AwsKinesisFirehoseCloudwatchLogProcessingProcessor`

Unwrap CloudWatch Logs subscription envelopes into individual log
events. Use after a decompression processor when the source is a
CloudWatch Logs subscription filter.

### spec.httpEndpoint.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction

`bool`

When true, extract only the log event message field, discarding the
CloudWatch envelope metadata. When false, records pass through with the
envelope intact.

### spec.httpEndpoint.processing.processors[].appendDelimiter

`AwsKinesisFirehoseAppendDelimiterProcessor`

Append a delimiter to every record. Use to produce newline-delimited
JSON (JSON lines) output for query engines and log consumers.

Only supported on the extended_s3 destination -- delimiting is an
S3-object formatting concern; other destinations frame records
natively.

### spec.httpEndpoint.processing.processors[].appendDelimiter.delimiter

`string` · required

Delimiter appended to each record. Use "\\n" for newline-delimited
output -- the format Athena, Spark, and most log consumers expect.

- rule: {"string":{"minLen":"1"}}

### spec.httpEndpoint.processing.processors[].recordDeaggregation

`AwsKinesisFirehoseRecordDeaggregationProcessor`

Split multi-record aggregates (e.g., KPL-aggregated or delimited
payloads) into individual records before further processing. Required
before dynamic partitioning when producers aggregate records.

Only supported on the extended_s3 destination -- AWS rejects it for
every other destination type at creation.

- rule: sub_record_type must be 'JSON' or 'DELIMITED'
- rule: delimiter is required when sub_record_type is 'DELIMITED'

### spec.httpEndpoint.processing.processors[].recordDeaggregation.subRecordType

`string` · required

How records are aggregated in the payload:
- "JSON" -- concatenated JSON documents (no delimiter needed).
- "DELIMITED" -- records separated by a custom delimiter; requires
  the delimiter field.

- rule: {"string":{"minLen":"1"}}

### spec.httpEndpoint.processing.processors[].recordDeaggregation.delimiter

`string`

Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
newline). Required when sub_record_type is "DELIMITED".

### spec.httpEndpoint.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch error logging for HTTP delivery failures.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.httpEndpoint.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.httpEndpoint.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.httpEndpoint.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.httpEndpoint.requestConfig

`AwsKinesisFirehoseRequestConfig`

Request configuration for customizing the HTTP request format.

- rule: content_encoding must be 'NONE' or 'GZIP' when set

### spec.httpEndpoint.requestConfig.contentEncoding

`string`

Content encoding for the HTTP request body.
Valid values: "NONE" (default), "GZIP".
GZIP reduces payload size but adds CPU overhead.

### spec.httpEndpoint.requestConfig.commonAttributes

`[]AwsKinesisFirehoseRequestAttribute`

Custom key-value pairs sent as HTTP headers with every request.
Use this for endpoint-specific metadata (e.g., dataset name,
environment identifier, API version).

### spec.httpEndpoint.requestConfig.commonAttributes[].name

`string` · required

Header name.

- rule: {"string":{"minLen":"1"}}

### spec.httpEndpoint.requestConfig.commonAttributes[].value

`string` · required

Header value.

- rule: {"string":{"minLen":"1"}}

### spec.redshift

`AwsKinesisFirehoseRedshiftDestination`

Redshift destination for data warehouse loading. Firehose stages data
in S3, then issues a Redshift COPY command to load it. Supports record
transformation and optional S3 backup of source records.

- rule: configure exactly one authentication mode: username+password, or secrets_manager
- rule: username and password must be set together
- rule: retry_duration_in_seconds must be between 0 and 7200 when set
- rule: s3_backup_mode must be 'Disabled' or 'Enabled' when set
- rule: s3_backup requires s3_backup_mode to be 'Enabled'
- rule: record_deaggregation and append_delimiter processors are only supported on the extended_s3 destination

### spec.redshift.clusterJdbcurl

`string` · required

JDBC URL of the Redshift cluster. Format:
  "jdbc:redshift://<endpoint>:<port>/<database>"
Example: "jdbc:redshift://my-cluster.abcdef.us-east-1.redshift.amazonaws.com:5439/mydb"

- rule: {"string":{"minLen":"1"}}

### spec.redshift.roleArn

`string | valueFrom` · required

IAM role ARN granting Firehose permission to COPY from S3 to Redshift
and write to the S3 staging bucket. Must have:
- S3 read access to the staging bucket
- Redshift COPY permission

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.redshift.dataTableName

`string` · required

Name of the target Redshift table for the COPY command.

- rule: {"string":{"minLen":"1"}}

### spec.redshift.dataTableColumns

`string`

Comma-separated list of column names for the COPY command. When set,
only the specified columns are loaded. When absent, COPY loads into
all columns in table order.

### spec.redshift.copyOptions

`string`

Additional COPY command options (e.g., "JSON 'auto'", "GZIP",
"DELIMITER ','", "IGNOREHEADER 1"). Appended to the COPY command.

### spec.redshift.username

`string`

Redshift database username. Required together with password when
authenticating with plaintext credentials; must be empty when
secrets_manager is set.

### spec.redshift.password

`string` · sensitive

Redshift database password. Sensitive -- the value lands in IaC state.
Prefer secrets_manager for production, which keeps the credential in
Secrets Manager entirely. Required together with username when
authenticating with plaintext credentials.

### spec.redshift.secretsManager

`AwsKinesisFirehoseSecretsManagerConfig`

Source the credentials from AWS Secrets Manager instead of
username/password. The secret shape is
{"username": "...", "password": "..."}. Setting this block enables
Secrets Manager authentication (ForceNew).

### spec.redshift.secretsManager.secretArn

`string | valueFrom` · required

ARN of the Secrets Manager secret holding the destination credential.
An ARN reference resolved by Firehose at delivery time -- never the
secret material itself.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.redshift.secretsManager.roleArn

`string | valueFrom`

IAM role ARN granting Firehose permission to read the secret
(secretsmanager:GetSecretValue). When absent, the destination's
delivery role is used.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.redshift.s3Config

`AwsKinesisFirehoseS3Config` · required

S3 configuration for the intermediate staging bucket. Firehose writes
data to this S3 location, then issues a COPY command to load it into
Redshift. This is NOT a backup -- it's the primary data path.

- rule: {"required":true}
- rule: compression_format must be 'UNCOMPRESSED', 'GZIP', 'ZIP', 'Snappy', or 'HADOOP_SNAPPY' when set

### spec.redshift.s3Config.bucketArn

`string | valueFrom` · required

S3 bucket ARN where records are delivered.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.redshift.s3Config.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to write to the S3 bucket.
The role must have s3:PutObject, s3:AbortMultipartUpload,
s3:GetBucketLocation, and s3:ListBucket permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.redshift.s3Config.prefix

`string`

S3 key prefix prepended to delivered objects. Supports Firehose expression
syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").

### spec.redshift.s3Config.errorOutputPrefix

`string`

S3 key prefix for error output. When Firehose cannot deliver or transform
a record, it writes to this prefix. Uses the same expression syntax as prefix.

### spec.redshift.s3Config.compressionFormat

`string`

Compression format for delivered objects. Applied before writing to S3.
Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
Default: "UNCOMPRESSED".

### spec.redshift.s3Config.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
When absent, S3 uses its default encryption settings (SSE-S3 or bucket
default encryption).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.redshift.s3Config.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for S3 delivery.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.redshift.s3Config.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.redshift.s3Config.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.redshift.s3Config.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch logging configuration for S3 delivery errors.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.redshift.s3Config.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.redshift.s3Config.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.redshift.s3Config.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.redshift.retryDurationInSeconds

`int32`

Retry duration in seconds for failed Redshift COPY commands.
Range: 0-7200. Default: 3600 seconds (1 hour).
Redshift COPY can be slow, so a longer default is appropriate.

### spec.redshift.s3BackupMode

`string`

S3 backup mode for source records (in addition to the staging S3).
When "Enabled", a copy of the original records is written to
s3_backup. Useful for auditing and reprocessing.

Valid values: "Disabled" (default), "Enabled".

### spec.redshift.s3Backup

`AwsKinesisFirehoseS3Config`

S3 configuration for source record backup. Required when
s3_backup_mode is "Enabled".

- rule: compression_format must be 'UNCOMPRESSED', 'GZIP', 'ZIP', 'Snappy', or 'HADOOP_SNAPPY' when set

### spec.redshift.s3Backup.bucketArn

`string | valueFrom` · required

S3 bucket ARN where records are delivered.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.redshift.s3Backup.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to write to the S3 bucket.
The role must have s3:PutObject, s3:AbortMultipartUpload,
s3:GetBucketLocation, and s3:ListBucket permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.redshift.s3Backup.prefix

`string`

S3 key prefix prepended to delivered objects. Supports Firehose expression
syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").

### spec.redshift.s3Backup.errorOutputPrefix

`string`

S3 key prefix for error output. When Firehose cannot deliver or transform
a record, it writes to this prefix. Uses the same expression syntax as prefix.

### spec.redshift.s3Backup.compressionFormat

`string`

Compression format for delivered objects. Applied before writing to S3.
Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
Default: "UNCOMPRESSED".

### spec.redshift.s3Backup.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
When absent, S3 uses its default encryption settings (SSE-S3 or bucket
default encryption).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.redshift.s3Backup.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for S3 delivery.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.redshift.s3Backup.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.redshift.s3Backup.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.redshift.s3Backup.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch logging configuration for S3 delivery errors.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.redshift.s3Backup.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.redshift.s3Backup.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.redshift.s3Backup.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.redshift.processing

`AwsKinesisFirehoseProcessing`

Record-transformation pipeline applied before staging to S3.

- rule: processors require enabled to be true
- rule: at least one processor is required when processing is enabled

### spec.redshift.processing.enabled

`bool`

Enable the processing pipeline. When true, at least one processor should
be configured.

### spec.redshift.processing.processors

`[]AwsKinesisFirehoseProcessor`

Ordered list of processors. Each entry configures exactly one processor
type; Firehose executes them in order.

- rule: exactly one processor must be set: lambda, metadata_extraction, decompression, cloudwatch_log_processing, append_delimiter, or record_deaggregation

### spec.redshift.processing.processors[].lambda

`AwsKinesisFirehoseLambdaProcessor`

Invoke an AWS Lambda function to transform records. The function
receives batches of records and returns transformed records with a
status (Ok, Dropped, ProcessingFailed) per record.

- rule: buffer_size_in_mbs must be between 0.2 and 3 when set
- rule: buffer_interval_in_seconds must be between 60 and 900 when set
- rule: number_of_retries must be between 0 and 300 when set

### spec.redshift.processing.processors[].lambda.lambdaArn

`string | valueFrom` · required

ARN of the Lambda function that transforms records. May include a
version or alias qualifier to pin the deployed transformation.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.redshift.processing.processors[].lambda.bufferSizeInMbs

`double`

Buffer size in MiB that Firehose accumulates before invoking Lambda.
Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
destination is Splunk).

Smaller buffers invoke Lambda more frequently with smaller batches.
Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
invocation costs.

### spec.redshift.processing.processors[].lambda.bufferIntervalInSeconds

`int32`

Buffer interval in seconds. Firehose invokes Lambda when this interval
elapses, even if the buffer size threshold has not been reached.
Range: 60-900 seconds. Default: 60 seconds.

### spec.redshift.processing.processors[].lambda.numberOfRetries

`int32`

Number of times Firehose retries a failed Lambda invocation before
writing the record to the error output prefix.
Range: 0-300. Default: 3.

### spec.redshift.processing.processors[].lambda.roleArn

`string | valueFrom`

IAM role ARN Firehose assumes to invoke the Lambda function. When
absent, Firehose uses the delivery stream's destination role -- the
right choice for almost every pipeline. Set this only when the
transformation function must be invoked with a DIFFERENT role than the
one that writes to the destination (e.g. the function lives in another
account). Note: AWS reports the delivery role back for unset values, and
the provider does not store default-valued processor parameters in
state -- so set this only to a non-default role, never to the delivery
role itself (that would cause perpetual plan diffs).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.redshift.processing.processors[].metadataExtraction

`AwsKinesisFirehoseMetadataExtractionProcessor`

Extract partition keys from JSON records with a JQ expression. Used
with Extended S3 dynamic partitioning -- extracted keys are referenced
in the S3 prefix as !{partitionKeyFromQuery:<key>}.

- rule: json_parsing_engine must be 'JQ-1.6' when set

### spec.redshift.processing.processors[].metadataExtraction.query

`string` · required

JQ expression that extracts partition keys from each JSON record.
The result must be an object whose keys become partition keys.

Example: "{customer_id: .customer_id, event_type: .type}" extracts two
keys, referenced in the prefix as
"data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".

- rule: {"string":{"minLen":"1"}}

### spec.redshift.processing.processors[].metadataExtraction.jsonParsingEngine

`string`

JSON parsing engine used to evaluate the query.
Valid value: "JQ-1.6" (default -- the only engine AWS supports today).

### spec.redshift.processing.processors[].decompression

`AwsKinesisFirehoseDecompressionProcessor`

Decompress GZIP-compressed records before delivery. Typically the first
processor when the source sends compressed payloads (e.g., CloudWatch
Logs subscription filters).

- rule: compression_format must be 'GZIP'

### spec.redshift.processing.processors[].decompression.compressionFormat

`string` · required

Compression format of the incoming records.
Valid value: "GZIP" (the only format AWS supports today).

- rule: {"string":{"minLen":"1"}}

### spec.redshift.processing.processors[].cloudwatchLogProcessing

`AwsKinesisFirehoseCloudwatchLogProcessingProcessor`

Unwrap CloudWatch Logs subscription envelopes into individual log
events. Use after a decompression processor when the source is a
CloudWatch Logs subscription filter.

### spec.redshift.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction

`bool`

When true, extract only the log event message field, discarding the
CloudWatch envelope metadata. When false, records pass through with the
envelope intact.

### spec.redshift.processing.processors[].appendDelimiter

`AwsKinesisFirehoseAppendDelimiterProcessor`

Append a delimiter to every record. Use to produce newline-delimited
JSON (JSON lines) output for query engines and log consumers.

Only supported on the extended_s3 destination -- delimiting is an
S3-object formatting concern; other destinations frame records
natively.

### spec.redshift.processing.processors[].appendDelimiter.delimiter

`string` · required

Delimiter appended to each record. Use "\\n" for newline-delimited
output -- the format Athena, Spark, and most log consumers expect.

- rule: {"string":{"minLen":"1"}}

### spec.redshift.processing.processors[].recordDeaggregation

`AwsKinesisFirehoseRecordDeaggregationProcessor`

Split multi-record aggregates (e.g., KPL-aggregated or delimited
payloads) into individual records before further processing. Required
before dynamic partitioning when producers aggregate records.

Only supported on the extended_s3 destination -- AWS rejects it for
every other destination type at creation.

- rule: sub_record_type must be 'JSON' or 'DELIMITED'
- rule: delimiter is required when sub_record_type is 'DELIMITED'

### spec.redshift.processing.processors[].recordDeaggregation.subRecordType

`string` · required

How records are aggregated in the payload:
- "JSON" -- concatenated JSON documents (no delimiter needed).
- "DELIMITED" -- records separated by a custom delimiter; requires
  the delimiter field.

- rule: {"string":{"minLen":"1"}}

### spec.redshift.processing.processors[].recordDeaggregation.delimiter

`string`

Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
newline). Required when sub_record_type is "DELIMITED".

### spec.redshift.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch error logging for Redshift COPY failures.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.redshift.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.redshift.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.redshift.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.splunk

`AwsKinesisFirehoseSplunkDestination`

Splunk destination for delivery to a Splunk HTTP Event Collector (HEC)
endpoint -- Splunk Cloud, Splunk Enterprise, or Splunk-managed AWS.
Failed events are backed up to S3.

- rule: hec_endpoint must start with 'https://'
- rule: hec_endpoint_type must be 'Raw' or 'Event' when set
- rule: configure exactly one authentication mode: hec_token, or secrets_manager
- rule: hec_acknowledgment_timeout_in_seconds must be between 180 and 600 when set
- rule: retry_duration_in_seconds must be between 0 and 7200 when set
- rule: s3_backup_mode must be 'FailedEventsOnly' or 'AllEvents' when set
- rule: buffering for Splunk destinations must not exceed 60 seconds / 5 MiB
- rule: record_deaggregation and append_delimiter processors are only supported on the extended_s3 destination

### spec.splunk.hecEndpoint

`string` · required

Splunk HTTP Event Collector endpoint URL, including the port.
Examples:
- "https://http-inputs-mycompany.splunkcloud.com:443"
- "https://splunk.example.com:8088"

- rule: {"string":{"minLen":"1"}}

### spec.splunk.hecEndpointType

`string`

HEC endpoint type:
- "Raw" (default) -- events are sent to the raw endpoint as-is. Use for
  preformatted events (the common case for Firehose delivery).
- "Event" -- events are sent to the event endpoint and must be JSON
  objects in Splunk's event format.

### spec.splunk.hecToken

`string` · sensitive

HEC token that authorizes delivery, minted in Splunk when the HEC input
is created. Sensitive -- the value lands in IaC state. Prefer
secrets_manager for production, which keeps the token in Secrets
Manager entirely.

### spec.splunk.secretsManager

`AwsKinesisFirehoseSecretsManagerConfig`

Source the HEC token from AWS Secrets Manager instead of hec_token.
The secret shape is {"hec_token": "..."}. Setting this block enables
Secrets Manager authentication (ForceNew).

### spec.splunk.secretsManager.secretArn

`string | valueFrom` · required

ARN of the Secrets Manager secret holding the destination credential.
An ARN reference resolved by Firehose at delivery time -- never the
secret material itself.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.splunk.secretsManager.roleArn

`string | valueFrom`

IAM role ARN granting Firehose permission to read the secret
(secretsmanager:GetSecretValue). When absent, the destination's
delivery role is used.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.splunk.hecAcknowledgmentTimeoutInSeconds

`int32`

Time in seconds that Firehose waits for the Splunk indexer
acknowledgment after sending data. Unacknowledged data is retried or
backed up to S3.
Range: 180-600. Default: 180 seconds.

### spec.splunk.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for Splunk delivery. Default: 60s interval, 5 MiB.
Splunk enforces the tightest limits of any destination:
interval 0-60 seconds, size 1-5 MiB.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.splunk.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.splunk.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.splunk.retryDurationInSeconds

`int32`

Retry duration in seconds for failed or unacknowledged HEC deliveries.
Range: 0-7200. Default: 3600 seconds (1 hour).

### spec.splunk.s3BackupMode

`string`

S3 backup mode for events. Controls when records are written to S3.

Valid values:
- "FailedEventsOnly" (default) -- only events that fail HEC delivery
  are backed up to S3.
- "AllEvents" -- all events are backed up to S3 in addition to being
  sent to Splunk.

### spec.splunk.s3Config

`AwsKinesisFirehoseS3Config` · required

S3 configuration for backing up failed (or all) events. Required.

- rule: {"required":true}
- rule: compression_format must be 'UNCOMPRESSED', 'GZIP', 'ZIP', 'Snappy', or 'HADOOP_SNAPPY' when set

### spec.splunk.s3Config.bucketArn

`string | valueFrom` · required

S3 bucket ARN where records are delivered.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.splunk.s3Config.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to write to the S3 bucket.
The role must have s3:PutObject, s3:AbortMultipartUpload,
s3:GetBucketLocation, and s3:ListBucket permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.splunk.s3Config.prefix

`string`

S3 key prefix prepended to delivered objects. Supports Firehose expression
syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").

### spec.splunk.s3Config.errorOutputPrefix

`string`

S3 key prefix for error output. When Firehose cannot deliver or transform
a record, it writes to this prefix. Uses the same expression syntax as prefix.

### spec.splunk.s3Config.compressionFormat

`string`

Compression format for delivered objects. Applied before writing to S3.
Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
Default: "UNCOMPRESSED".

### spec.splunk.s3Config.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
When absent, S3 uses its default encryption settings (SSE-S3 or bucket
default encryption).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.splunk.s3Config.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for S3 delivery.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.splunk.s3Config.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.splunk.s3Config.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.splunk.s3Config.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch logging configuration for S3 delivery errors.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.splunk.s3Config.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.splunk.s3Config.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.splunk.s3Config.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.splunk.processing

`AwsKinesisFirehoseProcessing`

Record-transformation pipeline applied before HEC delivery.

- rule: processors require enabled to be true
- rule: at least one processor is required when processing is enabled

### spec.splunk.processing.enabled

`bool`

Enable the processing pipeline. When true, at least one processor should
be configured.

### spec.splunk.processing.processors

`[]AwsKinesisFirehoseProcessor`

Ordered list of processors. Each entry configures exactly one processor
type; Firehose executes them in order.

- rule: exactly one processor must be set: lambda, metadata_extraction, decompression, cloudwatch_log_processing, append_delimiter, or record_deaggregation

### spec.splunk.processing.processors[].lambda

`AwsKinesisFirehoseLambdaProcessor`

Invoke an AWS Lambda function to transform records. The function
receives batches of records and returns transformed records with a
status (Ok, Dropped, ProcessingFailed) per record.

- rule: buffer_size_in_mbs must be between 0.2 and 3 when set
- rule: buffer_interval_in_seconds must be between 60 and 900 when set
- rule: number_of_retries must be between 0 and 300 when set

### spec.splunk.processing.processors[].lambda.lambdaArn

`string | valueFrom` · required

ARN of the Lambda function that transforms records. May include a
version or alias qualifier to pin the deployed transformation.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.splunk.processing.processors[].lambda.bufferSizeInMbs

`double`

Buffer size in MiB that Firehose accumulates before invoking Lambda.
Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
destination is Splunk).

Smaller buffers invoke Lambda more frequently with smaller batches.
Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
invocation costs.

### spec.splunk.processing.processors[].lambda.bufferIntervalInSeconds

`int32`

Buffer interval in seconds. Firehose invokes Lambda when this interval
elapses, even if the buffer size threshold has not been reached.
Range: 60-900 seconds. Default: 60 seconds.

### spec.splunk.processing.processors[].lambda.numberOfRetries

`int32`

Number of times Firehose retries a failed Lambda invocation before
writing the record to the error output prefix.
Range: 0-300. Default: 3.

### spec.splunk.processing.processors[].lambda.roleArn

`string | valueFrom`

IAM role ARN Firehose assumes to invoke the Lambda function. When
absent, Firehose uses the delivery stream's destination role -- the
right choice for almost every pipeline. Set this only when the
transformation function must be invoked with a DIFFERENT role than the
one that writes to the destination (e.g. the function lives in another
account). Note: AWS reports the delivery role back for unset values, and
the provider does not store default-valued processor parameters in
state -- so set this only to a non-default role, never to the delivery
role itself (that would cause perpetual plan diffs).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.splunk.processing.processors[].metadataExtraction

`AwsKinesisFirehoseMetadataExtractionProcessor`

Extract partition keys from JSON records with a JQ expression. Used
with Extended S3 dynamic partitioning -- extracted keys are referenced
in the S3 prefix as !{partitionKeyFromQuery:<key>}.

- rule: json_parsing_engine must be 'JQ-1.6' when set

### spec.splunk.processing.processors[].metadataExtraction.query

`string` · required

JQ expression that extracts partition keys from each JSON record.
The result must be an object whose keys become partition keys.

Example: "{customer_id: .customer_id, event_type: .type}" extracts two
keys, referenced in the prefix as
"data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".

- rule: {"string":{"minLen":"1"}}

### spec.splunk.processing.processors[].metadataExtraction.jsonParsingEngine

`string`

JSON parsing engine used to evaluate the query.
Valid value: "JQ-1.6" (default -- the only engine AWS supports today).

### spec.splunk.processing.processors[].decompression

`AwsKinesisFirehoseDecompressionProcessor`

Decompress GZIP-compressed records before delivery. Typically the first
processor when the source sends compressed payloads (e.g., CloudWatch
Logs subscription filters).

- rule: compression_format must be 'GZIP'

### spec.splunk.processing.processors[].decompression.compressionFormat

`string` · required

Compression format of the incoming records.
Valid value: "GZIP" (the only format AWS supports today).

- rule: {"string":{"minLen":"1"}}

### spec.splunk.processing.processors[].cloudwatchLogProcessing

`AwsKinesisFirehoseCloudwatchLogProcessingProcessor`

Unwrap CloudWatch Logs subscription envelopes into individual log
events. Use after a decompression processor when the source is a
CloudWatch Logs subscription filter.

### spec.splunk.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction

`bool`

When true, extract only the log event message field, discarding the
CloudWatch envelope metadata. When false, records pass through with the
envelope intact.

### spec.splunk.processing.processors[].appendDelimiter

`AwsKinesisFirehoseAppendDelimiterProcessor`

Append a delimiter to every record. Use to produce newline-delimited
JSON (JSON lines) output for query engines and log consumers.

Only supported on the extended_s3 destination -- delimiting is an
S3-object formatting concern; other destinations frame records
natively.

### spec.splunk.processing.processors[].appendDelimiter.delimiter

`string` · required

Delimiter appended to each record. Use "\\n" for newline-delimited
output -- the format Athena, Spark, and most log consumers expect.

- rule: {"string":{"minLen":"1"}}

### spec.splunk.processing.processors[].recordDeaggregation

`AwsKinesisFirehoseRecordDeaggregationProcessor`

Split multi-record aggregates (e.g., KPL-aggregated or delimited
payloads) into individual records before further processing. Required
before dynamic partitioning when producers aggregate records.

Only supported on the extended_s3 destination -- AWS rejects it for
every other destination type at creation.

- rule: sub_record_type must be 'JSON' or 'DELIMITED'
- rule: delimiter is required when sub_record_type is 'DELIMITED'

### spec.splunk.processing.processors[].recordDeaggregation.subRecordType

`string` · required

How records are aggregated in the payload:
- "JSON" -- concatenated JSON documents (no delimiter needed).
- "DELIMITED" -- records separated by a custom delimiter; requires
  the delimiter field.

- rule: {"string":{"minLen":"1"}}

### spec.splunk.processing.processors[].recordDeaggregation.delimiter

`string`

Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
newline). Required when sub_record_type is "DELIMITED".

### spec.splunk.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch error logging for Splunk delivery failures.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.splunk.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.splunk.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.splunk.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.snowflake

`AwsKinesisFirehoseSnowflakeDestination`

Snowflake destination for direct streaming into a Snowflake table via
Snowpipe Streaming. Authenticates with key-pair credentials (or AWS
Secrets Manager) and supports PrivateLink. Failed data is backed up
to S3.

- rule: account_url must start with 'https://'
- rule: configure exactly one authentication mode: user+private_key, or secrets_manager
- rule: user and private_key must be set together
- rule: key_passphrase requires private_key
- rule: data_loading_option must be 'JSON_MAPPING', 'VARIANT_CONTENT_MAPPING', or 'VARIANT_CONTENT_AND_METADATA_MAPPING' when set
- rule: content_column_name is required for VARIANT_CONTENT_MAPPING and VARIANT_CONTENT_AND_METADATA_MAPPING
- rule: metadata_column_name is required for VARIANT_CONTENT_AND_METADATA_MAPPING
- rule: retry_duration_in_seconds must be between 0 and 7200 when set
- rule: s3_backup_mode must be 'FailedDataOnly' or 'AllData' when set
- rule: record_deaggregation and append_delimiter processors are only supported on the extended_s3 destination

### spec.snowflake.accountUrl

`string` · required

Snowflake account URL.
Format: "https://<account-identifier>.snowflakecomputing.com"

- rule: {"string":{"minLen":"1"}}

### spec.snowflake.database

`string` · required

Name of the Snowflake database containing the target table.

- rule: {"string":{"minLen":"1"}}

### spec.snowflake.schema

`string` · required

Name of the Snowflake schema containing the target table.

- rule: {"string":{"minLen":"1"}}

### spec.snowflake.table

`string` · required

Name of the target Snowflake table.

- rule: {"string":{"minLen":"1"}}

### spec.snowflake.roleArn

`string | valueFrom` · required

IAM role ARN granting Firehose permission to write to the S3 backup
bucket and read the Secrets Manager secret (when used).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.snowflake.user

`string`

Snowflake user that owns the key pair. Required together with
private_key when authenticating with inline credentials; must be empty
when secrets_manager is set.

### spec.snowflake.privateKey

`string` · sensitive

RSA private key for key-pair authentication: the PEM body only, without
the "-----BEGIN/END PRIVATE KEY-----" header and footer lines.
Sensitive -- the value lands in IaC state. Prefer secrets_manager for
production, which keeps the key in Secrets Manager entirely.

### spec.snowflake.keyPassphrase

`string` · sensitive

Passphrase for an encrypted private key. Only set when the private key
is encrypted. Length: 7-255 characters. Sensitive.

### spec.snowflake.secretsManager

`AwsKinesisFirehoseSecretsManagerConfig`

Source the credentials from AWS Secrets Manager instead of
user/private_key. The secret shape is
{"user": "...", "private_key": "...", "key_passphrase": "..."}.
Setting this block enables Secrets Manager authentication (ForceNew).

### spec.snowflake.secretsManager.secretArn

`string | valueFrom` · required

ARN of the Secrets Manager secret holding the destination credential.
An ARN reference resolved by Firehose at delivery time -- never the
secret material itself.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.snowflake.secretsManager.roleArn

`string | valueFrom`

IAM role ARN granting Firehose permission to read the secret
(secretsmanager:GetSecretValue). When absent, the destination's
delivery role is used.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.snowflake.dataLoadingOption

`string`

How records map onto the target table:
- "JSON_MAPPING" (default) -- each top-level JSON key maps to a
  same-named table column.
- "VARIANT_CONTENT_MAPPING" -- the whole record lands in one VARIANT
  column (content_column_name).
- "VARIANT_CONTENT_AND_METADATA_MAPPING" -- the record lands in a
  VARIANT column and Firehose metadata lands in a second VARIANT column
  (metadata_column_name).

### spec.snowflake.contentColumnName

`string`

Name of the VARIANT column that receives record content. Required for
the VARIANT_CONTENT_MAPPING and VARIANT_CONTENT_AND_METADATA_MAPPING
loading options.

### spec.snowflake.metadataColumnName

`string`

Name of the VARIANT column that receives Firehose metadata. Required
for the VARIANT_CONTENT_AND_METADATA_MAPPING loading option.

### spec.snowflake.snowflakeRole

`string`

Snowflake role to assume for the insert. When absent, the user's
default role is used. Setting a dedicated ingestion role with
insert-only privileges is the recommended least-privilege posture.

### spec.snowflake.privateLinkVpceId

`string`

AWS PrivateLink VPCE ID for private connectivity to Snowflake
(privatelink account URLs). When absent, Firehose connects over the
public internet. Format: "com.amazonaws.vpce.<region>.vpce-svc-<id>".

### spec.snowflake.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for Snowflake delivery. Default: 0s interval, 1 MiB --
near-real-time ingestion via Snowpipe Streaming. Raise the interval to
trade latency for fewer, larger inserts.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.snowflake.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.snowflake.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.snowflake.retryDurationInSeconds

`int32`

Retry duration in seconds for failed Snowflake inserts.
Range: 0-7200. Default: 60 seconds.

### spec.snowflake.s3BackupMode

`string`

S3 backup mode. Controls when records are written to S3.

Valid values:
- "FailedDataOnly" (default) -- only records that fail Snowflake
  delivery are backed up to S3.
- "AllData" -- all records are backed up to S3 in addition to being
  delivered to Snowflake.

### spec.snowflake.s3Config

`AwsKinesisFirehoseS3Config` · required

S3 configuration for backing up failed (or all) records. Required.

- rule: {"required":true}
- rule: compression_format must be 'UNCOMPRESSED', 'GZIP', 'ZIP', 'Snappy', or 'HADOOP_SNAPPY' when set

### spec.snowflake.s3Config.bucketArn

`string | valueFrom` · required

S3 bucket ARN where records are delivered.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.snowflake.s3Config.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to write to the S3 bucket.
The role must have s3:PutObject, s3:AbortMultipartUpload,
s3:GetBucketLocation, and s3:ListBucket permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.snowflake.s3Config.prefix

`string`

S3 key prefix prepended to delivered objects. Supports Firehose expression
syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").

### spec.snowflake.s3Config.errorOutputPrefix

`string`

S3 key prefix for error output. When Firehose cannot deliver or transform
a record, it writes to this prefix. Uses the same expression syntax as prefix.

### spec.snowflake.s3Config.compressionFormat

`string`

Compression format for delivered objects. Applied before writing to S3.
Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
Default: "UNCOMPRESSED".

### spec.snowflake.s3Config.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
When absent, S3 uses its default encryption settings (SSE-S3 or bucket
default encryption).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.snowflake.s3Config.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for S3 delivery.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.snowflake.s3Config.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.snowflake.s3Config.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.snowflake.s3Config.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch logging configuration for S3 delivery errors.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.snowflake.s3Config.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.snowflake.s3Config.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.snowflake.s3Config.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.snowflake.processing

`AwsKinesisFirehoseProcessing`

Record-transformation pipeline applied before delivery.

- rule: processors require enabled to be true
- rule: at least one processor is required when processing is enabled

### spec.snowflake.processing.enabled

`bool`

Enable the processing pipeline. When true, at least one processor should
be configured.

### spec.snowflake.processing.processors

`[]AwsKinesisFirehoseProcessor`

Ordered list of processors. Each entry configures exactly one processor
type; Firehose executes them in order.

- rule: exactly one processor must be set: lambda, metadata_extraction, decompression, cloudwatch_log_processing, append_delimiter, or record_deaggregation

### spec.snowflake.processing.processors[].lambda

`AwsKinesisFirehoseLambdaProcessor`

Invoke an AWS Lambda function to transform records. The function
receives batches of records and returns transformed records with a
status (Ok, Dropped, ProcessingFailed) per record.

- rule: buffer_size_in_mbs must be between 0.2 and 3 when set
- rule: buffer_interval_in_seconds must be between 60 and 900 when set
- rule: number_of_retries must be between 0 and 300 when set

### spec.snowflake.processing.processors[].lambda.lambdaArn

`string | valueFrom` · required

ARN of the Lambda function that transforms records. May include a
version or alias qualifier to pin the deployed transformation.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.snowflake.processing.processors[].lambda.bufferSizeInMbs

`double`

Buffer size in MiB that Firehose accumulates before invoking Lambda.
Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
destination is Splunk).

Smaller buffers invoke Lambda more frequently with smaller batches.
Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
invocation costs.

### spec.snowflake.processing.processors[].lambda.bufferIntervalInSeconds

`int32`

Buffer interval in seconds. Firehose invokes Lambda when this interval
elapses, even if the buffer size threshold has not been reached.
Range: 60-900 seconds. Default: 60 seconds.

### spec.snowflake.processing.processors[].lambda.numberOfRetries

`int32`

Number of times Firehose retries a failed Lambda invocation before
writing the record to the error output prefix.
Range: 0-300. Default: 3.

### spec.snowflake.processing.processors[].lambda.roleArn

`string | valueFrom`

IAM role ARN Firehose assumes to invoke the Lambda function. When
absent, Firehose uses the delivery stream's destination role -- the
right choice for almost every pipeline. Set this only when the
transformation function must be invoked with a DIFFERENT role than the
one that writes to the destination (e.g. the function lives in another
account). Note: AWS reports the delivery role back for unset values, and
the provider does not store default-valued processor parameters in
state -- so set this only to a non-default role, never to the delivery
role itself (that would cause perpetual plan diffs).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.snowflake.processing.processors[].metadataExtraction

`AwsKinesisFirehoseMetadataExtractionProcessor`

Extract partition keys from JSON records with a JQ expression. Used
with Extended S3 dynamic partitioning -- extracted keys are referenced
in the S3 prefix as !{partitionKeyFromQuery:<key>}.

- rule: json_parsing_engine must be 'JQ-1.6' when set

### spec.snowflake.processing.processors[].metadataExtraction.query

`string` · required

JQ expression that extracts partition keys from each JSON record.
The result must be an object whose keys become partition keys.

Example: "{customer_id: .customer_id, event_type: .type}" extracts two
keys, referenced in the prefix as
"data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".

- rule: {"string":{"minLen":"1"}}

### spec.snowflake.processing.processors[].metadataExtraction.jsonParsingEngine

`string`

JSON parsing engine used to evaluate the query.
Valid value: "JQ-1.6" (default -- the only engine AWS supports today).

### spec.snowflake.processing.processors[].decompression

`AwsKinesisFirehoseDecompressionProcessor`

Decompress GZIP-compressed records before delivery. Typically the first
processor when the source sends compressed payloads (e.g., CloudWatch
Logs subscription filters).

- rule: compression_format must be 'GZIP'

### spec.snowflake.processing.processors[].decompression.compressionFormat

`string` · required

Compression format of the incoming records.
Valid value: "GZIP" (the only format AWS supports today).

- rule: {"string":{"minLen":"1"}}

### spec.snowflake.processing.processors[].cloudwatchLogProcessing

`AwsKinesisFirehoseCloudwatchLogProcessingProcessor`

Unwrap CloudWatch Logs subscription envelopes into individual log
events. Use after a decompression processor when the source is a
CloudWatch Logs subscription filter.

### spec.snowflake.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction

`bool`

When true, extract only the log event message field, discarding the
CloudWatch envelope metadata. When false, records pass through with the
envelope intact.

### spec.snowflake.processing.processors[].appendDelimiter

`AwsKinesisFirehoseAppendDelimiterProcessor`

Append a delimiter to every record. Use to produce newline-delimited
JSON (JSON lines) output for query engines and log consumers.

Only supported on the extended_s3 destination -- delimiting is an
S3-object formatting concern; other destinations frame records
natively.

### spec.snowflake.processing.processors[].appendDelimiter.delimiter

`string` · required

Delimiter appended to each record. Use "\\n" for newline-delimited
output -- the format Athena, Spark, and most log consumers expect.

- rule: {"string":{"minLen":"1"}}

### spec.snowflake.processing.processors[].recordDeaggregation

`AwsKinesisFirehoseRecordDeaggregationProcessor`

Split multi-record aggregates (e.g., KPL-aggregated or delimited
payloads) into individual records before further processing. Required
before dynamic partitioning when producers aggregate records.

Only supported on the extended_s3 destination -- AWS rejects it for
every other destination type at creation.

- rule: sub_record_type must be 'JSON' or 'DELIMITED'
- rule: delimiter is required when sub_record_type is 'DELIMITED'

### spec.snowflake.processing.processors[].recordDeaggregation.subRecordType

`string` · required

How records are aggregated in the payload:
- "JSON" -- concatenated JSON documents (no delimiter needed).
- "DELIMITED" -- records separated by a custom delimiter; requires
  the delimiter field.

- rule: {"string":{"minLen":"1"}}

### spec.snowflake.processing.processors[].recordDeaggregation.delimiter

`string`

Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
newline). Required when sub_record_type is "DELIMITED".

### spec.snowflake.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch error logging for Snowflake delivery failures.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.snowflake.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.snowflake.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.snowflake.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.iceberg

`AwsKinesisFirehoseIcebergDestination`

Iceberg destination for delivery into Apache Iceberg tables managed by
the AWS Glue Data Catalog. Supports routing records to multiple tables
and update/delete semantics via unique keys. Failed data is backed up
to S3.

- rule: retry_duration_in_seconds must be between 0 and 7200 when set
- rule: s3_backup_mode must be 'FailedDataOnly' or 'AllData' when set
- rule: record_deaggregation and append_delimiter processors are only supported on the extended_s3 destination

### spec.iceberg.catalogArn

`string | valueFrom` · required

ARN of the Glue Data Catalog that owns the Iceberg tables. The catalog
ARN embeds the owning account:
  "arn:aws:glue:<region>:<account-id>:catalog"
ForceNew -- changing the catalog replaces the delivery stream.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.iceberg.roleArn

`string | valueFrom` · required

IAM role ARN granting Firehose permission to write to the Iceberg
tables: Glue table read/update and S3 read/write on the warehouse
location, plus the S3 backup bucket.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.iceberg.destinationTables

`[]AwsKinesisFirehoseIcebergDestinationTable` · required

Destination tables for delivered records -- at least one is required
(an Iceberg destination with no table cannot route records; AWS
rejects it at apply time, so it is rejected here at validation time).
When exactly one table is listed, all records land there. When
multiple tables are listed, records must carry routing metadata
(produced by a metadata_extraction or Lambda processor) selecting the
target table per record.

ForceNew -- changing the table routing replaces the delivery stream.

- rule: {"repeated":{"minItems":"1"}}

### spec.iceberg.destinationTables[].databaseName

`string` · required

Glue Data Catalog database containing the Iceberg table.

- rule: {"string":{"minLen":"1"}}

### spec.iceberg.destinationTables[].tableName

`string` · required

Name of the Iceberg table.

- rule: {"string":{"minLen":"1"}}

### spec.iceberg.destinationTables[].s3ErrorOutputPrefix

`string`

S3 key prefix for records that fail delivery to this table. Uses the
Firehose expression syntax.

### spec.iceberg.destinationTables[].uniqueKeys

`[]string`

Columns that uniquely identify a row, enabling update/delete semantics:
an incoming record whose unique-key values match an existing row
updates it instead of appending. Leave empty (with append_only) for
pure append workloads.

### spec.iceberg.appendOnly

`bool`

Append-only mode. When true, Firehose only appends new snapshots --
update/delete semantics via unique keys are disabled. ForceNew.

### spec.iceberg.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for Iceberg delivery. Default: 300s interval, 5 MiB.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.iceberg.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.iceberg.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.iceberg.retryDurationInSeconds

`int32`

Retry duration in seconds for failed Iceberg commits.
Range: 0-7200. Default: 300 seconds.

### spec.iceberg.s3BackupMode

`string`

S3 backup mode. Controls when records are written to S3.

Valid values:
- "FailedDataOnly" (default) -- only records that fail Iceberg delivery
  are backed up to S3.
- "AllData" -- all records are backed up to S3 in addition to being
  committed to Iceberg.

### spec.iceberg.s3Config

`AwsKinesisFirehoseS3Config` · required

S3 configuration for backing up failed (or all) records. Required.

- rule: {"required":true}
- rule: compression_format must be 'UNCOMPRESSED', 'GZIP', 'ZIP', 'Snappy', or 'HADOOP_SNAPPY' when set

### spec.iceberg.s3Config.bucketArn

`string | valueFrom` · required

S3 bucket ARN where records are delivered.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.iceberg.s3Config.roleArn

`string | valueFrom` · required

IAM role ARN that grants Firehose permission to write to the S3 bucket.
The role must have s3:PutObject, s3:AbortMultipartUpload,
s3:GetBucketLocation, and s3:ListBucket permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.iceberg.s3Config.prefix

`string`

S3 key prefix prepended to delivered objects. Supports Firehose expression
syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").

### spec.iceberg.s3Config.errorOutputPrefix

`string`

S3 key prefix for error output. When Firehose cannot deliver or transform
a record, it writes to this prefix. Uses the same expression syntax as prefix.

### spec.iceberg.s3Config.compressionFormat

`string`

Compression format for delivered objects. Applied before writing to S3.
Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
Default: "UNCOMPRESSED".

### spec.iceberg.s3Config.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
When absent, S3 uses its default encryption settings (SSE-S3 or bucket
default encryption).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.iceberg.s3Config.buffering

`AwsKinesisFirehoseBufferingHints`

Buffering hints for S3 delivery.

- rule: interval_in_seconds must be between 0 and 900 when set
- rule: size_in_mbs must be between 1 and 128 when set

### spec.iceberg.s3Config.buffering.intervalInSeconds

`int32`

Buffer interval in seconds. Firehose flushes when this time elapses since
the last flush, even if the buffer size threshold has not been reached.

Range: 0-900 seconds. Default varies by destination (typically 300;
Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
values improve batching efficiency and reduce S3 object count.

Some destinations enforce a tighter maximum (Splunk: 60s) -- the
destination message carries that rule.

### spec.iceberg.s3Config.buffering.sizeInMbs

`int32`

Buffer size in MiB. Firehose flushes when the accumulated data reaches
this threshold.

Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
for query engines); smaller buffers provide faster delivery.

Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.

### spec.iceberg.s3Config.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch logging configuration for S3 delivery errors.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.iceberg.s3Config.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.iceberg.s3Config.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.iceberg.s3Config.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

### spec.iceberg.processing

`AwsKinesisFirehoseProcessing`

Record-transformation pipeline applied before the Iceberg commit. Use a
metadata_extraction processor to produce per-record table routing when
multiple destination tables are configured.

- rule: processors require enabled to be true
- rule: at least one processor is required when processing is enabled

### spec.iceberg.processing.enabled

`bool`

Enable the processing pipeline. When true, at least one processor should
be configured.

### spec.iceberg.processing.processors

`[]AwsKinesisFirehoseProcessor`

Ordered list of processors. Each entry configures exactly one processor
type; Firehose executes them in order.

- rule: exactly one processor must be set: lambda, metadata_extraction, decompression, cloudwatch_log_processing, append_delimiter, or record_deaggregation

### spec.iceberg.processing.processors[].lambda

`AwsKinesisFirehoseLambdaProcessor`

Invoke an AWS Lambda function to transform records. The function
receives batches of records and returns transformed records with a
status (Ok, Dropped, ProcessingFailed) per record.

- rule: buffer_size_in_mbs must be between 0.2 and 3 when set
- rule: buffer_interval_in_seconds must be between 60 and 900 when set
- rule: number_of_retries must be between 0 and 300 when set

### spec.iceberg.processing.processors[].lambda.lambdaArn

`string | valueFrom` · required

ARN of the Lambda function that transforms records. May include a
version or alias qualifier to pin the deployed transformation.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.iceberg.processing.processors[].lambda.bufferSizeInMbs

`double`

Buffer size in MiB that Firehose accumulates before invoking Lambda.
Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
destination is Splunk).

Smaller buffers invoke Lambda more frequently with smaller batches.
Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
invocation costs.

### spec.iceberg.processing.processors[].lambda.bufferIntervalInSeconds

`int32`

Buffer interval in seconds. Firehose invokes Lambda when this interval
elapses, even if the buffer size threshold has not been reached.
Range: 60-900 seconds. Default: 60 seconds.

### spec.iceberg.processing.processors[].lambda.numberOfRetries

`int32`

Number of times Firehose retries a failed Lambda invocation before
writing the record to the error output prefix.
Range: 0-300. Default: 3.

### spec.iceberg.processing.processors[].lambda.roleArn

`string | valueFrom`

IAM role ARN Firehose assumes to invoke the Lambda function. When
absent, Firehose uses the delivery stream's destination role -- the
right choice for almost every pipeline. Set this only when the
transformation function must be invoked with a DIFFERENT role than the
one that writes to the destination (e.g. the function lives in another
account). Note: AWS reports the delivery role back for unset values, and
the provider does not store default-valued processor parameters in
state -- so set this only to a non-default role, never to the delivery
role itself (that would cause perpetual plan diffs).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.iceberg.processing.processors[].metadataExtraction

`AwsKinesisFirehoseMetadataExtractionProcessor`

Extract partition keys from JSON records with a JQ expression. Used
with Extended S3 dynamic partitioning -- extracted keys are referenced
in the S3 prefix as !{partitionKeyFromQuery:<key>}.

- rule: json_parsing_engine must be 'JQ-1.6' when set

### spec.iceberg.processing.processors[].metadataExtraction.query

`string` · required

JQ expression that extracts partition keys from each JSON record.
The result must be an object whose keys become partition keys.

Example: "{customer_id: .customer_id, event_type: .type}" extracts two
keys, referenced in the prefix as
"data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".

- rule: {"string":{"minLen":"1"}}

### spec.iceberg.processing.processors[].metadataExtraction.jsonParsingEngine

`string`

JSON parsing engine used to evaluate the query.
Valid value: "JQ-1.6" (default -- the only engine AWS supports today).

### spec.iceberg.processing.processors[].decompression

`AwsKinesisFirehoseDecompressionProcessor`

Decompress GZIP-compressed records before delivery. Typically the first
processor when the source sends compressed payloads (e.g., CloudWatch
Logs subscription filters).

- rule: compression_format must be 'GZIP'

### spec.iceberg.processing.processors[].decompression.compressionFormat

`string` · required

Compression format of the incoming records.
Valid value: "GZIP" (the only format AWS supports today).

- rule: {"string":{"minLen":"1"}}

### spec.iceberg.processing.processors[].cloudwatchLogProcessing

`AwsKinesisFirehoseCloudwatchLogProcessingProcessor`

Unwrap CloudWatch Logs subscription envelopes into individual log
events. Use after a decompression processor when the source is a
CloudWatch Logs subscription filter.

### spec.iceberg.processing.processors[].cloudwatchLogProcessing.dataMessageExtraction

`bool`

When true, extract only the log event message field, discarding the
CloudWatch envelope metadata. When false, records pass through with the
envelope intact.

### spec.iceberg.processing.processors[].appendDelimiter

`AwsKinesisFirehoseAppendDelimiterProcessor`

Append a delimiter to every record. Use to produce newline-delimited
JSON (JSON lines) output for query engines and log consumers.

Only supported on the extended_s3 destination -- delimiting is an
S3-object formatting concern; other destinations frame records
natively.

### spec.iceberg.processing.processors[].appendDelimiter.delimiter

`string` · required

Delimiter appended to each record. Use "\\n" for newline-delimited
output -- the format Athena, Spark, and most log consumers expect.

- rule: {"string":{"minLen":"1"}}

### spec.iceberg.processing.processors[].recordDeaggregation

`AwsKinesisFirehoseRecordDeaggregationProcessor`

Split multi-record aggregates (e.g., KPL-aggregated or delimited
payloads) into individual records before further processing. Required
before dynamic partitioning when producers aggregate records.

Only supported on the extended_s3 destination -- AWS rejects it for
every other destination type at creation.

- rule: sub_record_type must be 'JSON' or 'DELIMITED'
- rule: delimiter is required when sub_record_type is 'DELIMITED'

### spec.iceberg.processing.processors[].recordDeaggregation.subRecordType

`string` · required

How records are aggregated in the payload:
- "JSON" -- concatenated JSON documents (no delimiter needed).
- "DELIMITED" -- records separated by a custom delimiter; requires
  the delimiter field.

- rule: {"string":{"minLen":"1"}}

### spec.iceberg.processing.processors[].recordDeaggregation.delimiter

`string`

Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
newline). Required when sub_record_type is "DELIMITED".

### spec.iceberg.logging

`AwsKinesisFirehoseCloudwatchLogging`

CloudWatch error logging for Iceberg delivery failures.

- rule: log_group_name is required when logging is enabled
- rule: log_stream_name is required when logging is enabled

### spec.iceberg.logging.enabled

`bool`

Enable CloudWatch error logging for this delivery target.

### spec.iceberg.logging.logGroupName

`string`

CloudWatch Logs log group name where errors are published.
Required when enabled is true. Firehose neither creates nor
validates this group: CreateDeliveryStream accepts a nonexistent
name (live-verified 2026-08-12), and delivery errors are silently
dropped until a log group with exactly this name exists -- create
it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
trail you configured here never materializes.

### spec.iceberg.logging.logStreamName

`string`

CloudWatch Logs log stream name within the log group.
Required when enabled is true.

## Validation Rules

- `single_source`: kinesis_stream_source and msk_source are mutually exclusive
- `sse_conflicts_with_stream_source`: sse_enabled must be false when kinesis_stream_source or msk_source is configured (the source handles encryption)
- `sse_key_requires_enabled`: sse_kms_key_arn requires sse_enabled to be true

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsKinesisFirehose, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.delivery_stream_arn` | `string` | The Amazon Resource Name (ARN) of the delivery stream. Used for IAM policies, CloudWatch alarm dimensions, and as a reference in other resources that need to send data to this delivery stream. |
| `status.outputs.delivery_stream_name` | `string` | The name of the delivery stream. Used for Firehose API calls (PutRecord, PutRecordBatch) and for human-readable identification. The name is unique within an AWS account and region. |
| `status.outputs.destination_id` | `string` | Identifier of the destination configuration within the delivery stream. AWS assigns it at creation (e.g., "destinationId-000000000001"); the UpdateDestination API requires it when modifying destination settings out of band. |
| `status.outputs.version_id` | `string` | Version of the delivery stream configuration. AWS increments it on every configuration update; the UpdateDestination API requires the current version as an optimistic-concurrency token. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kinesisStreamSource.streamArn` | AwsKinesisStream | `status.outputs.stream_arn` |
| `spec.kinesisStreamSource.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.mskSource.mskClusterArn` | AwsMskCluster | `status.outputs.cluster_arn` |
| `spec.mskSource.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.sseKmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.extendedS3.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.extendedS3.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.extendedS3.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.extendedS3.s3Backup.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.extendedS3.s3Backup.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.extendedS3.s3Backup.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.extendedS3.processing.processors[].lambda.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.extendedS3.processing.processors[].lambda.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.extendedS3.dataFormatConversion.schema.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.opensearch.domainArn` | AwsOpenSearchDomain | `status.outputs.domain_arn` |
| `spec.opensearch.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.opensearch.s3Config.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.opensearch.s3Config.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.opensearch.s3Config.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.opensearch.processing.processors[].lambda.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.opensearch.processing.processors[].lambda.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.opensearch.vpcConfig.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.opensearch.vpcConfig.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.opensearch.vpcConfig.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.opensearchServerless.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.opensearchServerless.s3Config.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.opensearchServerless.s3Config.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.opensearchServerless.s3Config.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.opensearchServerless.processing.processors[].lambda.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.opensearchServerless.processing.processors[].lambda.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.opensearchServerless.vpcConfig.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.opensearchServerless.vpcConfig.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.opensearchServerless.vpcConfig.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.httpEndpoint.secretsManager.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.httpEndpoint.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.httpEndpoint.s3Config.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.httpEndpoint.s3Config.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.httpEndpoint.s3Config.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.httpEndpoint.processing.processors[].lambda.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.httpEndpoint.processing.processors[].lambda.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.redshift.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.redshift.secretsManager.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.redshift.s3Config.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.redshift.s3Config.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.redshift.s3Config.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.redshift.s3Backup.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.redshift.s3Backup.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.redshift.s3Backup.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.redshift.processing.processors[].lambda.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.redshift.processing.processors[].lambda.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.splunk.secretsManager.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.splunk.s3Config.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.splunk.s3Config.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.splunk.s3Config.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.splunk.processing.processors[].lambda.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.splunk.processing.processors[].lambda.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.snowflake.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.snowflake.secretsManager.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.snowflake.s3Config.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.snowflake.s3Config.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.snowflake.s3Config.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.snowflake.processing.processors[].lambda.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.snowflake.processing.processors[].lambda.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.iceberg.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.iceberg.s3Config.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.iceberg.s3Config.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.iceberg.s3Config.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.iceberg.processing.processors[].lambda.lambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.iceberg.processing.processors[].lambda.roleArn` | AwsIamRole | `status.outputs.role_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsCognitoUserPool | `spec.logConfigurations[].firehoseStreamArn` | `status.outputs.delivery_stream_arn` |
| AwsEventBridgePipe | `spec.logConfiguration.firehose.deliveryStreamArn` | `status.outputs.delivery_stream_arn` |
| AwsMskCluster | `spec.logging.firehose.deliveryStream` | `status.outputs.delivery_stream_name` |
| AwsSesConfigurationSet | `spec.eventDestinations[].firehose.deliveryStream` | `status.outputs.delivery_stream_arn` |

## See Also

- [Overview](../README.md)
