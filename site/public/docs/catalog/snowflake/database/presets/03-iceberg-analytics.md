---
title: "Iceberg Analytics Database"
description: "This preset creates a Snowflake database configured for Apache Iceberg tables with Snowflake as the catalog. Iceberg tables store data in an open format (Parquet) on external storage, enabling..."
type: "preset"
rank: "03"
presetSlug: "03-iceberg-analytics"
componentSlug: "database"
componentTitle: "Database"
provider: "snowflake"
icon: "package"
order: 3
---

# Iceberg Analytics Database

This preset creates a Snowflake database configured for Apache Iceberg tables with Snowflake as the catalog. Iceberg tables store data in an open format (Parquet) on external storage, enabling interoperability with Spark, Trino, and other compute engines while leveraging Snowflake's query engine.

## When to Use

- Analytics workloads using Apache Iceberg open table format
- Data lakehouse architectures where data must be accessible by multiple compute engines
- Scenarios where Snowflake manages the Iceberg catalog but data lives on external storage (S3, GCS, Azure Blob)

## Key Configuration Choices

- **Snowflake catalog** (`catalog: SNOWFLAKE`) -- Snowflake manages the Iceberg metadata catalog; use for Snowflake-managed Iceberg tables
- **External volume** (`externalVolume`) -- required for Iceberg; specifies the external storage location (S3, GCS, or Azure Blob) where Iceberg data files are stored
- **OPTIMIZED serialization** (`storageSerializationPolicy: OPTIMIZED`) -- best performance within Snowflake; use `COMPATIBLE` if third-party engines also read the data
- **Replace invalid characters** (`replaceInvalidCharacters: true`) -- replaces invalid UTF-8 characters with the Unicode replacement character in Iceberg query results
- **7-day Time Travel** (`dataRetentionTimeInDays: 7`) -- shorter retention for analytics databases where source data can be reloaded

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `ANALYTICS_ICEBERG` | Database name (uppercase Snowflake convention) | Your naming convention |
| `<your-external-volume>` | Name of the Snowflake external volume for Iceberg data storage | Snowflake `SHOW EXTERNAL VOLUMES` or your storage admin |

## Related Presets

- **01-production** -- Use instead for standard Snowflake tables without Iceberg
- **02-development** -- Use instead for transient dev/test databases
