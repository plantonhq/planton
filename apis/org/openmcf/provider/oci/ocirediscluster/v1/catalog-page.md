# OCI Redis Cluster

Deploys an Oracle Cloud Infrastructure Cache (Redis) cluster — a fully managed, Redis-compatible in-memory caching service that supports both non-sharded (single primary with replicas) and sharded (horizontally scaled) topologies.

## What Gets Created

When you deploy an OciRedisCluster resource, OpenMCF provisions:

- **Redis Cluster** — an `oci_redis_redis_cluster` resource in the specified compartment and subnet. Non-sharded clusters consist of a single primary with optional replicas. Sharded clusters distribute data across multiple shards, each with its own primary and replicas. Freeform tags are automatically populated from metadata labels, organization, and environment.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the Redis cluster will be created — either a literal value or a reference to an OciCompartment resource
- **A subnet OCID** where the Redis cluster will be placed — either a literal value or a reference to an OciSubnet resource
- **A supported Redis version** available in the target region (e.g. `V7.0.5`, `V7.1.1`)

## Quick Start

Create a file `redis-cluster.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciRedisCluster
metadata:
  name: my-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciRedisCluster.my-cache
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  nodeCount: 2
  nodeMemoryInGbs: 4
  softwareVersion: "V7.0.5"
```

Deploy:

```shell
openmcf apply -f redis-cluster.yaml
```

This creates a non-sharded Redis cluster with 2 nodes (1 primary + 1 replica), each with 4 GB of memory. The cluster ID, primary endpoint, replicas endpoint, and discovery endpoint are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the Redis cluster will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `subnetId` | `StringValueOrRef` | OCID of the subnet where the Redis cluster will be placed. Changing this forces recreation. Can reference an OciSubnet resource via `valueFrom`. | Required |
| `nodeCount` | `int32` | Number of nodes. For non-sharded clusters: total node count (1 primary + N-1 replicas). For sharded clusters: nodes per shard. | >= 1 |
| `nodeMemoryInGbs` | `float` | Memory allocated to each node in gigabytes. Common values: 2, 4, 8, 16, 32. | > 0 |
| `softwareVersion` | `string` | OCI Cache engine version (e.g. `V7.0.5`, `V7.1.1`). Available versions depend on the region. | Non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name shown in the OCI Console. Falls back to `metadata.name` if not provided. |
| `clusterMode` | `ClusterMode` | `cluster_mode_unspecified` | Cluster topology. `nonsharded`: single primary with replicas. `sharded`: multiple shards for horizontal scaling. When unset, OCI defaults to non-sharded. Changing this forces recreation. |
| `shardCount` | `int32` | — | Number of shards. Only applicable when `clusterMode` is `sharded`. Each shard gets `nodeCount` nodes. Must be > 0 when `clusterMode` is `sharded`. |
| `nsgIds` | `StringValueOrRef[]` | — | OCIDs of network security groups controlling access to the cluster. Can reference OciSecurityGroup resources via `valueFrom`. |
| `configSetId` | `StringValueOrRef` | — | OCID of an OCI Cache Config Set providing custom Redis configuration parameters (e.g. maxmemory-policy, timeout). When omitted, the default configuration is used. |

## Examples

### Non-Sharded Cluster for Development

A minimal non-sharded cluster with a single replica for development workloads:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciRedisCluster
metadata:
  name: dev-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciRedisCluster.dev-cache
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  nodeCount: 2
  nodeMemoryInGbs: 2
  softwareVersion: "V7.0.5"
```

### Sharded Cluster for Production

A sharded cluster with 3 shards and 3 nodes per shard for horizontally scaled production workloads:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciRedisCluster
metadata:
  name: prod-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciRedisCluster.prod-cache
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  displayName: "Production Redis Cache"
  nodeCount: 3
  nodeMemoryInGbs: 16
  softwareVersion: "V7.1.1"
  clusterMode: sharded
  shardCount: 3
  nsgIds:
    - value: "ocid1.networksecuritygroup.oc1..example"
```

### Cluster with Foreign Key References

Reference OpenMCF-managed compartment, subnet, and NSG resources instead of hardcoding OCIDs:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciRedisCluster
metadata:
  name: ref-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciRedisCluster.ref-cache
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: staging-compartment
      fieldPath: status.outputs.compartmentId
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: app-subnet
      fieldPath: status.outputs.subnetId
  nodeCount: 2
  nodeMemoryInGbs: 8
  softwareVersion: "V7.0.5"
  nsgIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: cache-nsg
        fieldPath: status.outputs.networkSecurityGroupId
```

### Cluster with Custom Config Set

A non-sharded cluster using a custom OCI Cache Config Set for tuned Redis configuration:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciRedisCluster
metadata:
  name: custom-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciRedisCluster.custom-cache
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  nodeCount: 3
  nodeMemoryInGbs: 32
  softwareVersion: "V7.1.1"
  configSetId:
    value: "ocid1.ocicacheconfigset.oc1..example"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `clusterId` | `string` | OCID of the Redis cluster |
| `primaryFqdn` | `string` | FQDN of the primary (read-write) endpoint. Main connection point for non-sharded clusters. |
| `primaryEndpointIpAddress` | `string` | Private IP address of the primary endpoint |
| `replicasFqdn` | `string` | FQDN of the replica (read-only) endpoint |
| `discoveryFqdn` | `string` | FQDN of the discovery endpoint for sharded clusters. Clients use this to discover shard topology. |

## Related Components

- [OciSubnet](/docs/catalog/oci/ocisubnet) — provides the subnet referenced by `subnetId`
- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId`
- [OciSecurityGroup](/docs/catalog/oci/ocisecuritygroup) — controls network access to the cluster via `nsgIds`
- [OciVcn](/docs/catalog/oci/ocivcn) — provides the VCN that contains the subnet
