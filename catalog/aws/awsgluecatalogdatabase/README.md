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
- **Govern new tables**: Set default Lake Formation permissions for tables
  created in the database — or disable the IAM-compatibility grant entirely
  when hardening a Lake Formation-managed lake.
- **Consume shared data**: Create a resource link to a database another
  account or region shared via AWS RAM, or project a Redshift datashare into
  the catalog as a federated database.
- **Namespace isolation**: Separate development, staging, and production table
  metadata into distinct databases within the same AWS account.

## Prerequisites

- An AWS account with Glue Data Catalog access (enabled by default in all regions)
- An S3 bucket (if setting `location_uri` for default table storage)
- An AWS RAM / Lake Formation share (only for `target_database` resource links)
- A Redshift datashare and federation connection (only for `federated_database`)

## Spec Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | "" | Human-readable description of the database (max 2048 chars) |
| `catalog_id` | string | account ID | Catalog to create the database in; set only for cross-account catalogs (ForceNew) |
| `location_uri` | string | "" | Default S3 URI for tables (e.g., `s3://bucket/prefix/`) |
| `parameters` | map | — | Catalog metadata properties read by engines and governance tooling |
| `create_table_default_permissions` | list | AWS default | Default Lake Formation grants on tables created here; an empty-permissions entry disables the `IAM_ALLOWED_PRINCIPALS` grant |
| `target_database` | object | — | Resource link to a shared database: `catalog_id`, `database_name`, optional `region` (all ForceNew) |
| `federated_database` | object | — | Federated projection of an external source: `identifier`, `connection_name` |

A database is exactly one shape — regular, resource link, or federated;
`target_database` and `federated_database` cannot be combined (enforced at
manifest validation).

### ForceNew Fields

- **Database name** (from `metadata.name`) — cannot be changed after creation.
  1-255 characters; AWS rejects uppercase letters.
- **`catalog_id`** and **`target_database`** — fixed at creation.

## Outputs

| Output | Description |
|--------|-------------|
| `database_name` | Name of the Glue Data Catalog database |
| `database_arn` | ARN of the database (for IAM policies, Lake Formation) |
| `catalog_id` | ID of the Glue Data Catalog (AWS Account ID) |

## Related Resources

- **AwsAthenaWorkgroup** — Queries data described by tables in this database
- **AwsS3Bucket** — Storage layer for data referenced by Glue tables
- **AwsKmsKey** — Encryption for data at rest in S3 (referenced by table definitions)
- **AwsRedshiftCluster** — Redshift Spectrum queries the Glue catalog for external tables

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
