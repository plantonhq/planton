# GcpRedisClusterEndpointSet — Terraform Implementation

This directory contains the Terraform implementation for registering
consumer-built Private Service Connect connections on a Memorystore for
Redis Cluster from the Planton spec: one
`google_redis_cluster_user_created_connections`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the bare cluster name derived from the cluster reference, the connection count |
| `main.tf` | `google_redis_cluster_user_created_connections` |
| `outputs.tf` | `cluster_name`, `endpoint_count`, `connection_count`, `region` |

## Send Posture

- **`name`** -- the bare cluster name Google's resource is keyed by,
  derived as the last path segment of `spec.cluster` (a `GcpRedisCluster`
  reference resolves to the full resource path; a bare literal passes
  through unchanged) -- PARITY with the Pulumi module.
- **`cluster_endpoints[].connections[].psc_connection`** -- the five
  identifiers arrive as flattened references (the rule's self link and PSC
  connection id, the address, the network path, the attachment handle) and
  are sent verbatim; `project_id` is Optional+Computed and sent only when
  set, so Google records the rule's own project otherwise.
- **`deletion_policy`** -- DELETE (default), PREVENT, or ABANDON; sent only
  when set.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
