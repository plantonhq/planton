# Scaleway Block Volume

Deploys a Scaleway Block Storage volume as a standalone, network-attached SSD block device in a specified Availability Zone with a configurable performance tier and size. The volume persists independently of any Instance lifecycle and can be moved between Instances in the same zone.

## What Gets Created

When you deploy a ScalewayBlockVolume resource, Planton provisions:

- **Block Volume** — a `block.Volume` resource providing a raw, network-attached NVMe-backed block device with the specified size and IOPS tier
- **Scaleway Tags** — standard Planton resource tags applied to the volume for organization, environment, and resource identification

## Prerequisites

- **Scaleway credentials** configured via environment variables or Planton provider config
- **An Availability Zone** that matches the zone of the Instance to which the volume will be attached (block volumes are zonal resources)
- **OS-level formatting** — Scaleway block volumes are raw block devices; after attaching to an Instance you must format (`mkfs.ext4`, `mkfs.xfs`, etc.) and mount the volume

## Quick Start

Create a file `block-volume.yaml`:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayBlockVolume
metadata:
  name: my-volume
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayBlockVolume.my-volume
spec:
  zone: fr-par-1
  sizeGb: 20
  performanceTier: sbs_5k
```

Deploy:

```shell
planton apply -f block-volume.yaml
```

This creates a 20 GB block volume with standard performance (5,000 IOPS) in the `fr-par-1` Availability Zone.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zone` | `string` | Scaleway Availability Zone where the volume is created (e.g., `"fr-par-1"`, `"nl-ams-1"`, `"pl-waw-2"`). Cannot be changed after creation. | Required |
| `sizeGb` | `uint32` | Volume size in gigabytes. Can be increased in-place after creation but cannot be shrunk. After increasing via IaC, grow the partition and filesystem inside the OS. | Required, 5–10240 |
| `performanceTier` | `enum` | IOPS performance tier. Options: `sbs_5k` (5,000 IOPS, standard) or `sbs_15k` (15,000 IOPS, high performance). Can be changed in-place after creation. | Required, must be `sbs_5k` or `sbs_15k` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `snapshotId` | `string` | — | Block Storage snapshot UUID to clone the volume from. Must be in the same zone. When set, `sizeGb` must be >= the snapshot's source volume size. If omitted, a blank volume is created. |

## Examples

### Development Data Volume

A minimal 10 GB volume for a development Instance in Paris:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayBlockVolume
metadata:
  name: dev-data
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayBlockVolume.dev-data
spec:
  zone: fr-par-1
  sizeGb: 10
  performanceTier: sbs_5k
```

### High-Performance Database Volume

A 500 GB volume with the high-performance tier for a database workload in Amsterdam. The `sbs_15k` tier provides 15,000 IOPS suitable for PostgreSQL, MySQL, or MongoDB data directories:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayBlockVolume
metadata:
  name: prod-db-data
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayBlockVolume.prod-db-data
spec:
  zone: nl-ams-1
  sizeGb: 500
  performanceTier: sbs_15k
```

### Volume from Snapshot

A volume restored from an existing Block Storage snapshot. The size must be at least as large as the snapshot's source volume:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayBlockVolume
metadata:
  name: restored-volume
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.ScalewayBlockVolume.restored-volume
spec:
  zone: fr-par-1
  sizeGb: 100
  performanceTier: sbs_5k
  snapshotId: 11111111-1111-1111-1111-111111111111
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `volume_id` | `string` | Zoned identifier of the created volume (format: `"{zone}/{uuid}"`). Primary output for downstream cross-resource references. |
| `volume_name` | `string` | Name of the volume as it exists in Scaleway Block Storage. Derived from `metadata.name`. |
| `zone` | `string` | Availability Zone where the volume is deployed. Used by downstream resources to verify zone co-location. |

## Related Components

- [ScalewayPrivateNetwork](/docs/catalog/scaleway/scalewayprivatenetwork) — provides private connectivity between Instances and other Scaleway resources in the same region
- [ScalewayKapsuleCluster](/docs/catalog/scaleway/scalewaykapsulecluster) — deploys Kubernetes clusters whose nodes can use block volumes for persistent storage
- [ScalewayRdbInstance](/docs/catalog/scaleway/scalewayrdbinstance) — deploys managed databases that can use block storage volume types for data persistence
