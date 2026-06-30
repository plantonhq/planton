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
- **Daily backups** (`0 2 * * *`) -- base backup at 2:00 AM UTC, retaining 14 base backups, streamed by WAL-G to an S3-compatible bucket
- **Referenceable bucket** -- `bucket` is a value-or-ref: use a literal name as shown, or wire it to a `CloudflareR2Bucket` (or any S3-compatible bucket) output via `valueFrom`
- **Pre-configured database** (`app`) -- owned by `app_owner` role with LOGIN and CREATEDB privileges

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
| `my-postgres-backups` | Backup bucket name | Your S3-compatible bucket, or reference a `CloudflareR2Bucket` resource |
| `<cloudflare-account-id>` | Cloudflare account that owns the R2 bucket | Cloudflare dashboard |
| `<r2-access-key-id>` | R2 access-key ID | Cloudflare R2 API token |
| `<r2-secret-access-key>` | R2 secret access key | Cloudflare R2 API token (store as a managed secret) |

## Related Presets

- **01-single-instance** -- Minimal single-replica PostgreSQL for development
