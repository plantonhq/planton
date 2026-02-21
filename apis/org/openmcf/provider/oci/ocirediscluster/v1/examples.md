# OCI Redis Cluster — Examples

Complete YAML manifests for common OCI Redis Cluster deployment patterns.

## Table of Contents

- [Minimal Non-Sharded Cluster](#minimal-non-sharded-cluster)
- [Non-Sharded HA Cluster with NSGs](#non-sharded-ha-cluster-with-nsgs)
- [Sharded Production Cluster](#sharded-production-cluster)
- [Foreign Key References](#foreign-key-references)
- [Custom Config Set](#custom-config-set)

---

## Minimal Non-Sharded Cluster

A 2-node non-sharded cluster (1 primary + 1 replica) for development or testing. Uses the smallest practical memory allocation.

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

**What this creates:** A non-sharded Redis cluster with 2 GB per node. OCI defaults to non-sharded mode when `clusterMode` is not specified.

---

## Non-Sharded HA Cluster with NSGs

A 3-node non-sharded cluster (1 primary + 2 replicas) with network security group restrictions. Suitable for staging or production session stores.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciRedisCluster
metadata:
  name: session-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciRedisCluster.session-cache
  env: staging
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  displayName: "Session Cache"
  nodeCount: 3
  nodeMemoryInGbs: 8
  softwareVersion: "V7.0.5"
  clusterMode: nonsharded
  nsgIds:
    - value: "ocid1.networksecuritygroup.oc1..example-app"
    - value: "ocid1.networksecuritygroup.oc1..example-mgmt"
```

**What this creates:** A non-sharded cluster with 3 nodes and 8 GB each. Two NSGs restrict access — one for application traffic, one for management.

---

## Sharded Production Cluster

A sharded cluster with 3 shards and 3 nodes per shard (9 nodes total) for high-throughput production workloads. Data is distributed across shards for horizontal scaling.

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

**What this creates:** A sharded cluster with 3 shards x 3 nodes = 9 total nodes. Each node has 16 GB memory, giving 48 GB usable cache per shard (144 GB total across all shards before replication overhead). Clients connect via `discoveryFqdn` to discover shard topology.

---

## Foreign Key References

Reference OpenMCF-managed compartment, subnet, and NSG resources instead of hardcoding OCIDs. The `valueFrom` syntax resolves OCIDs from the referenced resource's stack outputs at deploy time.

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

**What this creates:** The same cluster as a literal-OCID example, but compartment, subnet, and NSG are resolved from other OpenMCF resources. This enables full infrastructure-as-code graphs without copy-pasting OCIDs.

---

## Custom Config Set

A non-sharded cluster with a custom OCI Cache Config Set for tuned Redis parameters. The config set is managed separately and referenced by OCID.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciRedisCluster
metadata:
  name: tuned-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciRedisCluster.tuned-cache
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  displayName: "Tuned Cache"
  nodeCount: 3
  nodeMemoryInGbs: 32
  softwareVersion: "V7.1.1"
  configSetId:
    value: "ocid1.ocicacheconfigset.oc1..example"
  nsgIds:
    - value: "ocid1.networksecuritygroup.oc1..example"
```

**What this creates:** A 3-node, 32 GB-per-node non-sharded cluster using a custom config set for Redis parameters like `maxmemory-policy` or `timeout`.

---

## Common Operations

### Scaling Node Memory

Update `nodeMemoryInGbs` in the manifest and re-apply. This is an in-place update — no recreation.

### Adding Replicas

Increase `nodeCount` and re-apply. For non-sharded clusters, additional nodes become replicas. For sharded clusters, each shard gains nodes.

### Adding Shards

Increase `shardCount` on a sharded cluster and re-apply. New shards are added and data is rebalanced.

### Upgrading Redis Version

Update `softwareVersion` to a newer version (e.g. `V7.0.5` to `V7.1.1`) and re-apply.

## Best Practices

- Start with non-sharded mode for workloads under 100 GB. Switch to sharded only when you need horizontal throughput scaling.
- Use `nodeCount` >= 2 for non-sharded clusters to get automatic failover.
- Place clusters in private subnets and use `nsgIds` to restrict access to application subnets only.
- Use `valueFrom` references instead of literal OCIDs to keep manifests portable across environments.
- Set `displayName` explicitly in production to make clusters identifiable in the OCI Console.
