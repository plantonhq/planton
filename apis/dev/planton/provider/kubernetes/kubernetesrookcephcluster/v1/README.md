# KubernetesRookCephCluster

The **KubernetesRookCephCluster** component deploys a Rook Ceph storage cluster on Kubernetes, providing distributed block, file, and object storage services through Ceph.

## Overview

This component creates a CephCluster custom resource along with optional CephBlockPools, CephFilesystems, and CephObjectStores. It requires the Rook Ceph Operator to be installed first (use `KubernetesRookCephOperator`).

### Key Features

- **Block Storage (RBD)**: High-performance block storage for databases and stateful applications
- **File Storage (CephFS)**: POSIX-compliant shared filesystem for ReadWriteMany workloads
- **Object Storage (RGW)**: S3-compatible object storage with RADOS Gateway
- **Automatic StorageClass Creation**: Creates Kubernetes StorageClasses for each storage pool
- **Self-Healing**: Automatic recovery from failures with data replication
- **Flexible Configuration**: Support for all-node or specific-node storage selection
- **Production-Ready**: Battle-tested distributed storage solution

### Use Cases

- **Persistent Block Storage**: Databases (PostgreSQL, MySQL), virtual machines, stateful applications
- **Shared File Systems**: Web server content, ML training data, shared configuration
- **Object Storage**: Backups, logs, media files, cloud-native application data
- **On-Premises Storage**: Software-defined storage for bare-metal Kubernetes clusters

## Prerequisites

Before deploying the Rook Ceph Cluster:

1. **Rook Operator Installed**: Deploy `KubernetesRookCephOperator` first
2. **Kubernetes Cluster**: Version 1.22+ with sufficient resources
3. **Raw Block Devices**: Unformatted disks or partitions on worker nodes
4. **Minimum Nodes**: At least 3 nodes for production (for replication factor of 3)

### Node Requirements

- **Linux Kernel**: 4.17+ recommended
- **LVM Package**: Required on nodes running OSDs
- **Storage Devices**: Raw block devices or partitions (not formatted)

## API Reference

### KubernetesRookCephCluster

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesRookCephCluster
metadata:
  name: <cluster-name>
spec:
  namespace:
    value: "rook-ceph"
  create_namespace: true
  operator_namespace: "rook-ceph"
  helm_chart_version: "v1.16.6"
  ceph_image:
    repository: "quay.io/ceph/ceph"
    tag: "v19.2.3"
  cluster:
    data_dir_host_path: "/var/lib/rook"
    mon:
      count: 3
    mgr:
      count: 2
    storage:
      use_all_nodes: true
      use_all_devices: true
  block_pools:
    - name: "ceph-blockpool"
      replicated_size: 3
      storage_class:
        name: "ceph-block"
        is_default: true
  filesystems:
    - name: "ceph-filesystem"
      storage_class:
        name: "ceph-filesystem"
  enable_dashboard: true
  enable_toolbox: false
