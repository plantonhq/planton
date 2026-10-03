# PostgreSQL Read Replica

This preset attaches a read replica to an existing PostgreSQL primary. A replica is a full `GcpCloudSql` node of its own — same kind, own manifest — that references its primary through `masterInstanceName`, so read capacity scales by adding nodes to the resource graph rather than editing the primary.

## When to Use

- Offloading read-heavy traffic (reports, analytics, dashboards) from the primary
- Serving reads closer to users (a replica may live in a different region — set `region` accordingly)
- Preparing a promotion path: a replica can later be promoted to a standalone primary (an operational action outside IaC)

## Prerequisites

- The primary must have automated backups enabled (this component validates that at the primary; the API enforces it at replica creation)
- For a private-IP replica: the same VPC private-services-access prerequisites as any private instance

## Key Configuration Choices

- **`masterInstanceName` by reference** — resolves the primary's `instance_name` output; immutable (a replica cannot re-parent)
- **No backup block** — replicas do not take their own automated backups; recovery flows through the primary
- **Independent tier** — replicas may be sized differently than the primary (bigger for analytics, smaller for low-traffic reads)
- **Same `databaseVersion` as the primary** — required by the API

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-postgres-prod` | Your primary's resource name | The primary's manifest |
| `my-prod-vpc` | Your `GcpVpcNetwork` resource name | Your VPC manifest |

## Related Presets

- **01-postgres-production-private** — the primary this replica pairs with

## Related Kinds

- [GcpCloudSqlUser](/docs/catalog/gcp/gcpcloudsqluser) — users are instance-scoped; replicas inherit users from the primary
