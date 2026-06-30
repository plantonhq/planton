# OpenSearch Log Analytics Preset

OpenSearch destination for centralized log indexing with near real-time search. Indexes logs directly into OpenSearch with daily rotation and S3 backup for failed documents.

## When to Use

- **Centralized logging** — Aggregate logs from multiple services into a single searchable index
- **ELK alternative** — OpenSearch (and OpenSearch Dashboards) as a managed Elasticsearch replacement
- **Log analytics** — Full-text search, filtering, and dashboards over application logs
- **Security and compliance** — Retain and search audit logs with FailedDocumentsOnly backup

## Key Configuration

- **Direct PUT source** — Applications send logs via PutRecord/PutRecordBatch
- **domain_arn reference** — Use `valueFrom` to reference an AwsOpenSearchDomain resource
- **OneDay rotation** — Creates daily indices (e.g., `application-logs-2026-02-15`) for retention management
- **FailedDocumentsOnly backup** — Only documents that fail indexing are written to S3
- **60s buffering** — Near real-time indexing with minimal delay

## Prerequisites

| Resource | Description |
|----------|-------------|
| **OpenSearch domain** | Amazon OpenSearch Service domain (managed or serverless). Can be referenced via `valueFrom` from an AwsOpenSearchDomain resource. |
| **S3 backup bucket** | Bucket for failed document backup. Required for all OpenSearch destinations. |
| **IAM roles** | One role for Firehose to write to OpenSearch and S3; ensure `es:ESHttpPut`, `es:ESHttpGet`, and S3 write permissions. |

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `my-log-domain` | Name of your AwsOpenSearchDomain resource (when using valueFrom) |
| `my-firehose-backup-bucket` | S3 bucket for failed document backup |
| `123456789012` | Your AWS account ID |
| `firehose-opensearch-role` | IAM role for OpenSearch and S3 access |
| `firehose-s3-backup-role` | IAM role for S3 backup (can be same as above) |
