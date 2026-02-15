---
title: "Basic Database"
description: "This preset creates a minimal Cloud Spanner database with the default GoogleSQL dialect and 1-hour version retention. It is the simplest starting point for any Spanner workload."
type: "preset"
rank: "01"
presetSlug: "01-basic-database"
componentSlug: "spanner-database"
componentTitle: "Spanner Database"
provider: "gcp"
icon: "package"
order: 1
---

# Basic Database

This preset creates a minimal Cloud Spanner database with the default GoogleSQL dialect and 1-hour version retention. It is the simplest starting point for any Spanner workload.

## When to Use

- Getting started with Cloud Spanner
- Development and testing databases
- Applications that will manage schema via migration tools
- Quick prototyping where schema is not yet defined

## Key Configuration

- **GoogleSQL dialect** -- default dialect with full Spanner feature support (interleaved tables, STRUCT types)
- **1-hour version retention** -- default point-in-time recovery window
- **No CMEK** -- uses Google-managed encryption (default)
- **No drop protection** -- database can be deleted freely during development

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the Spanner instance lives | GCP Console or `GcpProject` outputs |
| `<spanner-instance-name>` | Name of the existing Spanner instance | `GcpSpannerInstance` outputs (`instance_name`) |
| `<database-name>` | Name for this database (2-30 chars, lowercase, hyphens/underscores allowed) | Choose a descriptive name (e.g., `app-db`) |

## Related Presets

- **02-postgresql-database** -- PostgreSQL dialect with extended retention
- **03-cmek-encrypted** -- GoogleSQL with customer-managed encryption and drop protection
