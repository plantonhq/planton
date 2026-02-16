---
title: "PostgreSQL Development Instance"
description: "This preset creates a minimal Cloud SQL PostgreSQL instance for development and testing. It uses the smallest available tier, public IP access (via Cloud SQL Proxy), and basic backups. No high..."
type: "preset"
rank: "03"
presetSlug: "03-postgresql-development"
componentSlug: "cloud-sql"
componentTitle: "Cloud SQL"
provider: "gcp"
icon: "package"
order: 3
---

# PostgreSQL Development Instance

This preset creates a minimal Cloud SQL PostgreSQL instance for development and testing. It uses the smallest available tier, public IP access (via Cloud SQL Proxy), and basic backups. No high availability, no deletion protection, and no query insights -- optimized for cost.

## When to Use

- Development and testing databases
- Local development connecting via Cloud SQL Auth Proxy
- Prototype or proof-of-concept environments

## Key Configuration Choices

- **Smallest tier** (`tier: db-f1-micro`) -- shared-core with 0.6 GB RAM; cheapest option
- **10 GB storage** -- minimum size with autoresize enabled for growth
- **Public IP** (`ipv4Enabled: true`) -- accessible via Cloud SQL Auth Proxy for developers
- **No HA** -- single-zone instance; no failover needed for development
- **No deletion protection** -- dev instances should be easy to tear down
- **Basic backups** -- 3-day retention, no PITR

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |

## Related Presets

- **01-postgresql-production** -- Use for production with HA, private IP, PITR, and deletion protection
- **02-mysql-production** -- Use for MySQL production workloads
