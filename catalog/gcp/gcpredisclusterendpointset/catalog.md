# GCP Redis Cluster Endpoint Set

Registers the Private Service Connect connections you built by hand on a Memorystore for Redis Cluster. A cluster created without automatic connectivity publishes service attachments and nothing else; each consumer VPC reserves an address and creates a forwarding rule per attachment, and this set tells the cluster about every one of those rules so Google treats them as the cluster's endpoints. It is the way to reach one cluster from several VPCs or projects.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **User-created connections** — a `redis.ClusterUserCreatedConnections` registration on the cluster, replacing the cluster's whole user-created endpoint list with the manifest's

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** — an active connection in the Connect module with `roles/redis.admin` on the cluster's project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### The Chain

- **A `GcpRedisCluster`** created without `pscConfigs`, exposing its `*_service_attachment` outputs.
- **Per consumer VPC, per attachment:** a `GcpAddress` and a regional `GcpGlobalForwardingRule` (empty scheme, the attachment as `target`).

## Deploy

### Console

Open the deployment store, find **GCP Redis Cluster Endpoint Set**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **One Consumer Network** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRedisClusterEndpointSet
metadata:
  name: orders-cache-endpoints
  org: acme-corp
  env: prod
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

This registers one consumer network's two connections, to the cluster's discovery and primary attachments: Google rejects a group that leaves out an attachment the cluster publishes. A cluster with replicas publishes a reader attachment too, and its group adds a third connection on `reader_service_attachment`. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the whole chain — cluster, addresses, forwarding rules, this set — wires by reference in one InfraPipeline; the set is the last block and depends on every rule.

## Key Configuration

These are the most important decisions when configuring an endpoint set. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One connection per attachment, per network** — a cluster publishes a discovery attachment, a primary attachment, and (with replicas) a reader attachment; each `endpoints[]` entry is one VPC's group and must carry a connection for each. Reference the cluster's `*_service_attachment` output on each connection.

**The rule is named twice** — `forwardingRule` takes its `self_link`, `pscConnectionId` takes its `psc_connection_id`; a reference names one output path, so the same rule appears on both fields.

**The list is the set** — Google replaces the cluster's whole user-created endpoint list on every apply. Removing a connection here breaks its forwarding rule; delete both in one change.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | Reference Path |
|------------|-------|-------------------|
| **GCP Project** | `projectId`, `endpoints[].connections[].projectId` | `status.outputs.project_id` |
| **GCP Redis Cluster** | `cluster` | `status.outputs.name` |
| **GCP Redis Cluster** | `endpoints[].connections[].serviceAttachment` | `status.outputs.discovery_service_attachment`, `primary_service_attachment`, `reader_service_attachment` |
| **GCP Forwarding Rule** | `endpoints[].connections[].forwardingRule` | `status.outputs.self_link` |
| **GCP Forwarding Rule** | `endpoints[].connections[].pscConnectionId` | `status.outputs.psc_connection_id` |
| **GCP Address** | `endpoints[].connections[].address` | `status.outputs.address` |
| **GCP VPC Network** | `endpoints[].connections[].network` | `status.outputs.network_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can reference:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_name` | Bare cluster name the set is registered on | Audit |
| `endpoint_count` | Consumer networks registered | Dashboards |
| `connection_count` | Connections registered in all | Dashboards |
| `region` | The cluster's region | Rebuilding the cluster path |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One consumer network** — Discovery and primary connections from one VPC. Start from the **One Consumer Network** preset.

## Works With

- [**GCP Redis Cluster**](/infra-catalog/gcp-redis-cluster) — the cluster whose attachments the connections target
- [**GCP Forwarding Rule**](/infra-catalog/gcp-global-forwarding-rule) — the consumer endpoints (regional arm, empty scheme)
- [**GCP Address**](/infra-catalog/gcp-address) — the reserved internal addresses
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) — the consumer networks
