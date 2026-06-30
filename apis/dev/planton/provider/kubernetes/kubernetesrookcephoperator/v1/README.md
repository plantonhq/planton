# KubernetesRookCephOperator

The **KubernetesRookCephOperator** component deploys the Rook Ceph Operator on a Kubernetes cluster, enabling automated management of Ceph distributed storage through Kubernetes custom resources.

## Overview

Rook is a cloud-native storage orchestrator for Kubernetes that turns distributed storage systems into self-managing, self-scaling, and self-healing storage services. The Rook Ceph Operator automates the deployment, configuration, and management of Ceph storage clusters on Kubernetes.

### Key Features

- **Cloud-Native Storage**: First-class Kubernetes integration with native storage orchestration
- **Multiple Storage Types**: Provides block storage (RBD), file storage (CephFS), and object storage (S3/Swift compatible)
- **Self-Healing**: Automatic recovery from failures with data replication and rebuilding
- **Self-Scaling**: Dynamic cluster expansion by adding OSDs when storage is needed
- **CSI Integration**: Native Container Storage Interface drivers for Kubernetes PersistentVolumes
- **Production-Ready**: Battle-tested in production environments across cloud and on-premises deployments
- **100% Open Source**: CNCF graduated project with Apache 2.0 license

### Use Cases

- **Persistent Block Storage**: High-performance block storage for databases and stateful applications
- **Shared File Systems**: CephFS for ReadWriteMany workloads requiring shared storage
- **Object Storage**: S3-compatible object storage for cloud-native applications
- **On-Premises Storage**: Software-defined storage for bare-metal Kubernetes clusters
- **Hybrid Cloud**: Consistent storage across cloud and on-premises environments

## Prerequisites

Before deploying the Rook Ceph Operator, ensure you have:

1. **Kubernetes Cluster**: Version 1.22+ with sufficient resources
2. **Kubernetes Credentials**: Valid credentials for the target cluster
3. **Raw Block Devices**: Unformatted disks or partitions on worker nodes for OSDs
4. **Resource Capacity**: Operator requires minimal resources (defaults: 200m CPU, 128Mi memory)

### Node Requirements

- **Linux Kernel**: 4.17+ recommended for optimal Ceph performance
- **LVM Package**: Required on nodes running OSDs
- **Storage Devices**: Raw block devices, partitions, or loop devices for development

## API Reference

### KubernetesRookCephOperator

The main resource for deploying the Rook Ceph Operator.

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesRookCephOperator
metadata:
  name: <operator-name>
spec:
  targetCluster: # Optional - where to deploy
    clusterName: "<cluster-name>"
  namespace: # Required
    value: "rook-ceph"
  create_namespace: true # Optional - whether to create the namespace
  operator_version: "v1.16.6" # Optional - defaults to v1.16.6
  crds_enabled: true # Optional - let Helm manage CRDs
  container:
    resources: # Optional - defaults provided
      requests:
        cpu: 200m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
  csi: # Optional - CSI driver configuration
    enable_rbd_driver: true
    enable_cephfs_driver: true
    disable_csi_driver: false
    enable_csi_host_network: true
    provisioner_replicas: 2
