# Production PostgreSQL (Private IP)

This preset deploys a production-grade PostgreSQL 16 instance reachable only over private IP inside your VPC — no public address at all — with regional high availability, point-in-time recovery, Query Insights, and both delete guards engaged.

## When to Use

- The primary relational database for a production service running inside your VPC (GKE, Cloud Run with VPC egress, GCE)
- Any environment where the database must never be reachable from the public internet
- Workloads that need automatic zonal failover (99.95% SLA on the ENTERPRISE edition)

## Prerequisites

Private IP requires **private services access** on the VPC before the instance can be created — GCP rejects the create otherwise:

1. A `GcpGlobalAddress` with `addressType: INTERNAL` and `purpose: VPC_PEERING` (the reserved range)
2. A `GcpServiceNetworkingConnection` on the VPC listing that range

## Key Configuration Choices

- **Private network by reference** — `network.privateNetwork` resolves the VPC's `network_id` output; setting it is what enables private IP
- **No `ipv4Enabled`** — the instance gets no public address; connect via the private IP or the Cloud SQL Auth Proxy
- **REGIONAL availability** with automated backups (required pairing, validated pre-deploy)
- **Point-in-time recovery** with 7 days of transaction logs — restore to any second after a bad migration or accidental delete
- **Both delete guards** — `deletionProtection` (IaC-side) and `deletionProtectionEnabled` (API-side, blocks console/gcloud deletion too)
- **`retainBackupsOnDelete`** — backups survive even if the instance is deleted

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-prod-vpc` | Your `GcpVpcNetwork` resource name | Your VPC manifest |
| `rootPassword` | Initial postgres superuser password | Generate one; rotate after first deploy |

## Related Presets

- **02-mysql-high-availability** — the MySQL equivalent (binary logs instead of PITR)
- **03-postgres-read-replica** — scale reads by attaching a replica to this primary

## Related Kinds

- [GcpServiceNetworkingConnection](/docs/catalog/gcp/gcpservicenetworkingconnection) — the private services access peering this preset depends on
- [GcpCloudSqlDatabase](/docs/catalog/gcp/gcpcloudsqldatabase) — create application databases on this instance
- [GcpCloudSqlUser](/docs/catalog/gcp/gcpcloudsqluser) — create per-application users instead of sharing the admin user
