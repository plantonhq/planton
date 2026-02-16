# AWS Glue Catalog Database

Deploys an AWS Glue Data Catalog database — a metadata namespace that organizes table definitions for data stored in S3, Redshift, RDS, and other data stores. The database is the namespace that Amazon Athena, Glue Crawlers, Glue ETL jobs, and Redshift Spectrum use to discover and query data via `database.table` naming.

## What Gets Created

When you deploy an AwsGlueCatalogDatabase resource, OpenMCF provisions:

- **Glue Catalog Database** — an `aws_glue_catalog_database` resource registered in the AWS Glue Data Catalog with the specified name, description, and optional default storage location

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An S3 bucket** if setting `locationUri` for default table storage (the bucket must exist before deploying)

## Quick Start

Create a file `glue-catalog-database.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: analytics
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsGlueCatalogDatabase.analytics
spec:
  description: "Analytics data catalog for ad-hoc queries and BI dashboards"
```

Deploy:

```shell
openmcf apply -f glue-catalog-database.yaml
```

This creates a Glue Data Catalog database named `analytics` that Athena workgroups, Glue crawlers, and ETL jobs can use as a namespace for table definitions.

## Configuration Reference

### Required Fields

No fields are strictly required. A minimal empty `spec: {}` creates a database with just the name from `metadata.name`. However, most practical deployments set at least `description` to document the database's purpose.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the database (max 2048 characters, enforced by AWS API) |
| `locationUri` | `string` | `""` | Default S3 URI for tables created in this database (e.g., `s3://bucket/prefix/`). Tables without an explicit location inherit this path |

### ForceNew Fields

- **Database name** (from `metadata.name`) — Cannot be changed after creation. Must be 1-255 characters: lowercase letters, numbers, and underscores only. No uppercase characters.

## Examples

### Minimal Data Catalog

An empty database for quick experimentation. Tables are added later via Glue Crawlers or DDL statements.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: experiments
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: data
    pulumi.openmcf.org/stack.name: dev.AwsGlueCatalogDatabase.experiments
spec: {}
```

### Descriptive Analytics Database

A database with a description documenting its purpose and contents.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: sales_analytics
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: analytics
    pulumi.openmcf.org/stack.name: prod.AwsGlueCatalogDatabase.sales_analytics
spec:
  description: >-
    Sales pipeline data lake — raw ingestion tables, curated transformations,
    and aggregated datasets for BI dashboards and ML feature stores.
```

### Production S3 Data Lake

A production database with a default S3 storage location so all tables inherit a consistent base path. Recommended for organized data lakes.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: prod_warehouse
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: data-platform
    pulumi.openmcf.org/stack.name: prod.AwsGlueCatalogDatabase.prod_warehouse
spec:
  description: >-
    Production data warehouse — curated datasets from ETL pipelines.
    Tables populated by Glue crawlers on a daily schedule. Accessed by
    Athena workgroups for ad-hoc analytics and Redshift Spectrum for BI.
  locationUri: "s3://acme-prod-data-lake/warehouse/"
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `database_name` | `string` | Name of the Glue Data Catalog database, used in Athena queries (`FROM database.table`), Glue crawler configs, and ETL job scripts |
| `database_arn` | `string` | ARN of the database, used for IAM policies and Lake Formation permissions |
| `catalog_id` | `string` | ID of the Glue Data Catalog (AWS Account ID), used by downstream resources needing catalog context |

## Related Components

- [AWS Athena Workgroup](/docs/catalog/aws/athena-workgroup) — Queries data described by tables in this database
- [AWS S3 Bucket](/docs/catalog/aws/s3-bucket) — Storage layer for data referenced by Glue tables
- [AWS KMS Key](/docs/catalog/aws/kms-key) — Encryption for data at rest in S3
- [AWS Redshift Cluster](/docs/catalog/aws/redshift-cluster) — Redshift Spectrum queries the Glue catalog for external tables
