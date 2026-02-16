# GcpBigtableInstance Examples

Copy-paste ready YAML manifests for deploying Cloud Bigtable instances via OpenMCF.

---

## Example 1: Minimal Single Cluster

**When to use:** Development or testing. Single cluster with automatic node allocation in one zone.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: dev-bigtable
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  instanceName: dev-bigtable
  deletionProtection: false
  clusters:
    - clusterId: dev-cluster-c1
      zone: us-central1-a
```

---

## Example 2: Single Cluster with Fixed Nodes

**When to use:** Predictable workloads where you know the required capacity. Three fixed nodes with SSD storage.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: fixed-bigtable
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  instanceName: fixed-bigtable
  displayName: Fixed Capacity Bigtable
  clusters:
    - clusterId: fixed-cluster-c1
      zone: us-central1-a
      numNodes: 3
      storageType: SSD
```

---

## Example 3: Single Cluster with Autoscaling

**When to use:** Variable workloads that benefit from automatic scaling based on CPU and storage utilization.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: autoscale-bigtable
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  instanceName: autoscale-bigtable
  displayName: Autoscaling Bigtable
  clusters:
    - clusterId: autoscale-cluster-c1
      zone: us-central1-a
      autoscalingConfig:
        minNodes: 1
        maxNodes: 10
        cpuTarget: 60
        storageTarget: 2560
```

---

## Example 4: HDD Storage for Batch Analytics

**When to use:** Large batch-analytics workloads where latency is less critical and cost optimization is a priority.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: analytics-bigtable
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  instanceName: analytics-bigtable
  displayName: Batch Analytics Bigtable
  clusters:
    - clusterId: analytics-cluster-c1
      zone: us-central1-a
      numNodes: 3
      storageType: HDD
```

---

## Example 5: Multi-Cluster High Availability

**When to use:** Production workloads requiring automatic replication and failover across zones or regions.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: ha-bigtable
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  instanceName: ha-bigtable
  displayName: HA Bigtable Instance
  deletionProtection: true
  clusters:
    - clusterId: ha-cluster-uscentral
      zone: us-central1-a
      numNodes: 3
      storageType: SSD
    - clusterId: ha-cluster-useast
      zone: us-east1-b
      numNodes: 3
      storageType: SSD
    - clusterId: ha-cluster-europe
      zone: europe-west1-b
      numNodes: 3
      storageType: SSD
```

---

## Example 6: CMEK Encrypted Clusters

**When to use:** Compliance requirements (HIPAA, PCI-DSS, FedRAMP) that mandate customer-managed encryption keys.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: cmek-bigtable
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  instanceName: cmek-bigtable
  displayName: CMEK Encrypted Bigtable
  deletionProtection: true
  clusters:
    - clusterId: cmek-cluster-c1
      zone: us-central1-a
      numNodes: 3
      storageType: SSD
      kmsKeyName:
        value: projects/my-gcp-project/locations/us-central1/keyRings/bigtable-kr/cryptoKeys/cluster-key
    - clusterId: cmek-cluster-c2
      zone: us-east1-b
      numNodes: 3
      storageType: SSD
      kmsKeyName:
        value: projects/my-gcp-project/locations/us-east1/keyRings/bigtable-kr/cryptoKeys/cluster-key
```

---

## Example 7: Full-Featured Production

**When to use:** Maximum configuration with all features: multi-cluster HA, autoscaling, CMEK, 2X scaling factor, deletion protection, and force destroy.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigtableInstance
metadata:
  name: prod-bigtable
  labels:
    openmcf.org/provisioner: pulumi
spec:
  projectId:
    value: my-gcp-project
  instanceName: prod-bigtable
  displayName: Production Bigtable Instance
  deletionProtection: true
  forceDestroy: false
  clusters:
    - clusterId: prod-cluster-uscentral
      zone: us-central1-a
      storageType: SSD
      nodeScalingFactor: NodeScalingFactor2X
      kmsKeyName:
        value: projects/my-gcp-project/locations/us-central1/keyRings/bigtable-kr/cryptoKeys/cluster-key
      autoscalingConfig:
        minNodes: 2
        maxNodes: 20
        cpuTarget: 60
        storageTarget: 2560
    - clusterId: prod-cluster-useast
      zone: us-east1-b
      storageType: SSD
      nodeScalingFactor: NodeScalingFactor2X
      kmsKeyName:
        value: projects/my-gcp-project/locations/us-east1/keyRings/bigtable-kr/cryptoKeys/cluster-key
      autoscalingConfig:
        minNodes: 2
        maxNodes: 20
        cpuTarget: 60
        storageTarget: 2560
    - clusterId: prod-cluster-europe
      zone: europe-west1-b
      storageType: SSD
      nodeScalingFactor: NodeScalingFactor2X
      kmsKeyName:
        value: projects/my-gcp-project/locations/europe-west1/keyRings/bigtable-kr/cryptoKeys/cluster-key
      autoscalingConfig:
        minNodes: 2
        maxNodes: 20
        cpuTarget: 60
        storageTarget: 2560
```

---

## Deployment

```shell
openmcf apply -f <manifest>.yaml
```

For more details, see the [main README](README.md).
