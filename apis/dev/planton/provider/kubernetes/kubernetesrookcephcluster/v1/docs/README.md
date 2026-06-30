# Rook Ceph Cluster: Research Documentation

## Introduction

This document provides comprehensive research into deploying Ceph storage clusters on Kubernetes using the Rook operator. Ceph is a unified, distributed storage system that provides object, block, and file storage with excellent performance, reliability, and scalability.

## What is Ceph?

Ceph is a software-defined storage platform that implements object storage on a single distributed cluster and provides interfaces for object, block, and file-level storage. Originally developed at UC Santa Cruz and now maintained by Red Hat as an open-source project.

### Ceph Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Ceph Storage Cluster                         │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                          RADOS Layer                              │  │
│  │  (Reliable Autonomic Distributed Object Store)                    │  │
│  │                                                                   │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐              │  │
│  │  │   OSD   │  │   OSD   │  │   OSD   │  │   OSD   │  ...         │  │
│  │  │ (node1) │  │ (node1) │  │ (node2) │  │ (node3) │              │  │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘              │  │
│  │                                                                   │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐                           │  │
│  │  │   MON   │  │   MON   │  │   MON   │  Cluster State/Maps       │  │
│  │  └─────────┘  └─────────┘  └─────────┘                           │  │
│  │                                                                   │  │
│  │  ┌─────────┐  ┌─────────┐                                        │  │
│  │  │   MGR   │  │   MGR   │  Management & Dashboard                │  │
│  │  └─────────┘  └─────────┘                                        │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐         │
│  │    librbd       │  │    libcephfs    │  │    librados     │         │
│  │   (Block)       │  │     (File)      │  │    (Object)     │         │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘         │
│                                                                         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐         │
│  │   RBD Driver    │  │      MDS        │  │      RGW        │         │
│  │   (CSI/KRBD)    │  │  (Metadata)     │  │  (S3 Gateway)   │         │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Core Components

1. **MON (Monitors)**: Maintain cluster membership and state maps. Run in odd numbers (3, 5) for quorum.
2. **OSD (Object Storage Daemons)**: Store actual data, handle replication, recovery, and rebalancing. One per disk.
3. **MGR (Managers)**: Provide monitoring, orchestration, and external interfaces (dashboard, Prometheus).
4. **MDS (Metadata Servers)**: Required only for CephFS. Store filesystem metadata.
5. **RGW (RADOS Gateway)**: S3/Swift compatible object storage interface.

## Deployment Methods Comparison

### 1. Manual Ceph Deployment

Traditional method using `ceph-deploy` or `cephadm`:

**Pros:**
- Direct control over every configuration
- Suitable for non-Kubernetes environments

**Cons:**
- Complex manual setup
- No Kubernetes integration
- Difficult to automate
- Separate management plane

### 2. Rook Ceph Operator (Recommended)

Kubernetes-native deployment using Custom Resource Definitions:

**Pros:**
- Native Kubernetes integration
- Declarative configuration
- Self-healing and self-scaling
- CSI driver integration
- Active CNCF graduated project

**Cons:**
- Kubernetes-only
- Learning curve for Rook CRDs

### 3. Helm Chart (Ceph-CSI Only)

Deploy CSI drivers only, connect to external Ceph:

**Pros:**
- Lightweight
- Use existing Ceph clusters

**Cons:**
- Doesn't deploy Ceph itself
- Requires external Ceph management

### 4. OpenEBS with cStor (Alternative)

Another Kubernetes-native storage option:

**Pros:**
- Simpler architecture
- Lower resource requirements

**Cons:**
- Different storage model
- Less mature than Ceph

## Why Rook for Kubernetes?

Rook transforms Ceph into a cloud-native storage solution by:

1. **Automated Lifecycle Management**: Deployment, upgrades, scaling, and recovery
2. **Kubernetes-Native**: Uses CRDs, operators, and standard Kubernetes patterns
3. **CSI Integration**: Native PersistentVolume support
4. **Self-Healing**: Automatic recovery from component failures
5. **Dynamic Provisioning**: On-demand storage allocation

## Planton's Approach

### 80/20 Scoping

This component exposes the 20% of configuration that covers 80% of use cases:

**Included (Essential):**
- Cluster-level configuration (MON, MGR, OSD counts)
- Storage selection (all nodes/devices or specific)
- Block pool configuration with StorageClass
- CephFS configuration with StorageClass
- Object store configuration with StorageClass
- Dashboard and toolbox enablement

**Excluded (Advanced):**
- Fine-grained placement rules
- Custom CRUSH maps
- Advanced OSD configuration (bluestore options)
- Ceph configuration overrides
- Multi-cluster replication
- Stretched clusters

### Component Relationship

