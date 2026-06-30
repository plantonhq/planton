---
title: "Filestore Instance"
description: "Filestore Instance deployment documentation"
icon: "package"
order: 100
componentName: "gcpfilestoreinstance"
---

# GCP Filestore Instance

Deploys a Google Cloud Filestore instance with a single NFS file share, VPC network connectivity, optional CMEK encryption, NFS export access controls, performance tuning, and deletion protection. Supports all eight Filestore tiers from cost-effective HDD to enterprise-grade regional HA.

## What Gets Created

When you deploy a GcpFilestoreInstance resource, Planton provisions:

- **Filestore Instance** — a fully managed NFS file server in the specified project and location, tagged with organization, environment, and resource labels
- **File Share** — a single NFS export on the instance with configurable capacity and access controls, mountable at `<ip>:/<share_name>`
- **VPC Network Attachment** — connects the instance to the specified VPC network using direct peering, Private Service Access, or Private Service Connect
- **NFS Export Options** — optional per-share access controls restricting which IP ranges can mount the share, with configurable read/write permissions and root squash settings

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** where the Filestore instance will be created
- **A VPC network** for the instance to connect to (referenced via `networkConfig.network`)
- **A Cloud KMS key** if using customer-managed encryption at rest (CMEK)

## Quick Start

Create a file `filestore.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFilestoreInstance
metadata:
  name: my-nfs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpFilestoreInstance.my-nfs
spec:
  projectId:
    value: my-gcp-project
  instanceName: my-nfs-server
  location: us-central1-a
  tier: BASIC_SSD
  fileShare:
    name: vol1
    capacityGb: 2560
  networkConfig:
    network:
      value: default
```

Deploy:

```shell
planton apply -f filestore.yaml
```

This creates a 2.5 TiB SSD-backed NFS instance in `us-central1-a` connected to the default VPC network.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the instance will be created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `instanceName` | `string` | Name of the Filestore instance. Becomes the GCP resource name. Immutable after creation. | Lowercase letters, numbers, hyphens; 2-63 characters; must start with a letter and end with a letter or number |
| `location` | `string` | Zone (e.g., `us-central1-a`) for zonal tiers, or region (e.g., `us-central1`) for ENTERPRISE/REGIONAL tiers. Immutable after creation. | Required |
| `tier` | `string` | Service tier controlling performance, availability, and pricing. Immutable after creation. | `STANDARD`, `PREMIUM`, `BASIC_HDD`, `BASIC_SSD`, `HIGH_SCALE_SSD`, `ZONAL`, `REGIONAL`, `ENTERPRISE` |
| `fileShare.name` | `string` | Name of the NFS file share. Becomes the export path. Immutable after creation. | Letters, numbers, underscores; 1-16 characters; must start with a letter |
| `fileShare.capacityGb` | `int` | File share capacity in GiB. Can be increased after creation but not decreased. | Minimum 1024 (tier-specific minimums enforced by GCP) |
| `networkConfig.network` | `StringValueOrRef` | VPC network the instance connects to. Immutable after creation. Can reference a GcpVpc resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description of the instance. |
| `protocol` | `string` | `NFS_V3` | NFS protocol version. `NFS_V3` or `NFS_V4_1`. NFSv4.1 available on HIGH_SCALE_SSD, ZONAL, REGIONAL, ENTERPRISE tiers. Immutable after creation. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | Cloud KMS key for CMEK encryption at rest. Immutable after creation. Can reference a GcpKmsKey resource via `valueFrom`. |
| `deletionProtectionEnabled` | `bool` | `false` | Prevents accidental deletion when enabled. |
| `deletionProtectionReason` | `string` | — | Reason for enabling deletion protection. Informational only. |
| `networkConfig.connectMode` | `string` | `DIRECT_PEERING` | Network connection mode. `DIRECT_PEERING`, `PRIVATE_SERVICE_ACCESS`, or `PRIVATE_SERVICE_CONNECT`. Immutable after creation. |
| `networkConfig.reservedIpRange` | `string` | GCP-selected | A `/29` CIDR block reserved for the instance. Must not overlap existing subnets. Immutable after creation. |
| `fileShare.nfsExportOptions` | `object[]` | All clients allowed | NFS export access controls. Maximum 10 entries per file share. |
| `fileShare.nfsExportOptions[].ipRanges` | `string[]` | All IPs | IPv4 addresses or CIDR ranges allowed to mount. Maximum 64 total across all export options. |
| `fileShare.nfsExportOptions[].accessMode` | `string` | `READ_WRITE` | `READ_ONLY` or `READ_WRITE`. |
| `fileShare.nfsExportOptions[].squashMode` | `string` | `NO_ROOT_SQUASH` | `NO_ROOT_SQUASH` or `ROOT_SQUASH`. ROOT_SQUASH maps root users to anonymous UID/GID. |
| `fileShare.nfsExportOptions[].anonUid` | `int` | `65534` | Anonymous user ID when squashMode is ROOT_SQUASH. |
| `fileShare.nfsExportOptions[].anonGid` | `int` | `65534` | Anonymous group ID when squashMode is ROOT_SQUASH. |
| `performanceConfig.fixedIops.maxIops` | `int` | Tier default | Fixed IOPS provisioning. Must be a multiple of 1000. Mutually exclusive with `iopsPerTb`. Available on ZONAL, REGIONAL, ENTERPRISE. |
| `performanceConfig.iopsPerTb.maxIopsPerTb` | `int` | Tier default | Dynamic IOPS per terabyte. Effective IOPS = capacity_tb * maxIopsPerTb. Mutually exclusive with `fixedIops`. |

