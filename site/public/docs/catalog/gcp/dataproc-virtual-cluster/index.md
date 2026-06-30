---
title: "Dataproc Virtual Cluster"
description: "Dataproc Virtual Cluster deployment documentation"
icon: "package"
order: 100
componentName: "gcpdataprocvirtualcluster"
---

# GCP Dataproc Virtual Cluster

Deploys a Dataproc on GKE virtual cluster that schedules Spark, PySpark, and SparkR workloads as Kubernetes pods on an existing GKE cluster. Instead of managing dedicated Compute Engine VMs, the virtual cluster shares GKE infrastructure with other workloads.

## What Gets Created

When you deploy a GcpDataprocVirtualCluster resource, Planton provisions:

- **Dataproc Cluster** — a `google_dataproc_cluster` resource with `virtual_cluster_config` pointing to the specified GKE cluster and namespace
- **Node Pool Target Bindings** — one or more GKE node pool assignments with Dataproc roles (DEFAULT, CONTROLLER, SPARK_DRIVER, SPARK_EXECUTOR) controlling where workloads are scheduled
- **Auxiliary Services** — created only when `auxiliaryServicesConfig` is specified, integrates an existing Dataproc Metastore and/or Spark History Server with the virtual cluster

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** with the Dataproc API enabled (`dataproc.googleapis.com`)
- **A GKE cluster** in the same project and region, referenced via `gkeClusterTarget`
- **At least one GKE node pool** assigned the DEFAULT role
- **A Kubernetes namespace** (optional — Dataproc creates one automatically if not specified)
- **A GCS bucket** if specifying a custom staging bucket

## Quick Start

Create a file `dataproc-virtual-cluster.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: my-spark-on-gke
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpDataprocVirtualCluster.my-spark-on-gke
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  gkeClusterTarget:
    value: projects/my-gcp-project/locations/us-central1/clusters/my-gke-cluster
  softwareConfig:
    componentVersion:
      SPARK: "3.5"
  nodePoolTargets:
    - nodePool:
        value: projects/my-gcp-project/locations/us-central1/clusters/my-gke-cluster/nodePools/default-pool
      roles:
        - DEFAULT
```

Deploy:

```shell
planton apply -f dataproc-virtual-cluster.yaml
```

This creates a Dataproc virtual cluster on an existing GKE cluster, scheduling Spark workloads on the default node pool.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the virtual cluster is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `region` | `string` | GCP region. Must match the region of the target GKE cluster. | Required |
| `gkeClusterTarget` | `StringValueOrRef` | Fully qualified GKE cluster resource ID. Format: `projects/{project}/locations/{location}/clusters/{name}`. Can reference a GcpGkeCluster resource via `valueFrom`. | Required |
| `softwareConfig.componentVersion` | `map<string, string>` | Component versions. The `SPARK` key is mandatory (e.g., `{"SPARK": "3.5"}`). | Required |
| `nodePoolTargets` | `object[]` | GKE node pool assignments with Dataproc roles. At least one must have the DEFAULT role. | Minimum 1 item |
| `nodePoolTargets[].nodePool` | `StringValueOrRef` | GKE node pool reference (short name or fully qualified path). Can reference a GcpGkeNodePool resource via `valueFrom`. | Required |
| `nodePoolTargets[].roles` | `string[]` | Dataproc roles for this node pool. Valid values: `DEFAULT`, `CONTROLLER`, `SPARK_DRIVER`, `SPARK_EXECUTOR`. | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `clusterName` | `string` | `metadata.name` | Explicit Dataproc cluster name. Lowercase letters, numbers, hyphens; starts with a letter. |
| `kubernetesNamespace` | `StringValueOrRef` | Auto-created | Kubernetes namespace for the virtual cluster. Can reference a KubernetesNamespace resource via `valueFrom`. |
| `stagingBucket` | `StringValueOrRef` | Default bucket | GCS bucket for staging job dependencies. Can reference a GcpGcsBucket resource via `valueFrom`. |
| `softwareConfig.properties` | `map<string, string>` | `{}` | Daemon config properties in `prefix:property` format (e.g., `{"spark:spark.kubernetes.container.image": "custom:latest"}`). |
| `nodePoolTargets[].nodePoolConfig.locations` | `string[]` | — | Compute Engine zones for node pool nodes. |
| `nodePoolTargets[].nodePoolConfig.machineType` | `string` | — | Machine type for nodes (e.g., `n1-standard-4`). |
| `nodePoolTargets[].nodePoolConfig.localSsdCount` | `int` | `0` | Local SSD disks per node. |
| `nodePoolTargets[].nodePoolConfig.minCpuPlatform` | `string` | — | Minimum CPU platform (e.g., `Intel Haswell`). |
| `nodePoolTargets[].nodePoolConfig.preemptible` | `bool` | `false` | Use preemptible VMs. Cannot be used with CONTROLLER or sole DEFAULT role. |
| `nodePoolTargets[].nodePoolConfig.spot` | `bool` | `false` | Use Spot VMs. Same restrictions as preemptible. |
| `nodePoolTargets[].nodePoolConfig.autoscaling.minNodeCount` | `int` | — | Minimum nodes. Must be >= 0. |
| `nodePoolTargets[].nodePoolConfig.autoscaling.maxNodeCount` | `int` | — | Maximum nodes. Must be >= `minNodeCount`. |
| `auxiliaryServicesConfig.metastoreService` | `string` | — | Fully qualified Dataproc Metastore service name for Hive metastore integration. |
| `auxiliaryServicesConfig.sparkHistoryServerCluster` | `string` | — | Fully qualified Dataproc cluster name serving as the Spark History Server. |