```
┌────────────────────────────────┐
│  KubernetesRookCephOperator    │  ← Install first
│  (Deploys Rook Operator)       │
└────────────────────────────────┘
              │
              ▼
┌────────────────────────────────┐
│  KubernetesRookCephCluster     │  ← This component
│  (Deploys Ceph Cluster)        │
└────────────────────────────────┘
              │
              ▼
┌────────────────────────────────┐
│  Storage Resources             │
│  - CephBlockPool + StorageClass│
│  - CephFilesystem + StorageClass
│  - CephObjectStore + StorageClass
└────────────────────────────────┘
```

## Storage Types

### Block Storage (RBD)

RBD (RADOS Block Device) provides block-level storage:

- **Access Mode**: ReadWriteOnce (single pod)
- **Use Cases**: Databases, VMs, stateful applications
- **Features**: Snapshots, cloning, encryption

**Performance Characteristics:**
- High IOPS for random access
- Low latency
- Suitable for transactional workloads

### File Storage (CephFS)

CephFS provides POSIX-compliant shared filesystem:

- **Access Mode**: ReadWriteMany (multiple pods)
- **Use Cases**: Shared content, ML datasets, log aggregation
- **Features**: Snapshots, quotas, multi-tenancy

**Performance Characteristics:**
- Good throughput for sequential access
- Suitable for file-based workloads
- Requires MDS daemons

### Object Storage (RGW)

RADOS Gateway provides S3/Swift-compatible object storage:

- **Access Mode**: HTTP/HTTPS (S3 API)
- **Use Cases**: Backups, logs, media, cloud-native apps
- **Features**: Versioning, lifecycle policies, multitenancy

**Performance Characteristics:**
- High throughput for large objects
- REST API access
- Suitable for unstructured data

## Production Best Practices

### Hardware Requirements

**Minimum Production Setup:**
- 3 nodes minimum (for replication factor 3)
- Dedicated storage nodes recommended
- SSD/NVMe for OSDs
- 10GbE networking

**Recommended Specifications:**
| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 2 cores/OSD | 4 cores/OSD |
| RAM | 4GB/OSD | 8GB/OSD |
| Network | 1GbE | 10GbE |
| Disk | SSD | NVMe |

### Network Configuration

1. **Dedicated Network**: Separate storage traffic from application traffic
2. **Jumbo Frames**: Enable 9000 MTU for storage network
3. **Encryption**: Use msgr2 with encryption for security

### Failure Domain Design

Configure failure domains based on physical topology:

- **Host**: Replicas on different nodes (default)
- **Rack**: Replicas in different racks
- **Zone**: Replicas in different availability zones

### Monitoring

Enable Prometheus integration for:
- Cluster health
- OSD performance
- Pool usage
- PG status
- RGW metrics

## Common Pitfalls

1. **Even Monitor Count**: Always use odd numbers (3, 5, 7)
2. **Co-located Daemons**: Separate storage nodes from compute
3. **Insufficient Resources**: MON/MGR need adequate memory
4. **Network Latency**: Storage network should have low latency
5. **Mixed SSD/HDD**: Don't mix without device classes
6. **Root Disk as OSD**: Never use the OS disk for OSDs

## Upgrade Considerations

1. **Test in Staging**: Always test upgrades in non-production
2. **One Component at a Time**: Upgrade operator, then cluster
3. **Health Checks**: Ensure cluster is HEALTH_OK before upgrading
4. **Backup Configuration**: Export CephCluster specs

## Comparison with Alternatives

| Feature | Rook Ceph | OpenEBS | Longhorn |
|---------|-----------|---------|----------|
| Block Storage | ✅ | ✅ | ✅ |
| File Storage | ✅ | ❌ | ❌ |
| Object Storage | ✅ | ❌ | ❌ |
| CNCF Status | Graduated | Sandbox | Incubating |
| Maturity | Very High | Medium | Medium |
| Resource Usage | Higher | Lower | Lower |
| Scalability | Excellent | Good | Good |

## Conclusion

Rook Ceph on Kubernetes provides enterprise-grade distributed storage with:

- Unified block, file, and object storage
- Self-healing and self-scaling capabilities
- Native Kubernetes integration
- Proven production reliability

The KubernetesRookCephCluster component simplifies deployment by exposing essential configuration options while maintaining production-ready defaults.

## References

- [Rook Documentation](https://rook.io/docs/rook/latest/)
- [Ceph Documentation](https://docs.ceph.com/)
- [Rook GitHub](https://github.com/rook/rook)
- [CNCF Rook Project](https://www.cncf.io/projects/rook/)
- [Ceph Performance Tuning](https://docs.ceph.com/en/latest/rados/configuration/bluestore-config-ref/)