## Examples

### Enterprise HA with Private Networking

A production-grade Filestore instance with regional high availability, private network connectivity, and deletion protection:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFilestoreInstance
metadata:
  name: prod-nfs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpFilestoreInstance.prod-nfs
spec:
  projectId:
    value: my-prod-project
  instanceName: prod-nfs-server
  location: us-central1
  tier: ENTERPRISE
  description: Production NFS for shared application data
  deletionProtectionEnabled: true
  deletionProtectionReason: "production data -- disable before destroying"
  fileShare:
    name: data
    capacityGb: 2048
    nfsExportOptions:
      - ipRanges:
          - "10.0.0.0/8"
        accessMode: READ_WRITE
        squashMode: ROOT_SQUASH
  networkConfig:
    network:
      value: my-vpc
    connectMode: PRIVATE_SERVICE_ACCESS
```

### CMEK-Encrypted with Performance Tuning

A zonal instance with customer-managed encryption and fixed IOPS for demanding workloads:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFilestoreInstance
metadata:
  name: perf-nfs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpFilestoreInstance.perf-nfs
spec:
  projectId:
    value: my-project
  instanceName: render-farm-nfs
  location: us-west1-a
  tier: ZONAL
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: my-kms-key
  fileShare:
    name: renders
    capacityGb: 5120
  networkConfig:
    network:
      valueFrom:
        kind: GcpVpc
        name: my-vpc
    connectMode: PRIVATE_SERVICE_ACCESS
  performanceConfig:
    fixedIops:
      maxIops: 30000
```

### NFSv4.1 with Restricted Access

An instance using NFSv4.1 protocol with IP-based access restrictions:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFilestoreInstance
metadata:
  name: secure-nfs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpFilestoreInstance.secure-nfs
spec:
  projectId:
    value: my-project
  instanceName: secure-nfs
  location: us-east1-b
  tier: ZONAL
  protocol: NFS_V4_1
  fileShare:
    name: secure_vol
    capacityGb: 1024
    nfsExportOptions:
      - ipRanges:
          - "10.0.1.0/24"
        accessMode: READ_WRITE
        squashMode: ROOT_SQUASH
      - ipRanges:
          - "10.0.2.0/24"
        accessMode: READ_ONLY
        squashMode: NO_ROOT_SQUASH
  networkConfig:
    network:
      value: my-vpc
```

### Cost-Effective HDD Storage

A budget-friendly HDD-backed instance for infrequently accessed data:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFilestoreInstance
metadata:
  name: archive-nfs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpFilestoreInstance.archive-nfs
spec:
  projectId:
    value: my-project
  instanceName: archive-nfs
  location: us-central1-a
  tier: STANDARD
  fileShare:
    name: archive
    capacityGb: 1024
  networkConfig:
    network:
      value: default
```

### Infra-Chart Composition with valueFrom

An instance composing with other Planton resources via foreign key references:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpFilestoreInstance
metadata:
  name: shared-nfs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpFilestoreInstance.shared-nfs
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
  instanceName: shared-nfs
  location: us-central1-a
  tier: BASIC_SSD
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: my-encryption-key
  fileShare:
    name: shared
    capacityGb: 2560
  networkConfig:
    network:
      valueFrom:
        kind: GcpVpc
        name: my-vpc
    connectMode: PRIVATE_SERVICE_ACCESS
    reservedIpRange: "10.10.0.0/29"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | Fully qualified resource ID (`projects/{project}/locations/{location}/instances/{instance}`) |
| `instance_name` | `string` | Short name of the Filestore instance |
| `ip_addresses` | `string[]` | IP addresses on the connected VPC network. Use the first address for NFS mounts. |
| `file_share_name` | `string` | Name of the file share. Mount path: `<ip_addresses[0]>:/<file_share_name>` |
| `create_time` | `string` | Instance creation timestamp (RFC3339 format) |

## Related Components

- [GcpVpc](/docs/catalog/gcp/vpc) — VPC network the Filestore instance connects to
- [GcpKmsKey](/docs/catalog/gcp/kms-key) — KMS key for customer-managed encryption at rest
- [GcpKmsKeyRing](/docs/catalog/gcp/kms-key-ring) — Key ring containing the KMS key
- [GcpProject](/docs/catalog/gcp/project) — GCP project where the instance is created
- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — GKE clusters can mount Filestore shares via PersistentVolume
