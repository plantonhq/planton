---
title: "Standard Zalando Postgres Operator"
description: "This preset deploys the Zalando Postgres Operator with recommended default resources and no backup configuration. The operator manages PostgreSQL clusters using the `postgresql` custom resource,..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "zalando-postgres-operator"
componentTitle: "Zalando Postgres Operator"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Zalando Postgres Operator

This preset deploys the Zalando Postgres Operator with recommended default resources and no backup configuration. The operator manages PostgreSQL clusters using the `postgresql` custom resource, providing automated failover, rolling updates, and connection pooling.

## When to Use

- You need to run PostgreSQL on Kubernetes with the Zalando operator
- Backups are not required or will be configured later
- Standard resource allocation is sufficient for the operator control plane

## Key Configuration Choices

- **Namespace** (`zalando-postgres-system`) -- dedicated namespace isolates the operator from managed PostgreSQL clusters
- **Create namespace** (`true`) -- namespace is created automatically if it does not exist
- **Resource requests** (`50m` CPU, `100Mi` memory) -- lightweight baseline for the operator pod
- **Resource limits** (`1000m` CPU, `1Gi` memory) -- headroom for reconciliation of multiple PostgreSQL clusters
- **No backup configuration** -- suitable for development or when backups are managed externally

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.

## Related Presets

- **02-with-r2-backup** -- Adds Cloudflare R2 backup configuration for production use
