# AWS Glue Catalog Database

Deploys an AWS Glue Data Catalog database — a metadata container that organizes
table definitions for data stored in Amazon S3, Redshift, RDS, and other data
stores. The Glue Data Catalog is the namespace layer that Amazon Athena, AWS Glue
ETL, Glue Crawlers, Redshift Spectrum, and Amazon EMR use to discover and query
data.

## When to Use

Use a Glue Catalog Database to:

- **Organize a data lake**: Group table definitions by domain (sales, marketing,
  clickstream) so Athena queries and Glue ETL jobs can discover data by
  `database.table` naming.
- **Set default storage locations**: Define a shared S3 prefix so crawlers and
  CREATE TABLE statements inherit a consistent base path.
- **Enable analytics workflows**: Athena workgroups, Glue crawlers, and Glue ETL
  jobs all operate within the context of a catalog database.
- **Namespace isolation**: Separate development, staging, and production table
  metadata into distinct databases within the same AWS account.

## Prerequisites

- An AWS account with Glue Data Catalog access (enabled by default in all regions)
- An S3 bucket (if setting `location_uri` for default table storage)

## Spec Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | "" | Human-readable description of the database (max 2048 chars) |
| `location_uri` | string | "" | Default S3 URI for tables (e.g., `s3://bucket/prefix/`) |

### ForceNew Fields

- **Database name** (from `metadata.name`) — Cannot be changed after creation.
  Must be 1-255 characters, lowercase letters, numbers, and underscores only.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `database_name` | Name of the Glue Data Catalog database |
| `database_arn` | ARN of the database (for IAM policies, Lake Formation) |
| `catalog_id` | ID of the Glue Data Catalog (AWS Account ID) |

## Deliberately Omitted (v1)

| Feature | Reason |
|---------|--------|
| `create_table_default_permission` | Lake Formation governance; default IAM behavior covers >80% of use cases |
| `federated_database` | Redshift Data Share federation, ~5% adoption, requires Lake Formation |
| `target_database` | Cross-region/cross-account references, ~5% adoption |
| `parameters` | Generic key-value metadata, rarely set by users directly |
| `catalog_id` | Defaults to AWS Account ID; cross-account scenarios covered by target_database |

These can be added in v2 based on demand.

## Related Resources

- **AwsAthenaWorkgroup** — Queries data described by tables in this database
- **AwsS3Bucket** — Storage layer for data referenced by Glue tables
- **AwsKmsKey** — Encryption for data at rest in S3 (referenced by table definitions)
- **AwsRedshiftCluster** — Redshift Spectrum queries the Glue catalog for external tables
