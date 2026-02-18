# Examples: AWS Glue Catalog Database

## Minimal — Empty Data Catalog Database

The simplest possible configuration. Creates a database with just a name. Useful
for quick experimentation or when Glue Crawlers will handle all table metadata.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: experiments
spec:
  region: us-east-1
```

## Descriptive — Database with Purpose Documentation

Adds a description so team members understand what data lives in this namespace.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: sales_analytics
spec:
  region: us-east-1
  description: >-
    Sales pipeline data lake containing raw ingestion tables, curated
    transformations, and aggregated datasets for BI dashboards.
```

## S3 Data Lake — Default Storage Location

Sets a default S3 location so tables created in this database inherit a
consistent storage prefix. Recommended for organized data lakes.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: clickstream
spec:
  region: us-east-1
  description: "Web and mobile clickstream events for product analytics"
  locationUri: "s3://analytics-data-lake-us-east-1/clickstream/"
```

## Production — Full Configuration

A production data catalog database with description and default storage location
for a well-organized data lake.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: prod_warehouse
  org: acme-corp
  env: production
spec:
  region: us-east-1
  description: >-
    Production data warehouse — curated datasets from ETL pipelines.
    Tables are populated by Glue crawlers and ETL jobs on a daily schedule.
    Accessed by Athena workgroups for ad-hoc analytics and Redshift Spectrum
    for BI reporting.
  locationUri: "s3://acme-prod-data-lake/warehouse/"
```

## Multi-Environment — Namespace Isolation

Use separate databases per environment to isolate table metadata. Crawlers,
Athena queries, and ETL jobs target the correct environment by database name.

```yaml
# Development
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: orders_dev
  env: development
spec:
  region: us-west-2
  description: "Development environment for orders data pipeline"
  locationUri: "s3://dev-data-lake/orders/"
---
# Production
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: orders_prod
  env: production
spec:
  region: us-east-1
  description: "Production orders data — DO NOT modify table schemas without approval"
  locationUri: "s3://prod-data-lake/orders/"
```

## Infra Chart Reference — Data Analytics Stack

In an infra chart, a Glue Catalog Database serves as the metadata foundation
for the analytics stack. Downstream resources (Athena workgroups, Glue crawlers)
reference this database by name.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlueCatalogDatabase
metadata:
  name: "{{ values.database_name }}"
  org: "{{ values.org }}"
  env: "{{ values.env }}"
spec:
  region: "{{ values.region }}"
  description: "{{ values.description }}"
  locationUri: "s3://{{ values.data_lake_bucket }}/{{ values.database_name }}/"
```
