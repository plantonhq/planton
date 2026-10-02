# GCP Redis Cluster Endpoint Set

Registers the Private Service Connect connections a consumer built by hand on a Memorystore for Redis Cluster -- Google's "user-created connections". A cluster created without `pscConfigs` publishes service attachments and nothing else; each consumer VPC reserves an address and creates a regional forwarding rule per attachment, and this set tells the cluster about every one of those rules so Google treats them as the cluster's endpoints. It is how a cluster is reached from VPCs or projects its own connectivity automation cannot place endpoints in.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **User-created connections** -- a `redis_cluster_user_created_connections` registration on the cluster, replacing the cluster's whole user-created endpoint list with the manifest's

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/redis.admin` on the cluster's project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### The Chain

1. **A `GcpRedisCluster`** created without `pscConfigs`, so it publishes service attachments (`discovery_service_attachment`, `primary_service_attachment`, `reader_service_attachment` outputs).
2. **In each consumer VPC, per service attachment:** one `GcpAddress` (an internal address in a subnet of that VPC) and one regional `GcpGlobalForwardingRule` with an empty `loadBalancingScheme`, that address as `ipAddress`, and the attachment handle as `target`.
3. **This set**, naming every rule. Declare exactly one set per cluster.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRedisClusterEndpointSet
metadata:
  name: orders-cache-endpoints
spec:
  cluster:
    valueFrom:
      kind: GcpRedisCluster
      name: orders-cache
      fieldPath: status.outputs.name
  region: us-central1
  endpoints:
    - connections:
        - forwardingRule:
            valueFrom: {kind: GcpGlobalForwardingRule, name: orders-cache-disc, fieldPath: status.outputs.self_link}
          pscConnectionId:
            valueFrom: {kind: GcpGlobalForwardingRule, name: orders-cache-disc, fieldPath: status.outputs.psc_connection_id}
          address:
            valueFrom: {kind: GcpAddress, name: orders-cache-disc-ip, fieldPath: status.outputs.address}
          network:
            valueFrom: {kind: GcpVpcNetwork, name: consumer-vpc, fieldPath: status.outputs.network_id}
          serviceAttachment:
            valueFrom: {kind: GcpRedisCluster, name: orders-cache, fieldPath: status.outputs.discovery_service_attachment}
        - forwardingRule:
            valueFrom: {kind: GcpGlobalForwardingRule, name: orders-cache-prim, fieldPath: status.outputs.self_link}
          pscConnectionId:
            valueFrom: {kind: GcpGlobalForwardingRule, name: orders-cache-prim, fieldPath: status.outputs.psc_connection_id}
          address:
            valueFrom: {kind: GcpAddress, name: orders-cache-prim-ip, fieldPath: status.outputs.address}
          network:
            valueFrom: {kind: GcpVpcNetwork, name: consumer-vpc, fieldPath: status.outputs.network_id}
          serviceAttachment:
            valueFrom: {kind: GcpRedisCluster, name: orders-cache, fieldPath: status.outputs.primary_service_attachment}
```

```shell
planton apply -f redis-cluster-endpoint-set.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `cluster` | `StringValueOrRef` | The cluster (`GcpRedisCluster` reference to its `name` output, or the full path or bare name). |
| `region` | `string` | The cluster's region. |
| `endpoints` | `[]object` | One entry per consumer VPC, each with `connections[]` (one per cluster service attachment). |

Each connection requires `forwardingRule` (`GcpGlobalForwardingRule` `self_link`), `pscConnectionId` (the same rule's `psc_connection_id`), `address` (`GcpAddress` `address`), `network` (`GcpVpcNetwork` `network_id`), and `serviceAttachment` (a `GcpRedisCluster` attachment handle).

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The cluster's project. Immutable. |
| `endpoints[].connections[].projectId` | `StringValueOrRef` | the rule's project | Set only when the rule lives in a different project than the cluster. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- At least one endpoint, each with at least one connection; every connection carries all five identifying fields.
- Google enforces at apply that each endpoint carries exactly one connection per service attachment the cluster publishes.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `cluster_name` | `string` | Bare cluster name the set is registered on |
| `endpoint_count` | `int32` | Consumer networks registered |
| `connection_count` | `int32` | Connections registered in all |
| `region` | `string` | The cluster's region |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The list is the set.** Google replaces the cluster's whole user-created endpoint list on every apply; a connection missing from the manifest is deregistered and its forwarding rule stops working -- delete the rule in the same change.
- **One set per cluster.** Two sets on one cluster overwrite each other.
- **Every field is an output of another block** -- the forwarding rule twice (two output paths), the address, the network, the cluster's attachment handle -- so nothing is copied by hand.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpRedisCluster** -- the cluster, created without `pscConfigs`
- **GcpGlobalForwardingRule** -- the consumer endpoint (regional, empty scheme, targeting an attachment)
- **GcpAddress** -- the reserved internal address each rule serves on
- **GcpVpcNetwork** -- the consumer network

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
