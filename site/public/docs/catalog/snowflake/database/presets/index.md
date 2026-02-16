---
title: "Presets"
description: "Ready-to-deploy configuration presets for Database"
type: "preset-list"
componentSlug: "database"
componentTitle: "Database"
provider: "snowflake"
icon: "package"
order: 200
presets:
  - slug: "01-production"
    rank: "01"
    title: "Production Database"
    excerpt: "This preset creates a production Snowflake database with 30-day Time Travel retention, warning-level logging, and the public schema dropped on creation. This is the standard configuration for..."
  - slug: "02-development"
    rank: "02"
    title: "Development Database"
    excerpt: "This preset creates a transient Snowflake database optimized for development. Transient databases have no Fail-safe period, reducing storage costs. Debug logging and console output are enabled for..."
  - slug: "03-iceberg-analytics"
    rank: "03"
    title: "Iceberg Analytics Database"
    excerpt: "This preset creates a Snowflake database configured for Apache Iceberg tables with Snowflake as the catalog. Iceberg tables store data in an open format (Parquet) on external storage, enabling..."
---

# Database Presets

Ready-to-deploy configuration presets for Database. Each preset is a complete manifest you can copy, customize, and deploy.
