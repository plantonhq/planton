---
title: "Rook Ceph Cluster"
description: "Rook Ceph Cluster deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesrookcephcluster"
---

# Kubernetes Rook Ceph Cluster

Deploys a production-grade Ceph distributed storage cluster on Kubernetes using the Rook operator. Provides block (RBD), file (CephFS), and object (S3-compatible RGW) storage through a single declarative resource, with automatic StorageClass creation, Ceph dashboard, and toolbox support for debugging.

## What Gets Created

When you deploy a KubernetesRookCephCluster resource, OpenMCF provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **Rook Ceph Cluster Helm Release** — deploys the `rook-ceph-cluster` chart from the official Rook repository, which creates:
  - CephCluster custom resource with configurable MON, MGR, and OSD daemons
  - Ceph dashboard (SSL-enabled) for web-based cluster management
  - Ceph toolbox deployment for CLI debugging (when enabled)
  - Prometheus monitoring integration (when enabled)
- **CephBlockPool resources** — one per entry in `blockPools`, providing RBD-backed persistent volumes
- **CephFilesystem resources** — one per entry in `filesystems`, providing CephFS shared filesystem storage
- **CephObjectStore resources** — one per entry in `objectStores`, providing S3-compatible RADOS Gateway endpoints
- **Kubernetes StorageClasses** — automatically created for each block pool, filesystem, and object store that has `storageClass.enabled` set to `true`

## Prerequisites

- **A Kubernetes cluster** with the Rook Ceph Operator already installed (the operator manages the CephCluster lifecycle)
- **kubectl** configured to access the target cluster
- **Raw block devices or partitions** available on cluster nodes for OSD storage (Ceph requires unformatted disks)
- **At least three nodes** for production deployments to satisfy the default replication factor of 3

## Quick Start

Create a file `ceph-cluster.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRookCephCluster
metadata:
  name: my-ceph
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesRookCephCluster.my-ceph
spec:
  namespace:
    value: rook-ceph
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f ceph-cluster.yaml
```

