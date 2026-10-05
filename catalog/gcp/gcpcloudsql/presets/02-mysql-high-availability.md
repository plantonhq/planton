# High-Availability MySQL (Auth Proxy Access)

This preset deploys a MySQL 8.0 instance with regional high availability and binary logging, exposed through a public IP that has **zero authorized networks** — so the only way in is the IAM-authenticated Cloud SQL Auth Proxy or a Cloud SQL connector.

## When to Use

- Production MySQL without VPC plumbing — the Auth Proxy pattern needs no private services access setup
- Teams that connect from many places (CI, laptops, multiple clusters) and want IAM as the perimeter instead of IP allowlists
- Any MySQL workload that needs automatic zonal failover

## Key Configuration Choices

- **Public IP, empty allowlist** — `ipv4Enabled: true` with no `authorizedNetworks` means direct connections are impossible; all access flows through the Auth Proxy / connectors with IAM authentication and built-in TLS
- **`binaryLogEnabled`** — MySQL's mechanism for point-in-time recovery, and required for REGIONAL availability and read replicas (validated pre-deploy)
- **REGIONAL availability** — standby in a second zone with automatic failover
- **`sslMode: ENCRYPTED_ONLY`** — any direct connection attempt must at least be TLS

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `rootPassword` | Initial root password | Generate one; rotate after first deploy |

## Related Presets

- **01-postgres-production-private** — the private-IP pattern when workloads live inside one VPC
- **03-postgres-read-replica** — the read-scaling pattern (works for MySQL too, thanks to the binary logs enabled here)

## Related Kinds

- [GcpCloudSqlDatabase](/docs/catalog/gcp/gcpcloudsqldatabase) — create application databases on this instance
- [GcpCloudSqlUser](/docs/catalog/gcp/gcpcloudsqluser) — per-application users; IAM-type users pair well with the Auth Proxy pattern
