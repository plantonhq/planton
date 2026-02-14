# Minimal Harbor Container Registry

This preset deploys Harbor with default settings and ingress access. Harbor is a cloud-native container registry with vulnerability scanning, content trust, replication, and RBAC.

## When to Use

- You need a self-hosted container registry on Kubernetes
- Development or testing environments where default storage (filesystem) is sufficient
- Evaluating Harbor before configuring production-grade external storage

## Key Configuration Choices

- **Ingress enabled** -- exposes the Harbor web UI and Docker registry API at the specified hostname
- **Default storage** -- uses local filesystem storage; suitable for development but not production (data lost on pod restart without PVCs)
- **Default database and cache** -- self-managed PostgreSQL and Redis; no external dependencies required

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-harbor.example.com>` | Hostname for the Harbor registry and UI | Your DNS provider |

## Related Presets

- **02-production-with-s3** -- External S3 storage, higher resources for production use
