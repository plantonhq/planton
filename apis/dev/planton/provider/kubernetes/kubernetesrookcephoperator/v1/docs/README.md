# Deploying Rook Ceph Operator on Kubernetes: A Comprehensive Guide

## Introduction

Storage remains one of the most challenging aspects of running stateful workloads on Kubernetes. While Kubernetes excels at managing stateless applications, persistent storage introduces complexity around data durability, performance, and availability. **Rook** emerged as a solution to this challenge, bringing **Ceph**—one of the most battle-tested distributed storage systems—to Kubernetes as a cloud-native experience.

The Rook Ceph Operator transforms Ceph from a complex, manually-managed distributed storage system into a self-managing, self-healing Kubernetes-native service. This document explores the deployment landscape for storage on Kubernetes, why Rook Ceph stands out, and how Planton leverages it to provide production-ready storage infrastructure.

## The Kubernetes Storage Challenge

### Why Storage is Different

Kubernetes was designed with stateless workloads in mind. Pods are ephemeral, can be rescheduled anywhere, and shouldn't rely on local state. However, real-world applications need persistent storage:

- **Databases** require durable block storage with consistent I/O performance
- **Content management systems** need shared file storage across multiple pods
- **Cloud-native applications** benefit from S3-compatible object storage
- **Machine learning workflows** require high-throughput data access

### Traditional Approaches and Their Limitations

**Cloud Provider Storage (EBS, Persistent Disks, Azure Disks):**
- Simple to use but vendor-locked
- Limited to single-availability-zone attachment
- Pricing scales linearly with capacity
- No shared storage (ReadWriteMany) for most types

**Network File Systems (NFS, EFS):**
- Provides shared storage but often becomes a bottleneck
- Limited IOPS and throughput
- Single point of failure without clustering
- Not suitable for database workloads

**Storage Area Networks (SAN/iSCSI):**
- Enterprise-grade but complex to integrate with Kubernetes
- Requires specialized hardware and expertise
- Not cloud-native or easily scalable

### Enter Software-Defined Storage

Software-defined storage (SDS) solutions like Ceph, GlusterFS, and Longhorn run on commodity hardware and provide storage abstractions through software. Among these, **Ceph** stands out for its:

- **Unified storage**: Block (RBD), File (CephFS), and Object (RGW) from a single system
- **Proven reliability**: Powers some of the largest storage deployments globally
- **Self-healing**: Automatic data replication and recovery
- **Scalability**: Linear scale-out by adding nodes

## Rook: Making Ceph Cloud-Native

### What is Rook?

Rook is a CNCF graduated project that provides cloud-native storage orchestration for Kubernetes. While Rook supports multiple storage backends (Ceph, NFS, Cassandra), Ceph is the primary and most mature implementation.

The Rook Ceph Operator:
- **Deploys** Ceph clusters using Kubernetes-native resources
- **Manages** the lifecycle of all Ceph daemons (MON, OSD, MGR, MDS, RGW)
- **Integrates** with Kubernetes through CSI drivers for PersistentVolume provisioning
- **Monitors** cluster health and triggers automatic recovery
- **Upgrades** Ceph versions with rolling updates

### Why Rook Over Manual Ceph Deployment?

Running Ceph without an operator requires:
- Manual daemon deployment and configuration
- Complex networking and service discovery setup
- Custom scripts for health monitoring and recovery
- Manual handling of failures, rebalancing, and upgrades

With Rook:
- Declare desired state in YAML, operator handles implementation
- Automatic daemon placement based on available resources
- Built-in health checks and self-healing
- One-command upgrades with automatic rollback on failure

## The Deployment Maturity Spectrum

### Level 0: Anti-Pattern – Manual Ceph Deployment

**What it is**: Deploying Ceph daemons directly as Pods or StatefulSets without operator management.

**Why it fails**:
- Ceph daemons have complex interdependencies
- MON quorum requires careful bootstrapping
- OSD deployment needs device discovery and preparation
- Failure recovery requires manual intervention
- No integration with Kubernetes storage primitives

**Verdict**: Avoid. The operational complexity is prohibitive for Kubernetes environments.

### Level 1: Helm Charts Without Operators

**What it is**: Using Helm charts to template Ceph deployment resources.

**Limitations**:
- Charts can deploy initial state but don't manage ongoing operations
- No automatic recovery from failures
- Scaling requires manual OSD configuration
- Upgrades are risky without orchestration