This creates a Ceph cluster using all nodes and all available devices with 3 MON daemons, 2 MGR daemons, the dashboard enabled, and no block pools, filesystems, or object stores (add them in the spec as needed).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace where the Ceph cluster will be deployed. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `true` | Create the namespace if it does not exist. |
| `operatorNamespace` | `string` | `"rook-ceph"` | Namespace where the Rook Ceph Operator is installed. |
| `helmChartVersion` | `string` | `"v1.16.6"` | Version of the Rook Ceph Cluster Helm chart. |
| `enableToolbox` | `bool` | `false` | Deploy the Ceph toolbox pod for CLI debugging (`ceph status`, `rados`, etc.). |
| `enableMonitoring` | `bool` | `false` | Enable Prometheus monitoring integration for Ceph daemons. |
| `enableDashboard` | `bool` | `true` | Enable the Ceph web dashboard for cluster management. |
| `cephImage.repository` | `string` | `"quay.io/ceph/ceph"` | Container image repository for Ceph daemons. |
| `cephImage.tag` | `string` | `"v19.2.3"` | Container image tag for Ceph daemons. |
| `cephImage.allowUnsupported` | `bool` | `false` | Allow unsupported Ceph versions. Not recommended for production. |
| `cluster.dataDirHostPath` | `string` | `"/var/lib/rook"` | Host path for Ceph configuration and data persistence. Must be unique per Ceph cluster. |
| `cluster.mon.count` | `int` | `3` | Number of MON daemons. Must be odd (1, 3, 5) for quorum. Range: 1-9. |
| `cluster.mon.allowMultiplePerNode` | `bool` | `false` | Allow multiple MON daemons on the same node. |
| `cluster.mgr.count` | `int` | `2` | Number of MGR daemons. Use 2 for high availability. Range: 1-5. |
| `cluster.mgr.allowMultiplePerNode` | `bool` | `false` | Allow multiple MGR daemons on the same node. |
| `cluster.storage.useAllNodes` | `bool` | `true` | Use all cluster nodes for OSD storage. |
| `cluster.storage.useAllDevices` | `bool` | `true` | Use all available devices on each node. |
| `cluster.storage.deviceFilter` | `string` | — | Regex filter for device names (e.g., `"^sd[a-z]$"`). |
| `cluster.storage.nodes` | `CephStorageNodeSpec[]` | `[]` | Per-node storage configuration. Only used when `useAllNodes` is `false`. |
| `cluster.storage.nodes[].name` | `string` | — | Node name matching the `kubernetes.io/hostname` label. Required. |
| `cluster.storage.nodes[].devices` | `string[]` | `[]` | Specific device names to use on this node. |
| `cluster.storage.nodes[].deviceFilter` | `string` | — | Device filter pattern for this node. |
| `cluster.network.enableEncryption` | `bool` | `false` | Encrypt data in transit between Ceph daemons. Requires kernel 5.11+. |
| `cluster.network.enableCompression` | `bool` | `false` | Compress data in transit between daemons. |
| `cluster.network.requireMsgr2` | `bool` | `false` | Require msgr2 protocol and disable legacy msgr v1. |
| `cluster.resources.mon` | `ContainerResources` | — | CPU/memory requests and limits for MON daemons. |
| `cluster.resources.mgr` | `ContainerResources` | — | CPU/memory requests and limits for MGR daemons. |
| `cluster.resources.osd` | `ContainerResources` | — | CPU/memory requests and limits for OSD daemons. |
| `blockPools` | `CephBlockPoolSpec[]` | `[]` | Block storage pools (RBD) to create. |
| `blockPools[].name` | `string` | — | Name of the block pool. Required. |
| `blockPools[].failureDomain` | `string` | `"host"` | Failure domain for data placement (`host`, `rack`, `zone`). |
| `blockPools[].replicatedSize` | `int` | `3` | Number of data replicas. Range: 1-7. |
| `blockPools[].storageClass` | `CephStorageClassSpec` | — | StorageClass configuration for this pool. |
| `filesystems` | `CephFilesystemSpec[]` | `[]` | CephFS filesystems to create. |
| `filesystems[].name` | `string` | — | Name of the filesystem. Required. |
| `filesystems[].metadataPoolReplicatedSize` | `int` | `3` | Metadata pool replication count. Range: 1-7. |
| `filesystems[].dataPoolReplicatedSize` | `int` | `3` | Data pool replication count. Range: 1-7. |
| `filesystems[].failureDomain` | `string` | `"host"` | Failure domain for data placement. |
| `filesystems[].activeMdsCount` | `int` | `1` | Number of active MDS daemons. Range: 1-10. |
| `filesystems[].activeStandby` | `bool` | `true` | Enable active-standby MDS for high availability. |
| `filesystems[].mdsResources` | `ContainerResources` | — | CPU/memory requests and limits for MDS daemons. |
| `filesystems[].storageClass` | `CephStorageClassSpec` | — | StorageClass configuration for this filesystem. |
| `objectStores` | `CephObjectStoreSpec[]` | `[]` | Ceph object stores (RGW) to create. |
| `objectStores[].name` | `string` | — | Name of the object store. Required. |
| `objectStores[].metadataPoolReplicatedSize` | `int` | `3` | Metadata pool replication count. Range: 1-7. |
| `objectStores[].dataPoolErasureDataChunks` | `int` | `2` | Erasure coding data chunks for the data pool. Range: 2-16. |
| `objectStores[].dataPoolErasureCodingChunks` | `int` | `1` | Erasure coding parity chunks for the data pool. Range: 1-8. |
| `objectStores[].failureDomain` | `string` | `"host"` | Failure domain for data placement. |
| `objectStores[].preservePoolsOnDelete` | `bool` | `true` | Preserve RADOS pools when the object store is deleted. |
| `objectStores[].gatewayPort` | `int` | `80` | RGW gateway listen port. Range: 1-65535. |
| `objectStores[].gatewayInstances` | `int` | `1` | Number of RGW gateway pod instances. Range: 1-10. |
| `objectStores[].gatewayResources` | `ContainerResources` | — | CPU/memory requests and limits for gateway pods. |
| `objectStores[].storageClass` | `CephStorageClassSpec` | — | StorageClass configuration for object bucket claims. |

**StorageClass fields** (shared by `blockPools[].storageClass`, `filesystems[].storageClass`, and `objectStores[].storageClass`):

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `storageClass.enabled` | `bool` | `true` | Create a Kubernetes StorageClass for this pool. |
| `storageClass.name` | `string` | — | Name of the StorageClass. Required when enabled. |
| `storageClass.isDefault` | `bool` | `false` | Set as the default StorageClass in the cluster. |
| `storageClass.reclaimPolicy` | `string` | `"Delete"` | Reclaim policy (`Delete` or `Retain`). |
| `storageClass.allowVolumeExpansion` | `bool` | `true` | Allow persistent volume expansion after creation. |
| `storageClass.volumeBindingMode` | `string` | `"Immediate"` | Volume binding mode (`Immediate` or `WaitForFirstConsumer`). |

## Examples

### Block Storage Only

A Ceph cluster with a single replicated block pool and a default StorageClass, suitable for general-purpose persistent volumes:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRookCephCluster
metadata:
  name: block-ceph
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesRookCephCluster.block-ceph
spec:
  namespace:
    value: rook-ceph
  createNamespace: true
  enableToolbox: true
  blockPools:
    - name: replicated-pool
      replicatedSize: 3
      failureDomain: host
      storageClass:
        enabled: true
        name: ceph-block
        isDefault: true
        reclaimPolicy: Delete
        allowVolumeExpansion: true