```

### Spec Fields

#### `targetCluster` (optional)

Specifies the target Kubernetes cluster for operator deployment.

- **`clusterName`** (string): Name of the Kubernetes cluster
- **`clusterKind`** (enum): Type of cluster (GKE, EKS, AKS, etc.)

#### `namespace` (required)

Kubernetes namespace where the operator will be deployed. Default: `"rook-ceph"`

#### `create_namespace` (optional)

Controls whether the component should create the namespace or use an existing one.

- **`true`**: Component creates the namespace if it doesn't exist
- **`false`** (default): Component uses an existing namespace

#### `operator_version` (optional)

The Rook Ceph Operator Helm chart version. Default: `"v1.16.6"`

#### `crds_enabled` (optional)

Whether the Helm chart should create and manage CRDs. Default: `true`

**WARNING**: Only set to `false` if managing CRDs independently.

#### `container` (required)

Container configuration for the Rook Ceph Operator deployment.

- **`resources`** (object): CPU and memory allocations

#### `csi` (optional)

CSI driver configuration for Ceph storage integration.

- **`enable_rbd_driver`**: Enable RBD (block) storage driver. Default: `true`
- **`enable_cephfs_driver`**: Enable CephFS (file) storage driver. Default: `true`
- **`disable_csi_driver`**: Disable all CSI drivers. Default: `false`
- **`enable_csi_host_network`**: Enable host networking for CSI plugins. Default: `true`
- **`provisioner_replicas`**: Number of CSI provisioner replicas. Default: `2`
- **`enable_csi_addons`**: Enable CSI Addons for additional features. Default: `false`
- **`enable_nfs_driver`**: Enable NFS driver support. Default: `false`

## Architecture

```
┌─────────────────────────────────────┐
│   Kubernetes Cluster                │
│                                     │
│  ┌──────────────────────────────┐  │
│  │  Rook Operator Pod           │  │
│  │  - Watches CephCluster CRDs  │  │
│  │  - Manages Ceph Daemons      │  │
│  │  - Handles Health/Recovery   │  │
│  └──────────────────────────────┘  │
│             │                       │
│             ↓                       │
│  ┌──────────────────────────────┐  │
│  │  Custom Resource Definitions  │  │
│  │  - CephCluster               │  │
│  │  - CephBlockPool             │  │
│  │  - CephFilesystem            │  │
│  │  - CephObjectStore           │  │
│  └──────────────────────────────┘  │
│             │                       │
│             ↓                       │
│  ┌──────────────────────────────┐  │
│  │  Managed Ceph Daemons        │  │
│  │  - MON (Monitors)            │  │
│  │  - MGR (Managers)            │  │
│  │  - OSD (Object Storage)      │  │
│  │  - MDS (Metadata Server)     │  │
│  │  - RGW (RADOS Gateway)       │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
```

### How It Works

1. **Operator Installation**: Deploys operator pod and registers Custom Resource Definitions (CRDs)
2. **Cluster Creation**: Users create `CephCluster` resources defining desired storage clusters
3. **Daemon Management**: Operator creates and manages Ceph daemons (MON, MGR, OSD, MDS, RGW)
4. **Storage Provisioning**: CSI drivers enable PersistentVolume provisioning from Ceph pools
5. **Health Monitoring**: Continuous monitoring and automatic recovery from failures

## Installation Methods

### Pulumi

Deploy using Planton's Pulumi module:

```bash
cd iac/pulumi
pulumi up
```

See [iac/pulumi/README.md](iac/pulumi/README.md) for detailed Pulumi usage.

### Terraform

Deploy using Planton's Terraform module:

```bash
cd iac/tf
terraform init
terraform plan
terraform apply
```

See [iac/tf/README.md](iac/tf/README.md) for detailed Terraform usage.

## Post-Installation

After the operator is deployed, create a CephCluster to provision storage:

```yaml
apiVersion: ceph.rook.io/v1
kind: CephCluster
metadata:
  name: rook-ceph
  namespace: rook-ceph
spec:
  cephVersion:
    image: quay.io/ceph/ceph:v19.2.0
  dataDirHostPath: /var/lib/rook
  mon:
    count: 3
    allowMultiplePerNode: false
  mgr:
    count: 2
    allowMultiplePerNode: false
  storage:
    useAllNodes: true
    useAllDevices: true
```

Apply with `kubectl apply -f cephcluster.yaml`.

## Outputs

After deployment, the operator provides these outputs:

- **namespace**: Kubernetes namespace where operator is deployed
- **helm_release_name**: Name of the Helm release
- **webhook_service**: Service name for operator webhook

## Resource Requirements

### Operator Pod

| Resource | Request | Limit  |
|----------|---------|--------|
| CPU      | 200m    | 500m   |
| Memory   | 128Mi   | 512Mi  |

### Ceph Cluster Requirements

Actual resource requirements depend on your CephCluster configuration. A minimal test cluster requires:

- 3 nodes with raw block devices for OSDs
- 3 MON pods (1 per node recommended)
- 2 MGR pods
- Additional resources for MDS (CephFS) and RGW (Object Storage)

## Best Practices

1. **Dedicated Nodes**: Use dedicated storage nodes with local SSDs/NVMe
2. **Three Monitors**: Always run 3 MON pods for quorum
3. **Network Separation**: Use dedicated network for Ceph traffic if possible
4. **Failure Domains**: Spread OSDs across failure domains (racks, zones)
5. **Regular Monitoring**: Use Prometheus/Grafana for Ceph metrics

## Security Considerations

- **RBAC**: Operator requires cluster-wide permissions for CRD and daemon management
- **Encryption**: Enable encryption at rest for sensitive data
- **Network Policies**: Apply policies to restrict Ceph daemon communication

## Additional Resources

- **Rook Documentation**: https://rook.io/docs/rook/latest/
- **Rook GitHub**: https://github.com/rook/rook
- **Ceph Documentation**: https://docs.ceph.com/
- **Research Documentation**: [docs/README.md](docs/README.md) - Deep dive into deployment patterns
- **Examples**: [examples.md](examples.md) - Practical deployment scenarios

## License

Rook is a CNCF graduated project licensed under Apache License 2.0. This component (KubernetesRookCephOperator) is part of Planton and follows Planton's licensing terms.
