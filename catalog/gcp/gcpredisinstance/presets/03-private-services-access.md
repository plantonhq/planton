# Private Services Access with Read Replicas

This preset deploys a STANDARD_HA Redis instance over the VPC's private services access connection — the connectivity mode Shared VPC requires — with a read endpoint for scale-out, CMEK encryption at rest, and an operator-allocated address range.

## When to Use

- Shared VPC topologies, where direct peering is not available
- Networks that centralize producer address planning on one private services access connection (the same one Cloud SQL private IP uses)
- Read-heavy workloads (feeds, leaderboards, content caches) that want a separate read endpoint
- Compliance regimes that require customer-managed encryption keys

## Prerequisites

Private services access must exist on the VPC before the instance is created — GCP rejects the create otherwise:

1. A `GcpGlobalAddress` with `addressType: INTERNAL` and `purpose: VPC_PEERING` (the reserved range named in `reservedIpRange`)
2. A `GcpServiceNetworkingConnection` on the VPC listing that range

## Key Configuration Choices

- **`connectMode: PRIVATE_SERVICE_ACCESS`** with `reservedIpRange` naming your allocated range — the instance consumes address space you planned, not an arbitrary /29
- **Read replicas enabled at creation** — `readReplicasMode` is immutable; the `read_endpoint` output serves read-only traffic
- **CMEK by reference** — `customerManagedKey` resolves the `GcpKmsKey` node's key id; grant the `persistence_iam_identity` output access for import/export
- **AUTH + TLS** — the same client hardening as the HA preset
- **`deletionProtection: true`** — destroying a production cache is a deliberate two-step (flip to false, apply, destroy)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-prod-vpc` | Your `GcpVpcNetwork` resource name | Your VPC manifest |
| `psa-range-services` | The allocated range's `GcpGlobalAddress` name | Your PSA manifests |
| `my-cache-cmek` | Your `GcpKmsKey` resource name | Your KMS manifests |

## Related Presets

- **02-ha-production** — the same hardening over direct peering

## Related Kinds

- [GcpServiceNetworkingConnection](/docs/catalog/gcp/gcpservicenetworkingconnection) — the private services access peering this preset depends on
- [GcpGlobalAddress](/docs/catalog/gcp/gcpglobaladdress) — the reserved range the instance consumes
- [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) — the customer-managed encryption key
