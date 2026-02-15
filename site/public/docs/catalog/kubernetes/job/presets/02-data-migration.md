---
title: "Data Migration Job"
description: "This preset creates a Kubernetes Job for database schema migrations or data migrations. Configured with a 1-hour deadline, higher resource limits, and minimal retries to prevent duplicate migration..."
type: "preset"
rank: "02"
presetSlug: "02-data-migration"
componentSlug: "job"
componentTitle: "Job"
provider: "kubernetes"
icon: "package"
order: 2
---

# Data Migration Job

This preset creates a Kubernetes Job for database schema migrations or data migrations. Configured with a 1-hour deadline, higher resource limits, and minimal retries to prevent duplicate migration runs.

## When to Use

- Database schema migrations (e.g., Flyway, Alembic, Rails migrations)
- One-time data transformation or migration jobs
- Jobs that should not retry aggressively to avoid duplicate writes

## Key Configuration Choices

- **Minimal retries** (`backoffLimit: 1`) -- migration jobs should not blindly retry; investigate failures before re-running
- **1-hour deadline** (`activeDeadlineSeconds: 3600`) -- prevents runaway migrations from consuming resources indefinitely
- **Higher resources** (`2Gi` memory, `2000m` CPU limits) -- migrations often process large datasets; adjust based on data volume
- **No TTL cleanup** -- migration job pods are preserved for manual inspection; delete manually after verification

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-migration-image>` | Container image with migration tooling | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline output |
| `<your-database-connection-string>` | Database connection URL | Your database configuration or secrets |
| `<your-migration-command>` | Migration command (e.g., `flyway migrate`, `alembic upgrade head`) | Your migration framework docs |

## Related Presets

- **01-batch-processing** -- General-purpose batch job with auto-cleanup
