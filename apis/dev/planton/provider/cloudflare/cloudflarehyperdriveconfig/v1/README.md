# CloudflareHyperdriveConfig

Provision a Cloudflare Hyperdrive: a connection pooler and global query cache
that lets a Worker reach a regional SQL database (PostgreSQL or MySQL) with low
latency, without paying the full connection-setup cost on every request.

## Why Hyperdrive

A Worker runs at the edge, but a traditional SQL database lives in one region.
Opening a fresh connection from the edge to that database on every request is
slow and can exhaust the database's connection slots. Hyperdrive sits in front of
the origin and:

- **Pools connections** so many Worker invocations share a small set of
  long-lived origin connections.
- **Caches read queries** at the edge (configurable freshness) to cut latency and
  origin load.

A Worker uses Hyperdrive through a `hyperdrive` binding; it then talks to the
database with an ordinary driver as if it were local.

## Quick start

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareHyperdriveConfig
metadata:
  name: app-prod-pg
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: app-prod-pg
  origin:
    database: app_production
    scheme: postgres
    user: app_user
    host: db.example.com
    port: 5432
    password:
      value: <managed-secret-reference>
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `name` | yes | Human-readable config name |
| `origin.database` | yes | Origin database name |
| `origin.scheme` | yes | `postgres`, `postgresql`, or `mysql` |
| `origin.user` | yes | Database user |
| `origin.password` | yes | Database password (managed secret, JIT-resolved) |
| `origin.host` | no | Hostname/IP reachable from Cloudflare |
| `origin.port` | no | Port (defaults to the engine default) |
| `origin.accessClientId` / `accessClientSecret` | no | Cloudflare Access service-token credentials for Access-fronted origins |
| `origin.serviceId` | no | Workers VPC Service to egress through for private origins (mutually exclusive with `mtls`) |
| `caching.disabled` | no | Disable query caching (default: caching on) |
| `caching.maxAge` | no | Max cache age in seconds (default 60) |
| `caching.staleWhileRevalidate` | no | Stale-serve window in seconds (default 15) |
| `mtls.caCertificateId` / `mtlsCertificateId` / `sslmode` | no | Mutual-TLS configuration |
| `originConnectionLimit` | no | Max pooled connections (5–100; 0 = plan default) |

## Outputs

| Output | Description |
|---|---|
| `hyperdrive_id` | The Hyperdrive config ID (referenced by a Worker `hyperdrive` binding) |
| `name` | The config name |

## Security

`origin.password` and `origin.accessClientSecret` are secret-by-default: provide
them as managed-secret references. They are resolved just-in-time at deploy and
never stored in plaintext. The Cloudflare API treats both as write-only and never
returns them.

## Related components

- `CloudflareWorker` — binds this config via `hyperdrive_configs`.
- `CloudflareD1Database` — Cloudflare-native serverless SQL (no Hyperdrive needed).
