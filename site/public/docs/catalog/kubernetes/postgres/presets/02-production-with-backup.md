---
title: "Production PostgreSQL with Backup"
description: "This preset deploys a 2-replica PostgreSQL cluster with streaming replication, daily backups, pre-configured database and user, and production-grade resources."
type: "preset"
rank: "02"
presetSlug: "02-production-with-backup"
componentSlug: "postgres"
componentTitle: "Postgres"
provider: "kubernetes"
icon: "package"
order: 2
---

# Production PostgreSQL with Backup

This preset deploys a 2-replica PostgreSQL cluster with streaming replication, daily backups, pre-configured database and user, and production-grade resources.

## When to Use

- Production workloads requiring high availability via streaming replication
- Applications needing pre-provisioned databases and roles
- Environments where automated backups are mandatory

## Key Configuration Choices

- **2 replicas** -- primary with one streaming replica for read scaling and automatic failover
- **50Gi disk** -- production-sized persistent storage; adjust based on your data growth projections
- **Higher resources** (`250m`/`512Mi` requests, `2000m`/`4Gi` limits) -- production-appropriate for moderate workloads
- **Daily backups** (`0 2 * * *`) -- base backup at 2:00 AM UTC; configure S3/R2 credentials separately if needed
- **Pre-configured database** (`app`) -- owned by `app_owner` role with LOGIN and CREATEDB privileges

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **01-single-instance** -- Minimal single-replica PostgreSQL for development
