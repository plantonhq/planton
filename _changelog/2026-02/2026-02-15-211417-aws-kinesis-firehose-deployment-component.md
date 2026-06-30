# AWS Kinesis Firehose Deployment Component

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Terraform Module, Documentation

## Summary

Added AwsKinesisFirehose (R17) -- a Kinesis Data Firehose delivery stream component supporting 4 destination types (Extended S3, OpenSearch, HTTP Endpoint, Redshift), 2 source modes (Direct PUT, Kinesis Data Stream), Lambda-based record transformation, Parquet/ORC format conversion via Glue catalog, dynamic partitioning, and server-side encryption. This is the most complex single AWS resource in the expansion project, with 16 proto message types and 51 validation tests.

## Problem Statement / Motivation

Kinesis Data Firehose is the standard AWS service for loading streaming data into storage and analytics destinations without writing custom consumer code. The Planton AWS provider lacked this component, leaving a gap in data pipeline coverage. With AwsKinesisStream (R16) and AwsKinesisStreamConsumer (R16a) already implemented, Firehose completes the Kinesis service family and enables end-to-end streaming data pipelines via infra charts.

### Pain Points

- No way to deploy managed streaming ETL pipelines through Planton
- Users had to manually configure complex Firehose delivery streams with 9 possible destination types
- Missing the bridge between Kinesis Data Streams (source) and storage/analytics destinations (S3, OpenSearch, Redshift)

## Solution / What's New

A complete deployment component covering the 4 most common Firehose destination types (~90% of real-world usage), with full IaC parity between Pulumi and Terraform, comprehensive validation, and production-quality documentation.

### Destination Coverage

| Destination | Usage | Coverage |
|-------------|-------|----------|
| Extended S3 | ~60% | Full (compression, format conversion, dynamic partitioning, Lambda processing) |
| OpenSearch | ~15% | Full (index rotation, VPC delivery, Lambda processing) |
| HTTP Endpoint | ~10% | Full (custom headers, content encoding, Lambda processing) |
| Redshift | ~5-8% | Full (COPY command, credentials, S3 staging + backup) |

### Deferred Destinations (v2)

Elasticsearch (deprecated), OpenSearch Serverless, Splunk, Snowflake, Iceberg -- collectively <10% of usage.

## Implementation Details

### Proto API (16 messages, ~80 fields, ~25 CEL validations)

The spec uses a proto `oneof destination_config` with 4 typed destination messages. Shared sub-messages reduce duplication:

- **AwsKinesisFirehoseBufferingHints** -- interval + size, reused across destinations
- **AwsKinesisFirehoseS3Config** -- S3 backup, reused by OpenSearch, HTTP, Redshift
- **AwsKinesisFirehoseLambdaProcessing** -- simplified Lambda-only processing model
- **AwsKinesisFirehoseCloudwatchLogging** -- error logging, reused across destinations
- **AwsKinesisFirehoseVpcConfig** -- ENI placement for VPC OpenSearch delivery

Extended S3 has additional sub-messages for dynamic partitioning, data format conversion, and Glue schema reference.

### Key Design Decisions

- **4 first-class destinations via oneof** -- not 9. The 80/20 rule: 4 destinations cover ~90% of real usage. Deferred destinations can be added as additional oneof fields without breaking changes.
- **No separate destination_type field** -- IaC modules derive the TF/Pulumi destination string from which oneof field is set. No redundancy for the user.
- **Lambda-only processing** -- simplified from the generic processor model (type + parameters). ~95% of processing configs use Lambda; other processor types deferred to v2.
- **SSE conflicts with Kinesis source enforced via CEL** -- when using a Kinesis stream source, the source handles encryption. This AWS constraint is surfaced as a validation error rather than a runtime failure.
- **MSK source excluded from v1** -- <5% of Firehose streams, adds authentication complexity. Can be added to the source config in v2.

### Surprise Findings (8)

1. **9 destination types in TF provider** (T02 only mentioned 3)
2. **Destination type is ForceNew** -- cannot change after creation
3. **SSE conflicts with Kinesis/MSK sources** -- mutual exclusion at the AWS API level
4. **Dynamic partitioning is ForceNew** -- must be decided at creation time
5. **Data format conversion requires Glue catalog** -- deeply nested config
6. **MSK source exists** (not in T02 guidance)
7. **Processing/logging are per-destination, not global** -- each destination has its own config
8. **Most complex single AWS resource** -- 4,400 lines in TF provider schema alone

### File Metrics

- **43 files** in `apis/dev/planton/provider/aws/awskinesisfirehose/v1/`
- **51 validation tests** (21 happy path + 26 failure scenarios + 4 API envelope)
- **8 Pulumi Go files** (main, locals, outputs, delivery_stream, extended_s3, opensearch, http_endpoint, redshift, processing)
- **5 Terraform files** (main, locals, outputs, variables, provider)
- **4 presets** (S3 data lake, OpenSearch log analytics, HTTP webhook, S3 Parquet analytics)
- **Enum registration**: AwsKinesisFirehose = 261

## Benefits

- Complete Kinesis service family: Stream (R16) + Consumer (R16a) + Firehose (R17)
- Enables data pipeline infra charts: Kinesis -> Firehose -> S3/OpenSearch/Redshift
- 15+ StringValueOrRef fields for infra-chart composability
- Rich CEL validation catches config errors at validate-time, not deploy-time
- Both Pulumi and Terraform modules with feature parity

## Impact

- **Users**: Can now deploy managed streaming ETL pipelines with a single YAML manifest
- **Infra Charts**: Enables data-pipeline chart patterns (streaming source -> transformation -> analytics destination)
- **Platform**: 21st new AWS resource kind in the expansion project; Phase 2 continues

## Related Work

- R16 AwsKinesisStream -- Kinesis Data Stream (source for Firehose)
- R16a AwsKinesisStreamConsumer -- Enhanced fan-out consumer (alternative to Firehose)
- R08 AwsOpenSearchDomain -- OpenSearch destination (referenced via StringValueOrRef)
- R01 AwsSqsQueue, R02 AwsSnsTopic -- Messaging siblings in the same expansion project

---

**Status**: Production Ready
