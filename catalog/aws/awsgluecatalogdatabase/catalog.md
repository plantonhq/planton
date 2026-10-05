# AWS Glue Catalog Database

Deploys a metadata container in the AWS Glue Data Catalog that organizes table definitions for data stored in S3, Redshift, RDS, and other data stores. The database serves as the namespace layer for Amazon Athena, AWS Glue ETL jobs, Glue Crawlers, Redshift Spectrum, and Amazon EMR queries. One resource covers three shapes — a regular database, a resource link to a database shared from another account, or a federated database projecting an external source — and the shape's coordinates are fixed at creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Glue Catalog Database** -- a metadata namespace in the AWS Glue Data Catalog, in one of three shapes: a regular database (with optional description, default S3 location, free-form catalog `parameters`, and Lake Formation `createTableDefaultPermissions`), a resource link (`targetDatabase`) pointing at a database shared from another account or region, or a federated database (`federatedDatabase`) projecting an external source such as a Redshift datashare into the catalog
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An S3 bucket** (optional) for the default table storage location. When `locationUri` is set, tables created in this database without an explicit location inherit this S3 path as their base directory. Format: `s3://bucket-name/optional-prefix/`.

## Deploy

### Console

Open the deployment store, find **AWS Glue Catalog Database**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Data Catalog** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsGlueCatalogDatabase
metadata:
  name: analytics
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: "Sales analytics data lake — raw and curated tables"
```

```shell
planton apply -f glue-database.yaml
```

This creates a Glue Catalog Database named `analytics` with a description. No default S3 location is configured, so each table must specify its own storage location explicitly. An Infra Job tracks the provisioning and streams progress in real time.

## Key Configuration

These are the most important decisions when configuring a Glue Catalog Database. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Database name** -- Derived from `metadata.name`. Must be 1-255 characters, lowercase letters, numbers, and underscores only. The name is immutable after creation (ForceNew). Choose a name that reflects the logical data domain (e.g., `sales_raw`, `clickstream_events`).

**Default storage location** -- Set `locationUri` to an S3 URI (e.g., `s3://my-datalake/sales/`) to provide a default base path for tables created in this database. Recommended for organized data lakes where all tables in a database share a common S3 prefix. When omitted, each table must specify its own location.

**Description** -- A human-readable description (up to 2048 characters) helps teams understand the purpose and contents of the catalog namespace. Visible in the AWS Glue console, Athena, and API responses.

**Database shape** -- A database is exactly one of three shapes. A regular database is the common case. A resource link (`targetDatabase` with the owning account's catalog ID and database name) makes a database shared via AWS RAM / Lake Formation queryable locally; it carries no storage or schema of its own. A federated database (`federatedDatabase` with a source identifier and Glue connection, e.g. a Redshift datashare over `aws:redshift`) projects an external source into the catalog without copying data. The shape's coordinates are fixed at creation.

**Catalog parameters** -- Free-form key-value properties (`parameters`) attached to the database entry and read by engines and governance tooling: classification hints, team ownership labels, engine-specific switches. These are catalog metadata, not AWS resource tags.

**Lake Formation default grants** -- `createTableDefaultPermissions` sets the permissions tables created in this database start with. When omitted, AWS grants ALL to the virtual group `IAM_ALLOWED_PRINCIPALS` (the IAM-compatibility mode). A single entry with an empty permission list and no principal disables that default -- the recommended hardening step when migrating to Lake Formation permissions.

**Cross-account catalog** -- `catalogId` creates the database inside another account's Data Catalog (requires a matching catalog resource policy there). Fixed at creation; changing it replaces the database.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — `locationUri` is a plain S3 URI with a user-defined path prefix, and the resource-link and federated-database coordinates are catalog identifiers, not resource references.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `database_name` | Glue Data Catalog database name | Athena queries (`FROM database.table`), Glue crawler targets, ETL job scripts |
| `database_arn` | Amazon Resource Name of the database | IAM policies, Lake Formation permissions |
| `catalog_id` | AWS Glue Data Catalog ID (AWS Account ID) | Glue crawler configuration, cross-account catalog references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic data catalog** -- A minimal database with just a description. Tables are created by Glue Crawlers, Athena CTAS statements, or Glue ETL jobs, each specifying its own S3 location. Start from the **Basic Data Catalog** preset.

**S3 data lake** -- A database with a default `locationUri` pointing to a shared S3 prefix, classification `parameters`, and Lake Formation default grants. New tables inherit this base path, keeping the data lake organized under a consistent directory structure. Start from the **S3 Data Lake** preset.

**Shared database link** -- A resource link to a database another account shared via AWS RAM / Lake Formation, making its tables queryable from Athena, Redshift Spectrum, and EMR in this account. Start from the **Shared Database Link** preset.

## Works With

The database carries no foreign key references of its own, but it genuinely co-deploys with:

- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- backs the `locationUri` default table storage path and holds the data the catalog describes
- [**AWS Athena Workgroup**](/infra-catalog/aws-athena-workgroup) -- runs SQL queries against the tables cataloged in this database