**Verdict**: Insufficient for production. Works for initial deployment but lacks lifecycle management.

### Level 2: Rook Ceph Operator

**What it is**: Full operator-based deployment with custom resources and controllers.

**Capabilities**:
- **CephCluster CRD**: Define entire cluster topology declaratively
- **CephBlockPool CRD**: Create replicated or erasure-coded block pools
- **CephFilesystem CRD**: Deploy CephFS with metadata servers
- **CephObjectStore CRD**: S3/Swift compatible object storage
- **CephObjectStoreUser CRD**: Manage object store access credentials

**Why it works**:
- Operator has deep domain knowledge of Ceph operations
- Continuous reconciliation ensures desired state
- Automatic OSD preparation on new devices
- Rolling upgrades with health checks
- CSI integration for Kubernetes PV provisioning

**Verdict**: The production-ready choice. Matches Ceph's complexity with appropriate automation.

## Rook Architecture Deep Dive

### Core Components

**Rook Operator Pod**:
- Watches CRD resources for desired state
- Launches and manages Ceph daemon pods
- Monitors cluster health
- Orchestrates upgrades and scaling

**Ceph MON (Monitors)**:
- Maintain cluster map (topology information)
- Require quorum (minimum 3 recommended)
- Critical for cluster availability

**Ceph MGR (Managers)**:
- Provide management interfaces (dashboard, Prometheus)
- Run modules for additional functionality
- Active/standby for high availability

**Ceph OSD (Object Storage Daemon)**:
- Store actual data on physical devices
- Handle replication and recovery
- One OSD per block device or partition

**Ceph MDS (Metadata Server)** - For CephFS:
- Manage file system metadata
- Enable POSIX-compatible file access
- Active/standby or active/active configurations

**Ceph RGW (RADOS Gateway)** - For Object Storage:
- S3 and Swift compatible API
- Multi-site replication support
- Integration with identity providers

### CSI Integration

Rook deploys Container Storage Interface (CSI) drivers that integrate with Kubernetes:

**RBD CSI Driver**:
- Provisions RBD (block) volumes from Ceph pools
- Supports volume cloning and snapshots
- ReadWriteOnce access mode

**CephFS CSI Driver**:
- Provisions CephFS subvolumes
- ReadWriteMany access mode for shared storage
- Dynamic subvolume group management

**NFS CSI Driver** (optional):
- NFS-over-CephFS exports
- Broader client compatibility

## Production Best Practices

### Hardware Considerations

**Minimum Cluster**:
- 3 nodes for MON quorum
- 1 OSD per node minimum
- 10 GbE networking recommended

**Production Cluster**:
- Dedicated storage nodes with local SSDs/NVMe
- Separated networks for cluster and public traffic
- Multiple OSDs per node for performance
- 25 GbE or higher for large clusters

### Configuration Guidelines

**Monitor Placement**:
```yaml
spec:
  mon:
    count: 3  # Always odd number, minimum 3
    allowMultiplePerNode: false
```

**Manager Configuration**:
```yaml
spec:
  mgr:
    count: 2  # Active + standby
    modules:
      - name: pg_autoscaler
        enabled: true
```

**OSD Configuration**:
```yaml
spec:
  storage:
    useAllNodes: false
    useAllDevices: false
    nodes:
      - name: "storage-node-1"
        devices:
          - name: "nvme0n1"
          - name: "nvme1n1"
```

### Failure Domain Awareness

Configure failure domains to ensure data availability:

```yaml
spec:
  storage:
    nodes:
      - name: "node1"
        config:
          crush_location: "rack=rack1"
      - name: "node2"
        config:
          crush_location: "rack=rack2"
```

### Resource Allocation

| Daemon | CPU Request | Memory Request | Notes |
|--------|-------------|----------------|-------|
| MON    | 500m        | 1Gi            | Scales with cluster size |
| MGR    | 500m        | 512Mi          | More for dashboard usage |
| OSD    | 500m-2      | 2-4Gi          | Varies by device count |
| MDS    | 500m        | 1Gi            | Per active MDS |
| RGW    | 500m        | 512Mi          | Per gateway instance |

### Monitoring and Alerting

Rook integrates with Prometheus:

```yaml
spec:
  monitoring:
    enabled: true
    metricsDisabled: false
```

Key metrics to monitor:
- `ceph_health_status`: Overall cluster health
- `ceph_osd_op_latency`: OSD operation latency
- `ceph_pool_used_bytes`: Pool capacity usage
- `ceph_pg_degraded`: Degraded placement groups

