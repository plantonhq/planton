# AWS Glue Catalog Database — Architecture and Design

## Overview

The AWS Glue Data Catalog is AWS's centralized metadata repository for data
assets. A **catalog database** is a namespace within the catalog that groups
related table definitions. It is the organizational unit that Amazon Athena, AWS
Glue ETL, Glue Crawlers, Amazon Redshift Spectrum, and Amazon EMR use to
discover and query data.

A Glue Catalog Database holds no data itself — it holds **metadata about data**:
table schemas, column types, partition keys, storage locations, and serialization
formats. The actual data resides in S3, RDS, Redshift, DynamoDB, or other
storage services.

## Data Catalog Architecture

### Hierarchy

```
AWS Account (Catalog ID)
└── Catalog Database (namespace)
    ├── Table A (schema + S3 location)
    │   ├── Column definitions
    │   ├── Partition keys
    │   └── SerDe information
    ├── Table B (schema + S3 location)
    └── Table C (schema + Redshift connection)
```

Each AWS account has a single Data Catalog per region. Within that catalog,
databases provide namespace isolation. Within each database, tables define the
schema and storage location of data assets.

### Key Relationships

- **Catalog → Database**: One-to-many. A catalog contains many databases.
- **Database → Table**: One-to-many. A database contains many tables.
- **Table → Data**: Each table points to a physical storage location.
- **Crawler → Database**: A crawler populates tables within a target database.
- **Athena → Database**: Athena queries run against tables in a database.

## How It Works

### Database Creation

When you create a Glue Catalog Database, AWS registers the namespace in the Data
Catalog. The database is immediately available for:

1. **Athena queries**: `SELECT * FROM my_database.my_table`
2. **Glue Crawlers**: Crawlers target a database and create/update tables
3. **Glue ETL jobs**: Jobs read from and write to tables in the database
4. **Redshift Spectrum**: External schemas reference Glue databases

### Name as Identifier

The database name is the primary identifier within a catalog. It is:
- **Unique per catalog** (per AWS account per region)
- **Case-sensitive in Athena** (by convention, use lowercase with underscores)
- **ForceNew** — changing the name requires creating a new database

### Default Location URI

The optional `location_uri` field sets a default S3 path for tables. When a
table is created (by a crawler or DDL statement) without an explicit location,
it inherits the database's `location_uri` as its base path.

This is useful for organized data lakes where all tables in a domain share a
common S3 prefix:

```
s3://data-lake/
├── sales/          ← location_uri for "sales" database
│   ├── orders/     ← table "orders" inherits this path
│   ├── customers/  ← table "customers" inherits this path
│   └── products/   ← table "products" inherits this path
└── marketing/      ← location_uri for "marketing" database
    ├── campaigns/
    └── attribution/
```

## Cost Model

**Glue Data Catalog databases are free.** There is no charge for creating or
maintaining databases. Costs arise from:

- **Catalog storage**: First million objects (databases + tables + partitions)
  free per account. After that, $1.00 per 100,000 objects/month.
- **Catalog requests**: First million requests/month free. After that, $1.00 per
  million requests.
- **Glue Crawlers**: Charged per DPU-hour while running.
- **Athena queries**: Charged per TB of data scanned ($5/TB).

For most accounts, the Data Catalog itself is effectively free. The costs are in
the services that use it (Athena, Glue ETL, Crawlers).

## Security Model

### IAM Policies

Access to Glue Catalog Databases is controlled by IAM policies:

```json
{
  "Effect": "Allow",
  "Action": [
    "glue:GetDatabase",
    "glue:GetDatabases",
    "glue:GetTables",
    "glue:GetTable"
  ],
  "Resource": [
    "arn:aws:glue:us-east-1:123456789012:catalog",
    "arn:aws:glue:us-east-1:123456789012:database/my_database",
    "arn:aws:glue:us-east-1:123456789012:table/my_database/*"
  ]
}
```

### Lake Formation (Advanced)

For fine-grained access control beyond IAM, AWS Lake Formation provides:
- Column-level permissions
- Row-level filtering
- Tag-based access control
- Cross-account data sharing

Lake Formation governance is deliberately deferred from v1 of this component
(see `create_table_default_permission` in the v2 roadmap).

## Service Limits

| Limit | Value |
|-------|-------|
| Databases per account per region | 10,000 |
| Tables per database | 3,000,000 |
| Partitions per table | 10,000,000 |
| Database name length | 1-255 characters |
| Database name pattern | Lowercase letters, numbers, underscores |
| Description length | 0-2,048 characters |

## Common Patterns

### Pattern 1: Athena Analytics Stack

```
Glue Catalog Database → Tables (via Crawler) → Athena Workgroup (queries)
```

The database provides the namespace, crawlers populate table metadata, and Athena
workgroups execute queries against those tables.

### Pattern 2: ETL Data Pipeline

```
S3 (raw) → Glue Crawler → Glue Database (raw_db)
                                ↓
                          Glue ETL Job
                                ↓
                    Glue Database (curated_db) → Athena / Redshift
```

Separate databases for raw and curated data provide clear lineage and access
control boundaries.

### Pattern 3: Multi-Environment Isolation

```
Account
├── sales_dev (database)     ← development tables
├── sales_staging (database) ← staging tables
└── sales_prod (database)    ← production tables
```

Environment-specific database names isolate table metadata. Crawlers and ETL jobs
target the correct database based on the deployment environment.

## v2 Roadmap

Features deliberately omitted from v1 that may be added based on demand:

- **Lake Formation governance** (`create_table_default_permission`) — Fine-grained
  permissions for new tables. Adds governance but increases spec complexity.
- **Federated databases** (`federated_database`) — Cross-service federation for
  Redshift Data Shares. Enables querying Redshift data through the catalog.
- **Cross-account references** (`target_database`) — Link to databases in other
  AWS accounts or regions via Resource Access Manager.
- **Parameters** — Generic key-value metadata for advanced Glue service integration.
