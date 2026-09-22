# GcpRedisCluster — Terraform Implementation

This directory contains the Terraform implementation for a Memorystore for
Redis Cluster from the Planton spec: one `google_redis_cluster` with the
two project APIs it depends on enabled.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, cluster name (spec or metadata.name fallback), the two explicit security-mode defaults, null-when-empty levers, the platform labels, and the per-connection-type service attachment map |
| `main.tf` | `google_project_service` x2, `google_redis_cluster` |
| `outputs.tf` | `name`, `uid`, `state`, the discovery endpoint, the three service attachment handles, `size_gb`, `shard_count`, `replica_count`, `backup_collection` |

## Send Posture

- **`authorization_mode`, `transit_encryption_mode`** -- always sent, with
  Google's defaults (`AUTH_MODE_DISABLED`,
  `TRANSIT_ENCRYPTION_MODE_DISABLED`) when the spec leaves them unset; both
  are immutable, so the posture must never depend on a provider default
  (PARITY with the Pulumi module).
- **`deletion_protection_enabled`** -- always sent (spec default `true`).
- **`replica_count`** -- always sent; 0 is an explicit "no replicas".
- **`node_type`, `server_ca_mode`, `server_ca_pool`, `kms_key`,
  `maintenance_version`, `acl_policy`, `redis_configs`, every persistence
  and zone-distribution leaf** -- Optional+Computed on the provider; sent
  only when set so Google's own defaults are never fought.
- **`psc_configs`** -- the network as the VPC's relative resource path (the
  `GcpVpcNetwork` `network_id` output), the only form the Service
  Connectivity API accepts; empty means no Google-placed endpoints.
- **Maintenance window** -- `day` and `start_time.hours` only. Google
  exposes the Redis Cluster window start by the hour; the TimeOfDay's
  finer fields are never sent.
- **`deletion_policy`** -- DELETE (default), PREVENT, or ABANDON; sent only
  when set.

## Outputs

`discovery_endpoint_address` / `_port` come from `discovery_endpoints[0]`
(empty / 0 without `psc_configs`). The three `*_service_attachment` handles
are picked from `psc_service_attachments` by `connection_type` in
`locals.tf`, so a consumer forwarding rule and a `GcpRedisClusterEndpointSet`
can reference the one they need -- a reference cannot index a list.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
