---
title: "PostgreSQL Database"
description: "This preset creates a Cloud Spanner database with the PostgreSQL-compatible dialect and a 7-day version retention period for extended point-in-time recovery. Ideal for teams with PostgreSQL expertise..."
type: "preset"
rank: "02"
presetSlug: "02-postgresql-database"
componentSlug: "spanner-database"
componentTitle: "Spanner Database"
provider: "gcp"
icon: "package"
order: 2
---

# PostgreSQL Database

This preset creates a Cloud Spanner database with the PostgreSQL-compatible dialect and a 7-day version retention period for extended point-in-time recovery. Ideal for teams with PostgreSQL expertise migrating to Spanner.

## When to Use

- Teams familiar with PostgreSQL syntax and tools
- Migrating existing PostgreSQL applications to Spanner
- Using PostgreSQL-compatible client libraries (psycopg2, JDBC PostgreSQL driver)
- Workloads that benefit from a longer recovery window

## Key Configuration

- **PostgreSQL dialect** -- permanent choice enabling PostgreSQL-compatible SQL syntax and wire protocol
- **7-day version retention** -- maximum point-in-time recovery window for production safety
- **No CMEK** -- uses Google-managed encryption (default)
- **No DDL** -- schema managed separately via PostgreSQL-compatible migration tools

## Important Notes

- The dialect choice is **permanent** and cannot be changed after creation
- Some Spanner-specific features (interleaved tables, STRUCT types) are not available with PostgreSQL dialect
- PostgreSQL mode uses PostgreSQL-style information schema and DDL syntax

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the Spanner instance lives | GCP Console or `GcpProject` outputs |
| `<spanner-instance-name>` | Name of the existing Spanner instance | `GcpSpannerInstance` outputs (`instance_name`) |
| `<database-name>` | Name for this database (2-30 chars, lowercase, hyphens/underscores allowed) | Choose a descriptive name (e.g., `pg-app-db`) |

## Related Presets

- **01-basic-database** -- GoogleSQL with minimal configuration
- **03-cmek-encrypted** -- GoogleSQL with customer-managed encryption and drop protection
