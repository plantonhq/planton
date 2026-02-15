# GcpFilestoreInstance

An OpenMCF deployment component for provisioning Google Cloud Filestore instances.

## Overview

Filestore provides fully managed, high-performance NFS file storage. This component creates a Filestore instance with a single file share, connecting it to a VPC network of your choice. The resulting NFS export can be mounted by any client in the connected VPC.

## Key Features

- **All eight tiers**: STANDARD, PREMIUM, BASIC_HDD, BASIC_SSD, HIGH_SCALE_SSD, ZONAL, REGIONAL, ENTERPRISE
- **NFS export controls**: IP-based access restrictions with configurable read/write and root squash settings
- **CMEK encryption**: Customer-managed encryption keys via Cloud KMS
- **Performance tuning**: Fixed IOPS or per-TB IOPS provisioning (ZONAL/REGIONAL/ENTERPRISE)
- **Deletion protection**: Prevents accidental destruction of production instances
- **NFSv3 and NFSv4.1**: Protocol selection for compatibility or security requirements
- **Infra-chart composability**: All cross-resource references use `StringValueOrRef` for dependency wiring

## Quick Start

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpFilestoreInstance
metadata:
  name: my-nfs
spec:
  projectId: my-gcp-project
  instanceName: my-nfs-server
  location: us-central1-a
  tier: BASIC_SSD
  fileShare:
    name: vol1
    capacityGb: 2560
  networkConfig:
    network: default
```

## Outputs

After deployment, the following values are available in `status.outputs`:

| Output | Description |
|--------|-------------|
| `instanceId` | Fully qualified resource ID |
| `instanceName` | Short name of the instance |
| `ipAddresses` | IP addresses for NFS mounts |
| `fileShareName` | File share name for mount path |
| `createTime` | Creation timestamp |

Mount the NFS share: `mount <ipAddresses[0]>:/<fileShareName> /mnt/nfs`

## Presets

| Preset | Tier | Use Case |
|--------|------|----------|
| dev-basic | BASIC_SSD | Development, testing, CI/CD |
| production-enterprise | ENTERPRISE | Production with regional HA |
| high-performance-zonal | ZONAL | Performance-sensitive workloads with IOPS tuning |

## Documentation

- [Examples](examples.md) — YAML manifests for common configurations
- [Research](docs/README.md) — Comprehensive analysis of Filestore deployment landscape
- [Catalog Page](catalog-page.md) — Full configuration reference
