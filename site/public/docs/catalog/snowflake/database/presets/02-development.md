---
title: "Development Database"
description: "This preset creates a transient Snowflake database optimized for development. Transient databases have no Fail-safe period, reducing storage costs. Debug logging and console output are enabled for..."
type: "preset"
rank: "02"
presetSlug: "02-development"
componentSlug: "database"
componentTitle: "Database"
provider: "snowflake"
icon: "package"
order: 2
---

# Development Database

This preset creates a transient Snowflake database optimized for development. Transient databases have no Fail-safe period, reducing storage costs. Debug logging and console output are enabled for troubleshooting stored procedures and tasks during development.

## When to Use

- Development and testing environments
- Proof-of-concept and experimentation databases
- CI/CD pipeline databases that are frequently created and destroyed
- Any scenario where Fail-safe protection is unnecessary and cost savings are preferred

## Key Configuration Choices

- **Transient** (`isTransient: true`) -- no Fail-safe period; data is not recoverable after Time Travel expires. Significantly reduces storage costs for dev databases
- **1-day Time Travel** (`dataRetentionTimeInDays: 1`) -- minimal retention to save storage; sufficient for "oops I dropped something" recovery in dev
- **Debug logging** (`logLevel: DEBUG`) -- verbose logging for troubleshooting stored procedures, UDFs, and tasks
- **Event tracing** (`traceLevel: ON_EVENT`) -- captures trace events triggered by code; useful for debugging complex pipelines
- **Console output** (`enableConsoleOutput: true`) -- routes `stdout`/`stderr` from anonymous stored procedures to the event table for debugging

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `DEVELOPMENT` | Database name (uppercase Snowflake convention) | Your naming convention |

## Related Presets

- **01-production** -- Use instead for production with full retention and Fail-safe protection
- **03-iceberg-analytics** -- Use instead for Iceberg table integration
