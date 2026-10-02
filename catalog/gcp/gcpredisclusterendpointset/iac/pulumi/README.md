# GcpRedisClusterEndpointSet — Pulumi Implementation

This directory contains the Pulumi implementation for registering
consumer-built Private Service Connect connections on a Memorystore for
Redis Cluster from the Planton spec: one
`gcp.redis.ClusterUserCreatedConnections`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `endpoint_set` |
| `module/locals.go` | The bare cluster name derived from the cluster reference |
| `module/endpoint_set.go` | Maps spec to `gcp.redis.ClusterUserCreatedConnections`; exports the outputs |
| `module/outputs.go` | Output key constants (`cluster_name`, `endpoint_count`, `connection_count`, `region`) |

## Send Posture (parity with Terraform)

- **`Name`** -- the bare cluster name, the last path segment of
  `spec.cluster` (a full resource path from a `GcpRedisCluster` reference,
  or a bare literal).
- **`ClusterEndpoints[].Connections[].PscConnection`** -- the five
  identifiers sent verbatim from the resolved references; `ProjectId` sent
  only when set.
- **`DeletionPolicy`** -- sent only when set.
- **`endpoint_count`, `connection_count`** -- the declared counts, the
  shape the Terraform module exports.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
