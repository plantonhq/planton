---
title: "Dataproc Cluster"
description: "Dataproc Cluster deployment documentation"
icon: "package"
order: 100
componentName: "gcpdataproccluster"
---

# GCP Dataproc Cluster

Deploys a standard (GCE-based) Google Cloud Dataproc cluster for running Apache Spark, Hadoop, and related data processing frameworks. The component supports master/worker node configuration, optional spot secondary workers for cost optimization, software component selection, CMEK encryption, and automatic lifecycle management for ephemeral clusters.

## What Gets Created

When you deploy a GcpDataprocCluster resource, Planton provisions:

- **Dataproc Cluster** — a `google_dataproc_cluster` resource with master nodes, primary workers, and optional secondary (spot/preemptible) workers
- **GCS Staging Bucket** — auto-created by GCP if not specified; stores job dependencies and intermediate data
- **GCS Temp Bucket** — auto-created by GCP if not specified; stores ephemeral shuffle and spill data
- **Component Gateway Endpoints** — authenticated HTTPS URLs for Spark UI, YARN ResourceManager, HDFS NameNode, Jupyter, and other web UIs (when `endpointConfig.enableHttpPortAccess` is `true`)

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** with the Dataproc API enabled (`dataproc.googleapis.com`)
- **VPC network or subnetwork** if specifying custom networking (otherwise GCP uses the default network)
- **A service account** with Dataproc Worker role if using a custom service account
- **A Cloud KMS key** if enabling CMEK encryption for persistent disks
- **Initialization scripts in GCS** if using init actions

## Quick Start

Create a file `dataproc.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDataprocCluster
metadata:
  name: my-spark-cluster
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpDataprocCluster.my-spark-cluster
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: my-spark-cluster
  clusterConfig:
    masterConfig:
      machineType: n2-standard-4
    workerConfig:
      numInstances: 2
      machineType: n2-standard-4
    softwareConfig:
      imageVersion: "2.2-debian12"
    endpointConfig:
      enableHttpPortAccess: true
    lifecycleConfig:
      idleDeleteTtl: "1800s"
```

Deploy:

```shell
planton apply -f dataproc.yaml
```

This creates a Dataproc cluster with 1 master, 2 workers, Spark 3.5, Component Gateway enabled, and auto-delete after 30 minutes idle.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the cluster is created. | Required |
| `projectId.value` | `string` | Direct project ID value | — |
| `projectId.valueFrom` | `object` | Foreign key reference to a GcpProject resource | Default kind: `GcpProject` |
| `region` | `string` | GCP region for the cluster (e.g., `us-central1`). | Required |
| `clusterName` | `string` | Cluster name. Lowercase letters, numbers, hyphens. | 2-55 chars, `^[a-z][a-z0-9-]{0,53}[a-z0-9]$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `gracefulDecommissionTimeout` | `string` | `"0s"` | Duration for YARN graceful decommissioning during scale-down (e.g., `"3600s"`). |
| `clusterConfig.stagingBucket` | `StringValueOrRef` | Auto-created | GCS bucket for staging job dependencies. Can reference a GcpGcsBucket. |
| `clusterConfig.tempBucket` | `StringValueOrRef` | Auto-created | GCS bucket for ephemeral data. Can reference a GcpGcsBucket. |
| `clusterConfig.gceConfig.network` | `StringValueOrRef` | Default VPC | VPC network for nodes. Mutually exclusive with `subnetwork`. Can reference GcpVpc. |
| `clusterConfig.gceConfig.subnetwork` | `StringValueOrRef` | — | VPC subnetwork for nodes. Mutually exclusive with `network`. Can reference GcpSubnetwork. |
| `clusterConfig.gceConfig.serviceAccount` | `StringValueOrRef` | Default CE SA | Service account for node VMs. Can reference GcpServiceAccount. |
| `clusterConfig.gceConfig.zone` | `string` | Auto-selected | Zone within region for node placement. |
| `clusterConfig.gceConfig.internalIpOnly` | `bool` | `false` | Restrict nodes to internal IP addresses only. |
| `clusterConfig.gceConfig.tags` | `string[]` | `[]` | GCE network tags for firewall targeting. |
| `clusterConfig.gceConfig.metadata` | `map` | `{}` | Instance metadata key-value pairs. |
| `clusterConfig.masterConfig.numInstances` | `int` | `1` | Number of masters. Use 3 for HA mode. |
| `clusterConfig.masterConfig.machineType` | `string` | GCP default | Machine type (e.g., `n2-standard-4`). |
| `clusterConfig.masterConfig.diskConfig.bootDiskSizeGb` | `int` | `500` | Boot disk size in GB (min 10). |
| `clusterConfig.masterConfig.diskConfig.bootDiskType` | `string` | `pd-standard` | Disk type: `pd-standard`, `pd-ssd`, `pd-balanced`. |
| `clusterConfig.masterConfig.diskConfig.numLocalSsds` | `int` | `0` | Local SSDs (375 GB each). |
| `clusterConfig.masterConfig.accelerators` | `object[]` | `[]` | GPU/TPU accelerators (`acceleratorType`, `acceleratorCount`). |
| `clusterConfig.workerConfig.numInstances` | `int` | `2` | Number of primary workers. |
| `clusterConfig.workerConfig.machineType` | `string` | GCP default | Machine type for workers. |
| `clusterConfig.workerConfig.minNumInstances` | `int` | — | Minimum workers for autoscaling. |
| `clusterConfig.workerConfig.diskConfig` | `object` | — | Same structure as master disk config. |
| `clusterConfig.workerConfig.accelerators` | `object[]` | `[]` | GPU/TPU accelerators on workers. |
| `clusterConfig.secondaryWorkerConfig.numInstances` | `int` | `0` | Number of secondary (spot/preemptible) workers. |
| `clusterConfig.secondaryWorkerConfig.preemptibility` | `string` | `PREEMPTIBLE` | `SPOT`, `PREEMPTIBLE`, or `NON_PREEMPTIBLE`. |
| `clusterConfig.secondaryWorkerConfig.diskConfig` | `object` | — | Disk config for secondary workers. |
| `clusterConfig.softwareConfig.imageVersion` | `string` | Latest stable | Dataproc image version (e.g., `2.2-debian12`). |
| `clusterConfig.softwareConfig.optionalComponents` | `string[]` | `[]` | Components: `JUPYTER`, `DOCKER`, `PRESTO`, `ZEPPELIN`, `FLINK`, `TRINO`. |
| `clusterConfig.softwareConfig.properties` | `map` | `{}` | Hadoop/Spark/YARN property overrides (e.g., `"spark:spark.executor.memory": "4g"`). |
| `clusterConfig.initializationActions` | `object[]` | `[]` | Startup scripts (`script` GCS URI, optional `timeoutSec`). |
| `clusterConfig.autoscalingPolicyUri` | `string` | — | URI of a Dataproc autoscaling policy resource. |
| `clusterConfig.encryptionKmsKeyName` | `StringValueOrRef` | Google-managed | Cloud KMS key for CMEK disk encryption. Can reference GcpKmsKey. |
| `clusterConfig.endpointConfig.enableHttpPortAccess` | `bool` | `false` | Enable Component Gateway for web UI access. |
| `clusterConfig.lifecycleConfig.idleDeleteTtl` | `string` | — | Auto-delete after idle (e.g., `"1800s"` for 30 min). |
| `clusterConfig.lifecycleConfig.autoDeleteTime` | `string` | — | Scheduled deletion timestamp (RFC3339). |

## Examples

### Development Cluster with Jupyter

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDataprocCluster
metadata:
  name: dev-jupyter
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpDataprocCluster.dev-jupyter
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: dev-jupyter
  clusterConfig:
    masterConfig:
      machineType: e2-standard-4
    workerConfig:
      numInstances: 2
      machineType: e2-standard-4
    softwareConfig:
      imageVersion: "2.2-debian12"
      optionalComponents:
        - JUPYTER
    endpointConfig:
      enableHttpPortAccess: true
    lifecycleConfig:
      idleDeleteTtl: "1800s"
```

