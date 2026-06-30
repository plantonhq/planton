# GcpDataprocVirtualCluster

Run Apache Spark workloads on an existing GKE cluster using Dataproc on GKE, instead of provisioning dedicated Compute Engine VMs.

## Overview

GcpDataprocVirtualCluster creates a Dataproc virtual cluster that schedules Spark, PySpark, and SparkR workloads as Kubernetes pods on a GKE cluster you already manage. This is the "Dataproc on GKE" deployment model — it shares GKE infrastructure across multiple Spark clusters, reducing cost and operational overhead compared to running dedicated Dataproc clusters with their own VMs.

### When to Use GcpDataprocVirtualCluster

- You already run a GKE cluster and want to consolidate Spark workloads onto it
- Multiple teams need isolated Spark environments on shared infrastructure
- You want fine-grained Kubernetes-level resource management (namespaces, quotas, RBAC)
- Your workloads benefit from GKE autoscaling and Spot/preemptible node pools

### When to Use GcpDataprocCluster Instead

- You need the full Hadoop ecosystem (HDFS, YARN, Hive, Pig, Presto)
- Your workloads require optional components like Jupyter, Zeppelin, or Flink
- You prefer fully managed VM-based clusters with no Kubernetes overhead
- You need a standalone cluster that doesn't depend on pre-existing infrastructure

## Key Configuration

### Project and Region

Every virtual cluster must specify a GCP project and region. The region must match the region of the target GKE cluster.

```yaml
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
```

### GKE Cluster Target

The `gkeClusterTarget` field is required and points to the GKE cluster where Spark pods will run. It uses the fully qualified resource ID format:

```yaml
spec:
  gkeClusterTarget:
    value: "projects/my-project/locations/us-central1/clusters/my-gke-cluster"
```

### Kubernetes Namespace

By default, Dataproc creates a namespace derived from the cluster name. You can specify an explicit namespace:

```yaml
spec:
  kubernetesNamespace:
    value: "spark-workloads"
```

### Software Config

The `softwareConfig` block defines which components to install and their versions. The `SPARK` component is mandatory:

```yaml
spec:
  softwareConfig:
    componentVersion:
      SPARK: "3.5-dataproc-17"
    properties:
      "spark:spark.executor.memory": "8g"
      "spark:spark.driver.memory": "4g"
```

Properties use the `prefix:property` format where the prefix identifies the daemon (e.g., `spark`, `mapred`).

### Node Pool Targets

Node pool targets map GKE node pools to Dataproc workload roles. At least one target must include the `DEFAULT` role:

| Role | Purpose |
|---|---|
| `DEFAULT` | Catch-all for workloads not assigned elsewhere. Required. |
| `CONTROLLER` | Cluster controller and agent pods |
| `SPARK_DRIVER` | Spark driver pods |
| `SPARK_EXECUTOR` | Spark executor pods |

```yaml
spec:
  nodePoolTargets:
    - nodePool:
        value: "default-pool"
      roles:
        - DEFAULT
    - nodePool:
        value: "spark-executors"
      roles:
        - SPARK_EXECUTOR
      nodePoolConfig:
        autoscaling:
          minNodeCount: 0
          maxNodeCount: 20
        spot: true
```

Each role can be assigned to only one node pool target. Each node pool target can hold multiple roles.

### Node Pool Config

Optionally, each node pool target can include a `nodePoolConfig` that defines the desired shape. Dataproc will create or verify a node pool matching this configuration:

- `locations`: Compute Engine zones for the nodes
- `machineType`: VM machine type (e.g., `n1-standard-4`)
- `localSsdCount`: Number of local SSDs per node
- `minCpuPlatform`: Minimum CPU generation
- `preemptible` / `spot`: Cost-optimized VMs (not for CONTROLLER or DEFAULT roles)
- `autoscaling`: Min/max node counts

### Staging Bucket

A GCS bucket for staging job dependencies and driver console output. If omitted, GCP creates a default bucket:

```yaml
spec:
  stagingBucket:
    value: "my-spark-staging-bucket"
```

### Auxiliary Services

Optional integrations for metadata management and job history:

```yaml
spec:
  auxiliaryServicesConfig:
    metastoreService: "projects/my-project/locations/us-central1/services/my-metastore"
    sparkHistoryServerCluster: "projects/my-project/regions/us-central1/clusters/history-server"
```

- **Metastore Service**: A Dataproc Metastore (managed Hive Metastore) for catalog access
- **Spark History Server**: An existing Dataproc cluster running the Spark History Server UI

## Cross-Resource References (StringValueOrRef)

Five fields in the spec support cross-resource references using `StringValueOrRef`. This enables composition where one resource can reference outputs from another:

| Field | Default Kind | Referenced Output |
|---|---|---|
| `projectId` | `GcpProject` | `status.outputs.project_id` |
| `gkeClusterTarget` | `GcpGkeCluster` | `status.outputs.cluster_id` |
| `kubernetesNamespace` | `KubernetesNamespace` | `spec.name` |
| `stagingBucket` | `GcpGcsBucket` | `status.outputs.bucket_id` |
| `nodePoolTargets[].nodePool` | `GcpGkeNodePool` | `status.outputs.node_pool_id` |

Each field can use either a literal value or a `valueFrom` reference:

```yaml
# Literal value
projectId:
  value: "my-gcp-project"

# Cross-resource reference
projectId:
  valueFrom:
    kind: GcpProject
    name: my-project
```

## IAM Prerequisites

Dataproc on GKE requires Workload Identity bindings so that Spark pods can authenticate as GCP service accounts. Before creating a virtual cluster, ensure the following IAM bindings exist:

1. The Dataproc service agent (`service-<PROJECT_NUMBER>@dataproc-accounts.iam.gserviceaccount.com`) must have the `roles/container.developer` role on the GKE cluster
2. Kubernetes service accounts in the Dataproc namespace must be bound to GCP service accounts via Workload Identity

If you manage GKE workload identity through Planton, use the `GcpGkeWorkloadIdentityBinding` component to create these bindings declaratively.

## Outputs

After deployment, the following outputs are available:

| Output | Description |
|---|---|
| `cluster_id` | Fully qualified Dataproc cluster resource name |
| `cluster_name` | Short name of the Dataproc cluster |
| `cluster_uuid` | Server-generated unique identifier |

## Quick Start

The simplest virtual cluster: one node pool with the DEFAULT role and Spark 3.5:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDataprocVirtualCluster
metadata:
  name: my-spark-on-gke
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  gkeClusterTarget:
    value: "projects/my-gcp-project/locations/us-central1/clusters/my-gke-cluster"
  softwareConfig:
    componentVersion:
      SPARK: "3.5-dataproc-17"
  nodePoolTargets:
    - nodePool:
        value: "default-pool"
      roles:
        - DEFAULT
```

## Related Components

- **GcpGkeCluster**: The target GKE cluster where Spark pods run
- **GcpGkeNodePool**: Node pools assigned to Dataproc roles
- **GcpGcsBucket**: Staging bucket for job artifacts
- **GcpGkeWorkloadIdentityBinding**: IAM bindings for Spark pod authentication
- **GcpDataprocCluster**: Standard (GCE-based) Dataproc clusters with managed VMs

## Deliberate Exclusions

The following features are excluded from v1 to keep the API focused on the 80% use case:

- **Dataproc Serverless**: Different deployment model (no cluster to manage)
- **Custom container images**: The `imageUri` field on node pool config
- **Boot disk KMS key**: CMEK for GKE node pool boot disks
- **Accelerators**: GPU/TPU attachment on node pool configs
- **Node affinity**: Kubernetes node affinity rules for pod placement