```

### Spec Fields

#### Core Configuration

- **`namespace`** (required): Kubernetes namespace for the Ceph cluster
- **`operator_namespace`** (optional): Namespace where operator is installed. Default: `rook-ceph`
- **`helm_chart_version`** (optional): Rook Ceph Cluster chart version. Default: `v1.16.6`

#### Ceph Image

- **`ceph_image.repository`**: Container image repository. Default: `quay.io/ceph/ceph`
- **`ceph_image.tag`**: Container image tag. Default: `v19.2.3`

#### Cluster Configuration

- **`cluster.data_dir_host_path`**: Host path for Ceph data. Default: `/var/lib/rook`
- **`cluster.mon.count`**: Number of monitor daemons (1, 3, 5). Default: `3`
- **`cluster.mgr.count`**: Number of manager daemons. Default: `2`
- **`cluster.storage.use_all_nodes`**: Use all nodes for storage. Default: `true`
- **`cluster.storage.use_all_devices`**: Use all devices on nodes. Default: `true`
- **`cluster.storage.device_filter`**: Regex filter for device names
- **`cluster.storage.nodes`**: Specific nodes and devices for storage

#### Block Pools

- **`block_pools[].name`**: Name of the CephBlockPool
- **`block_pools[].failure_domain`**: Failure domain (host, rack, zone). Default: `host`
- **`block_pools[].replicated_size`**: Number of replicas. Default: `3`
- **`block_pools[].storage_class`**: StorageClass configuration

#### Filesystems

- **`filesystems[].name`**: Name of the CephFilesystem
- **`filesystems[].metadata_pool_replicated_size`**: Metadata pool replicas. Default: `3`
- **`filesystems[].data_pool_replicated_size`**: Data pool replicas. Default: `3`
- **`filesystems[].active_mds_count`**: Active MDS daemons. Default: `1`
- **`filesystems[].storage_class`**: StorageClass configuration

#### Object Stores

- **`object_stores[].name`**: Name of the CephObjectStore
- **`object_stores[].gateway_port`**: RGW port. Default: `80`
- **`object_stores[].gateway_instances`**: Number of RGW instances. Default: `1`
- **`object_stores[].storage_class`**: StorageClass for object bucket claims

#### Features

- **`enable_dashboard`**: Enable Ceph dashboard. Default: `true`
- **`enable_toolbox`**: Deploy toolbox pod for debugging. Default: `false`
- **`enable_monitoring`**: Enable Prometheus integration. Default: `false`

## Architecture

```
┌────────────────────────────────────────────────────────────────┐
│   Kubernetes Cluster                                           │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  CephCluster (rook-ceph)                                 │  │
│  │                                                          │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐               │  │
│  │  │   MON    │  │   MON    │  │   MON    │  Monitors     │  │
│  │  └──────────┘  └──────────┘  └──────────┘               │  │
│  │                                                          │  │
│  │  ┌──────────┐  ┌──────────┐                             │  │
│  │  │   MGR    │  │   MGR    │  Managers (Dashboard)       │  │
│  │  └──────────┘  └──────────┘                             │  │
│  │                                                          │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐               │  │
│  │  │   OSD    │  │   OSD    │  │   OSD    │  Storage      │  │
│  │  └──────────┘  └──────────┘  └──────────┘               │  │
│  │                                                          │  │
│  │  ┌──────────┐  ┌──────────┐                             │  │
│  │  │   MDS    │  │   MDS    │  CephFS (if enabled)        │  │
│  │  └──────────┘  └──────────┘                             │  │
│  │                                                          │  │
│  │  ┌──────────┐                                           │  │
│  │  │   RGW    │  Object Store (if enabled)                │  │
│  │  └──────────┘                                           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  StorageClasses                                          │  │
│  │  - ceph-block (RBD)                                      │  │
│  │  - ceph-filesystem (CephFS)                              │  │
│  │  - ceph-bucket (Object Store)                            │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
```

## Installation Methods

### Pulumi

```bash
cd iac/pulumi
pulumi up
```

See [iac/pulumi/README.md](iac/pulumi/README.md) for detailed usage.

### Terraform

```bash
cd iac/tf
terraform init
terraform plan
terraform apply
```

See [iac/tf/README.md](iac/tf/README.md) for detailed usage.

## Outputs

After deployment:

- **namespace**: Kubernetes namespace where cluster is deployed
- **helm_release_name**: Name of the Helm release
- **ceph_cluster_name**: Name of the CephCluster resource
- **block_pool_names**: Names of created CephBlockPools
- **block_storage_class_names**: Names of block StorageClasses
- **filesystem_names**: Names of created CephFilesystems
- **filesystem_storage_class_names**: Names of CephFS StorageClasses
- **object_store_names**: Names of created CephObjectStores
- **dashboard_port_forward_command**: Command to access dashboard

## Resource Requirements

### Minimum Production Deployment

| Component | Count | CPU Request | Memory Request |
|-----------|-------|-------------|----------------|
| MON       | 3     | 1000m       | 1Gi            |
| MGR       | 2     | 500m        | 512Mi          |
| OSD       | 3+    | 1000m       | 4Gi            |
| MDS       | 2     | 1000m       | 4Gi            |
| RGW       | 1+    | 1000m       | 1Gi            |

## Best Practices

1. **Three Monitors**: Always run 3 MON pods for quorum
2. **Odd Monitor Count**: Use 1, 3, or 5 monitors (never even)
3. **Dedicated Devices**: Use dedicated disks for OSDs, not root disk
4. **Failure Domains**: Spread OSDs across failure domains
5. **Dashboard Access**: Enable dashboard for monitoring and management
6. **Toolbox for Debug**: Enable toolbox when troubleshooting

## Additional Resources

- **Rook Documentation**: https://rook.io/docs/rook/latest/
- **Ceph Documentation**: https://docs.ceph.com/
- **Research Documentation**: [docs/README.md](docs/README.md)
- **Examples**: [examples.md](examples.md)

## License

Rook is a CNCF graduated project licensed under Apache License 2.0.
