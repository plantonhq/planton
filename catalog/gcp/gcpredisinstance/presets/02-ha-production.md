# Production HA Cache

This preset deploys a STANDARD_HA Redis instance — a primary with an automatic-failover replica in a second zone (99.9% SLA) — hardened with AUTH, TLS, RDB persistence, and a pinned maintenance window.

## When to Use

- Session stores, queues, and caches whose unavailability is a production incident
- Workloads that need the data to survive a node failure (automatic failover) and a full restart (RDB snapshots)
- Environments with compliance requirements for encryption in transit

## Key Configuration Choices

- **STANDARD_HA with both zones pinned** — `locationId` + `alternativeLocationId` keep the primary and replica next to zonal workloads; failover flips `current_location_id` (watch the output)
- **AUTH + TLS** — clients present the `auth_string` output and must trust the `server_ca_certs` output's CA chain
- **RDB every 12 hours, anchored at 03:00 UTC** — `rdbSnapshotStartTime` places snapshot I/O in the quiet window instead of wherever creation time fell
- **Maintenance window to the minute** — Sunday 03:30 UTC, coordinated after the snapshot, with a `description` recording why the window sits there
- **`deletionProtection: true`** — destroying the session store is a deliberate two-step (flip to false, apply, destroy)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-prod-vpc` | Your `GcpVpcNetwork` resource name | Your VPC manifest |

## Related Presets

- **01-basic-cache** — when losing the cache on restart is acceptable
- **03-private-services-access** — Shared VPC / PSA connectivity with read replicas and CMEK

## Related Kinds

- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network the cache attaches to