## The 80/20 Configuration Principle

When designing the Planton API for KubernetesRookCephOperator, we focus on the essential 20% of configuration that 80% of users need:

**Essential Configuration (Included)**:
1. **Target Cluster**: Where to deploy the operator
2. **Namespace**: Kubernetes namespace for Rook components
3. **Operator Version**: Helm chart version for reproducibility
4. **CRD Management**: Whether operator manages CRDs
5. **Container Resources**: CPU/memory for operator pod
6. **CSI Configuration**: Which drivers to enable

**Advanced Settings (Defaulted)**:
- Node selectors and tolerations for operator pod
- Detailed CSI resource allocations
- Webhook configuration
- Discovery daemon settings
- Advanced RBAC configurations

The operator deployment is just the foundation. The actual storage cluster configuration (CephCluster CRD) contains significantly more options, which users apply separately after operator installation.

## Comparison with Alternatives

### Rook vs Longhorn

| Aspect | Rook Ceph | Longhorn |
|--------|-----------|----------|
| Storage Types | Block, File, Object | Block only |
| Complexity | Higher | Lower |
| Scalability | Massive | Medium |
| Maturity | 15+ years (Ceph) | Younger |
| Cloud-Native | Via Rook | Native |
| Best For | Enterprise, large scale | Edge, simpler needs |

### Rook vs OpenEBS

| Aspect | Rook Ceph | OpenEBS |
|--------|-----------|---------|
| Architecture | Distributed | Multiple engines |
| Storage Engines | Ceph only | Mayastor, cStor, Jiva |
| Management | Operator-based | Operator-based |
| Replication | Ceph native | Engine-dependent |
| Best For | Unified storage | Flexibility |

### When to Choose Rook Ceph

Choose Rook Ceph when you need:
- **Unified storage** (block + file + object)
- **Enterprise-proven** reliability
- **Massive scale** potential
- **On-premises** or **hybrid cloud** storage
- **S3-compatible** object storage

Consider alternatives when:
- You only need simple block storage
- Operational complexity must be minimal
- Edge or resource-constrained environments
- Cloud-provider storage is sufficient

## Planton's Approach

### Why Rook Ceph Operator?

**Open Source Foundation**: Rook is a CNCF graduated project with Apache 2.0 licensing. No vendor lock-in, no proprietary components.

**Production Proven**: Ceph powers storage at organizations like CERN, Bloomberg, and major cloud providers. Combined with Rook's operator pattern, it's battle-tested at scale.

**Comprehensive Automation**: The operator handles the complex lifecycle management that would otherwise require dedicated storage engineers.

**Kubernetes Native**: Deep integration with Kubernetes primitives—CSI, CRDs, RBAC, monitoring—makes it a natural fit for the ecosystem.

### What Planton Provides

1. **Simplified Deployment**: Deploy the operator with a clean, validated API
2. **Sensible Defaults**: Pre-configured for common production scenarios
3. **CSI Configuration**: Easy control over which storage drivers to enable
4. **Version Management**: Pinned operator versions for reproducibility
5. **Documentation**: Comprehensive examples and best practices

The KubernetesRookCephOperator deploys the foundation. Users then create CephCluster and related resources to provision actual storage, with the operator managing all complexity.

## Conclusion

Storage on Kubernetes has evolved from a significant challenge to a solved problem—thanks to operators like Rook that encode distributed systems expertise into Kubernetes controllers. The Rook Ceph Operator specifically brings the power of Ceph's unified storage to Kubernetes without requiring deep Ceph expertise.

For Planton, choosing Rook Ceph as the default storage operator means users get:
- Production-ready distributed storage
- Block, file, and object storage from a single system
- Self-healing and automatic recovery
- Kubernetes-native management

The complexity that once required dedicated storage teams is now managed by the operator, letting platform engineers focus on building applications rather than maintaining storage infrastructure.

## Additional Resources

- **Rook Documentation**: https://rook.io/docs/rook/latest/
- **Ceph Documentation**: https://docs.ceph.com/
- **Rook GitHub**: https://github.com/rook/rook
- **CNCF Rook Page**: https://www.cncf.io/projects/rook/
- **Ceph Performance Tuning**: https://docs.ceph.com/en/latest/rados/configuration/
- **Kubernetes CSI Documentation**: https://kubernetes-csi.github.io/docs/