## Examples

### Multi-Pool Cluster with Role Separation

Separate Spark drivers and executors onto different node pools for resource isolation:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: multi-pool-spark
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpDataprocVirtualCluster.multi-pool-spark
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  clusterName: multi-pool-spark
  gkeClusterTarget:
    value: projects/my-gcp-project/locations/us-central1/clusters/shared-gke
  softwareConfig:
    componentVersion:
      SPARK: "3.5"
  nodePoolTargets:
    - nodePool:
        value: driver-pool
      roles:
        - DEFAULT
        - CONTROLLER
        - SPARK_DRIVER
    - nodePool:
        value: executor-pool
      roles:
        - SPARK_EXECUTOR
      nodePoolConfig:
        autoscaling:
          minNodeCount: 2
          maxNodeCount: 20
```

### Metastore-Integrated Cluster

A virtual cluster connected to an existing Dataproc Metastore for shared Hive table access:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: metastore-spark
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpDataprocVirtualCluster.metastore-spark
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  clusterName: metastore-spark
  gkeClusterTarget:
    value: projects/my-gcp-project/locations/us-central1/clusters/shared-gke
  softwareConfig:
    componentVersion:
      SPARK: "3.5"
  nodePoolTargets:
    - nodePool:
        value: spark-pool
      roles:
        - DEFAULT
  auxiliaryServicesConfig:
    metastoreService: projects/my-gcp-project/locations/us-central1/services/shared-metastore
    sparkHistoryServerCluster: projects/my-gcp-project/regions/us-central1/clusters/history-server
```

### Using Foreign Key References

Reference other Planton-managed resources for fully composable infrastructure:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: composed-spark
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpDataprocVirtualCluster.composed-spark
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  region: us-central1
  gkeClusterTarget:
    valueFrom:
      kind: GcpGkeCluster
      name: shared-gke
      field: status.outputs.cluster_id
  kubernetesNamespace:
    valueFrom:
      kind: KubernetesNamespace
      name: spark-ns
      field: spec.name
  stagingBucket:
    valueFrom:
      kind: GcpGcsBucket
      name: spark-staging
      field: status.outputs.bucket_id
  softwareConfig:
    componentVersion:
      SPARK: "3.5"
  nodePoolTargets:
    - nodePool:
        valueFrom:
          kind: GcpGkeNodePool
          name: spark-pool
          field: status.outputs.node_pool_id
      roles:
        - DEFAULT
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | Fully qualified Dataproc cluster resource name (`projects/{project}/regions/{region}/clusters/{name}`) |
| `cluster_name` | `string` | Short name of the Dataproc cluster |
| `cluster_uuid` | `string` | Server-generated UUID for the cluster |

## Related Components

- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — provides the target GKE cluster for virtual cluster deployment
- [GcpGkeNodePool](/docs/catalog/gcp/gke-node-pool) — provides node pools assigned to Dataproc roles
- [GcpGcsBucket](/docs/catalog/gcp/gcs-bucket) — staging bucket for job dependencies
- [GcpDataprocCluster](/docs/catalog/gcp/dataproc-cluster) — standard GCE-based alternative for dedicated Spark clusters
- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project
