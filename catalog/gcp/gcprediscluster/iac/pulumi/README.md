# GcpRedisCluster — Pulumi Implementation

This directory contains the Pulumi implementation for a Memorystore for
Redis Cluster from the Planton spec: one `gcp.redis.Cluster` with the two
project APIs it depends on enabled.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `cluster` |
| `module/locals.go` | Cluster name (spec or metadata.name fallback), the platform labels |
| `module/cluster.go` | Enables `redis.googleapis.com` and `networkconnectivity.googleapis.com`; maps spec to `gcp.redis.Cluster`; exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`AuthorizationMode`, `TransitEncryptionMode`** -- always sent, with
  Google's defaults when the spec leaves them unset; both are immutable, so
  the posture never depends on a provider default.
- **`DeletionProtectionEnabled`** -- always sent (spec default `true`).
- **`ReplicaCount`** -- always sent; 0 is an explicit "no replicas".
- **`NodeType`, `ServerCaMode`, `ServerCaPool`, `KmsKey`,
  `MaintenanceVersion`, `AclPolicy`, `RedisConfigs`, every persistence and
  zone-distribution leaf** -- sent only when set (Optional+Computed on the
  provider).
- **`PscConfigs`** -- the network as the VPC's relative resource path.
- **Maintenance window** -- `Day` and `StartTime.Hours` only; Google
  exposes the window start by the hour.
- **`DeletionPolicy`** -- sent only when set.

## Outputs

The discovery endpoint is read from `DiscoveryEndpoints` (empty without
`PscConfigs`). The three `*_service_attachment` handles are picked from
`PscServiceAttachments` by `ConnectionType`, the shape the Terraform module
exports, so a consumer forwarding rule and a `GcpRedisClusterEndpointSet`
can reference the one they need.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