### HA Production Cluster

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDataprocCluster
metadata:
  name: prod-spark
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpDataprocCluster.prod-spark
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: prod-spark
  gracefulDecommissionTimeout: "3600s"
  clusterConfig:
    gceConfig:
      subnetwork:
        value: "projects/my-project/regions/us-central1/subnetworks/dataproc"
      serviceAccount:
        value: "dataproc-sa@my-project.iam.gserviceaccount.com"
      internalIpOnly: true
    masterConfig:
      numInstances: 3
      machineType: n2-standard-8
      diskConfig:
        bootDiskSizeGb: 200
        bootDiskType: pd-ssd
    workerConfig:
      numInstances: 5
      machineType: n2-standard-8
      diskConfig:
        bootDiskSizeGb: 500
        bootDiskType: pd-ssd
        numLocalSsds: 2
    softwareConfig:
      imageVersion: "2.2-debian12"
    encryptionKmsKeyName:
      value: "projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key"
    endpointConfig:
      enableHttpPortAccess: true
```

### Cost-Optimized Batch with Spot Workers

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDataprocCluster
metadata:
  name: batch-spark
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpDataprocCluster.batch-spark
spec:
  projectId:
    value: "my-gcp-project"
  region: us-central1
  clusterName: batch-spark
  clusterConfig:
    masterConfig:
      machineType: n2-standard-4
    workerConfig:
      numInstances: 2
      machineType: n2-standard-4
    secondaryWorkerConfig:
      numInstances: 10
      preemptibility: SPOT
    softwareConfig:
      imageVersion: "2.2-debian12"
    lifecycleConfig:
      idleDeleteTtl: "900s"
```

### Foreign Key References

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDataprocCluster
metadata:
  name: composed-spark
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpDataprocCluster.composed-spark
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
  region: us-central1
  clusterName: composed-spark
  clusterConfig:
    stagingBucket:
      valueFrom:
        kind: GcpGcsBucket
        name: staging-bucket
    gceConfig:
      subnetwork:
        valueFrom:
          kind: GcpSubnetwork
          name: dataproc-subnet
      serviceAccount:
        valueFrom:
          kind: GcpServiceAccount
          name: dataproc-sa
    encryptionKmsKeyName:
      valueFrom:
        kind: GcpKmsKey
        name: dataproc-key
    masterConfig:
      machineType: n2-standard-4
    workerConfig:
      numInstances: 4
      machineType: n2-standard-8
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | Fully qualified cluster resource name (`projects/{project}/regions/{region}/clusters/{cluster}`) |
| `cluster_name` | `string` | Short cluster name (same as `spec.clusterName`) |
| `cluster_uuid` | `string` | Server-generated unique identifier |
| `staging_bucket` | `string` | GCS bucket used for staging (user-supplied or auto-created) |

## Related Components

- [GcpGcsBucket](/docs/catalog/gcp/gcs-bucket) — Staging and temp bucket for job artifacts
- [GcpVpc](/docs/catalog/gcp/vpc) — VPC network for cluster node placement
- [GcpSubnetwork](/docs/catalog/gcp/subnetwork) — Subnetwork for controlled IP range allocation
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — Custom IAM identity for cluster VMs
- [GcpKmsKey](/docs/catalog/gcp/kms-key) — Customer-managed encryption keys for disk encryption