```

### Production Multi-Storage with Resource Tuning

A production deployment with block, file, and object storage, explicit daemon resources, and monitoring enabled:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRookCephCluster
metadata:
  name: prod-ceph
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesRookCephCluster.prod-ceph
spec:
  namespace:
    value: rook-ceph
  createNamespace: true
  enableToolbox: true
  enableMonitoring: true
  enableDashboard: true
  cluster:
    mon:
      count: 3
    mgr:
      count: 2
    storage:
      useAllNodes: true
      useAllDevices: true
    resources:
      mon:
        limits:
          cpu: "2000m"
          memory: "2Gi"
        requests:
          cpu: "500m"
          memory: "1Gi"
      mgr:
        limits:
          cpu: "1000m"
          memory: "1Gi"
        requests:
          cpu: "250m"
          memory: "512Mi"
      osd:
        limits:
          cpu: "2000m"
          memory: "4Gi"
        requests:
          cpu: "500m"
          memory: "2Gi"
    network:
      enableEncryption: true
      requireMsgr2: true
  blockPools:
    - name: replicated-pool
      replicatedSize: 3
      failureDomain: host
      storageClass:
        enabled: true
        name: ceph-block
        isDefault: true
        reclaimPolicy: Delete
  filesystems:
    - name: shared-fs
      metadataPoolReplicatedSize: 3
      dataPoolReplicatedSize: 3
      failureDomain: host
      activeMdsCount: 2
      activeStandby: true
      mdsResources:
        limits:
          cpu: "2000m"
          memory: "4Gi"
        requests:
          cpu: "500m"
          memory: "1Gi"
      storageClass:
        enabled: true
        name: ceph-filesystem
        reclaimPolicy: Delete
  objectStores:
    - name: s3-store
      metadataPoolReplicatedSize: 3
      dataPoolErasureDataChunks: 2
      dataPoolErasureCodingChunks: 1
      failureDomain: host
      preservePoolsOnDelete: true
      gatewayPort: 80
      gatewayInstances: 2
      gatewayResources:
        limits:
          cpu: "2000m"
          memory: "2Gi"
        requests:
          cpu: "500m"
          memory: "1Gi"
      storageClass:
        enabled: true
        name: ceph-bucket
```

### Targeted Node Storage with Device Filtering

A deployment that targets specific nodes and devices rather than using all available storage:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRookCephCluster
metadata:
  name: targeted-ceph
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesRookCephCluster.targeted-ceph
spec:
  namespace:
    value: rook-ceph
  createNamespace: true
  operatorNamespace: rook-ceph
  helmChartVersion: v1.16.6
  cephImage:
    repository: quay.io/ceph/ceph
    tag: v19.2.3
  enableDashboard: true
  enableToolbox: true
  cluster:
    dataDirHostPath: /var/lib/rook
    mon:
      count: 3
      allowMultiplePerNode: false
    mgr:
      count: 2
    storage:
      useAllNodes: false
      useAllDevices: false
      nodes:
        - name: storage-node-01
          devices:
            - sdb
            - sdc
        - name: storage-node-02
          deviceFilter: "^sd[b-d]$"
        - name: storage-node-03
          devices:
            - sdb
            - sdc
            - sdd
  blockPools:
    - name: fast-pool
      replicatedSize: 3
      failureDomain: host
      storageClass:
        enabled: true
        name: ceph-block-fast
        isDefault: true
        reclaimPolicy: Retain
        volumeBindingMode: WaitForFirstConsumer
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the Ceph cluster is deployed |
| `helmReleaseName` | `string` | Name of the Helm release for the Rook Ceph Cluster |
| `cephClusterName` | `string` | Name of the CephCluster custom resource |
| `blockPoolNames` | `string[]` | Names of the created CephBlockPool resources |
| `blockStorageClassNames` | `string[]` | Names of the created StorageClasses for block storage |
| `filesystemNames` | `string[]` | Names of the created CephFilesystem resources |
| `filesystemStorageClassNames` | `string[]` | Names of the created StorageClasses for CephFS |
| `objectStoreNames` | `string[]` | Names of the created CephObjectStore resources |
| `objectStorageClassNames` | `string[]` | Names of the created StorageClasses for object bucket claims |
| `dashboardPortForwardCommand` | `string` | Ready-to-run `kubectl port-forward` command for dashboard access on port 7000 |
| `dashboardUrl` | `string` | URL to access the Ceph dashboard after port-forwarding (`https://localhost:7000`) |
| `dashboardPasswordCommand` | `string` | Command to retrieve the Ceph dashboard admin password from the Kubernetes secret |
| `toolboxExecCommand` | `string` | Command to exec into the Ceph toolbox pod for CLI debugging |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — deploy the Rook Ceph Operator prerequisite via Helm
- [KubernetesPrometheus](/docs/catalog/kubernetes/kubernetesprometheus) — set up Prometheus to consume Ceph monitoring metrics
- [KubernetesStatefulSet](/docs/catalog/kubernetes/kubernetesstatefulset) — deploy stateful workloads backed by Ceph block or filesystem storage
