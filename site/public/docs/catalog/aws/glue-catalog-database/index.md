---
title: "Glue Catalog Database"
description: "Glue Catalog Database deployment documentation"
icon: "package"
order: 100
componentName: "awsgluecatalogdatabase"
---

# AWS Glue Catalog Database

Deploys an AWS Glue Data Catalog database — a metadata namespace that organizes table definitions for data stored in S3, Redshift, RDS, and other data stores. The database is the namespace that Amazon Athena, Glue Crawlers, Glue ETL jobs, and Redshift Spectrum use to discover and query data via `database.table` naming.

## What Gets Created

When you deploy an AwsGlueCatalogDatabase resource, Planton provisions:

- **Glue Catalog Database** — an `aws_glue_catalog_database` resource registered in the AWS Glue Data Catalog with the specified name, description, and optional default storage location

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **An S3 bucket** if setting `locationUri` for default table storage (the bucket must exist before deploying)

## Quick Start

Create a file `glue-catalog-database.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: analytics
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsGlueCatalogDatabase.analytics
spec:
  region: us-east-1
  description: "Analytics data catalog for ad-hoc queries and BI dashboards"
```

Deploy:

```shell
planton apply -f glue-catalog-database.yaml
```

This creates a Glue Data Catalog database named `analytics` that Athena workgroups, Glue crawlers, and ETL jobs can use as a namespace for table definitions.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | `string` | AWS region where the Glue Catalog Database will be created (e.g., `us-east-1`) |

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
apiVersion: aws.planton.dev/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: experiments
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: data
    pulumi.planton.dev/stack.name: dev.AwsGlueCatalogDatabase.experiments
spec:
  region: us-east-1
```

### Descriptive Analytics Database

A database with a description documenting its purpose and contents.

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: sales_analytics
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: analytics
    pulumi.planton.dev/stack.name: prod.AwsGlueCatalogDatabase.sales_analytics
spec:
  region: us-east-1
  description: >-
    Sales pipeline data lake — raw ingestion tables, curated transformations,
    and aggregated datasets for BI dashboards and ML feature stores.
```

### Production S3 Data Lake

A production database with a default S3 storage location so all tables inherit a consistent base path. Recommended for organized data lakes.

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: prod_warehouse
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: data-platform
    pulumi.planton.dev/stack.name: prod.AwsGlueCatalogDatabase.prod_warehouse
spec:
  region: us-east-1
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
