---
title: "Production Database"
description: "This preset creates a production Snowflake database with 30-day Time Travel retention, warning-level logging, and the public schema dropped on creation. This is the standard configuration for..."
type: "preset"
rank: "01"
presetSlug: "01-production"
componentSlug: "database"
componentTitle: "Database"
provider: "snowflake"
icon: "package"
order: 1
---

# Production Database

This preset creates a production Snowflake database with 30-day Time Travel retention, warning-level logging, and the public schema dropped on creation. This is the standard configuration for production data warehouses where data recovery and compliance are important but verbose logging is not needed.

## When to Use

- Production data warehouses and analytics databases
- Environments where 30-day Time Travel provides adequate recovery window
- Standard production setup without Iceberg or heavy task processing

## Key Configuration Choices

- **30-day Time Travel** (`dataRetentionTimeInDays: 30`) -- allows CLONE and UNDROP operations for up to 30 days; Snowflake's maximum for standard edition
- **14-day stream extension** (`maxDataExtensionTimeInDays: 14`) -- prevents streams from going stale on tables with infrequent changes
- **WARN logging** (`logLevel: WARN`) -- captures warnings and errors without verbose debug output; keeps event table manageable
- **Tracing off** (`traceLevel: OFF`) -- no trace events in production; enable for debugging specific issues
- **Drop public schema** (`dropPublicSchemaOnCreation: true`) -- enforces explicit schema design; prevents accidental use of the default `PUBLIC` schema
- **Persistent** (`isTransient` omitted) -- full Fail-safe protection for data recovery beyond Time Travel

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `PRODUCTION` | Database name (uppercase Snowflake convention) | Your naming convention |

## Related Presets

- **02-development** -- Use instead for dev/test with transient storage and debug logging
- **03-iceberg-analytics** -- Use instead when integrating with Apache Iceberg for open table format analytics
