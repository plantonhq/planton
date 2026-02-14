---
title: "Volume"
description: "Volume deployment documentation"
icon: "package"
order: 100
componentName: "openstackvolume"
---

# OpenStack Volume

Deploys an OpenStack Cinder block storage volume with configurable size, volume type, availability zone, and optional initialization from a Glance image, snapshot, or existing volume clone.

## What Gets Created

When you deploy an OpenStackVolume resource, OpenMCF provisions:

- **Cinder Block Storage Volume** — an `openstack_blockstorage_volume_v3` resource with the specified size and optional volume type. The volume can be created blank, initialized from a Glance image (bootable volume), restored from a snapshot, or cloned from an existing volume. These source options are mutually exclusive.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **A Cinder volume type** available in the target project if specifying `volumeType` (otherwise the project default is used)
- **A Glance image UUID** if creating a bootable volume via `imageId`
- **A snapshot UUID** if restoring from a snapshot via `snapshotId`
- **An existing volume UUID** if cloning via `sourceVolId`

## Quick Start

Create a file `volume.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: my-volume
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackVolume.my-volume
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolume/v1/iac/pulumi/module
spec:
  size: 20
```

Deploy:

```shell
openmcf apply -f volume.yaml
```

This creates a blank 20 GB Cinder volume using the project's default volume type and availability zone.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `size` | `int32` | Volume size in gigabytes (GB). For snapshot-based or clone-based volumes the value must be greater than or equal to the source size. | Must be greater than 0 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description stored on the OpenStack resource, visible in Horizon and API responses. |
| `volumeType` | `string` | project default | Cinder volume type (backend storage class), e.g. `SSD`, `HDD`, `ceph-ssd`, `lvm`. Changing this on an existing volume triggers a retype operation. |
| `availabilityZone` | `string` | Cinder default | Availability zone where the volume is created. Must match the instance AZ for attachment in most deployments. ForceNew: changing the AZ recreates the volume. |
| `snapshotId` | `string` | — | UUID of a Cinder volume snapshot to restore from. ForceNew. Mutually exclusive with `sourceVolId` and `imageId`. |
| `sourceVolId` | `string` | — | UUID of an existing Cinder volume to clone. ForceNew. Mutually exclusive with `snapshotId` and `imageId`. |
| `imageId` | `StringValueOrRef` | — | Glance image ID to initialize a bootable volume from. ForceNew. Mutually exclusive with `snapshotId` and `sourceVolId`. Can reference an OpenStackImage resource via `valueFrom`. |
| `metadata` | `map<string, string>` | `{}` | Key-value pairs stored on the volume, visible in the OpenStack API and Horizon. Useful for tagging, billing, or organizational purposes. |
| `region` | `string` | provider default | Overrides the region from the OpenStack provider config for this volume. ForceNew: changing the region recreates the volume. |

## Examples

### Blank Data Volume

A minimal blank volume suitable for attaching to a compute instance as additional storage:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: app-data
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackVolume.app-data
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolume/v1/iac/pulumi/module
spec:
  size: 50
  volumeType: SSD
  availabilityZone: az1
  description: Application data volume
```

### Bootable Volume from Glance Image

A volume initialized from a Glance image, creating a persistent boot disk that can be attached to an instance via OpenStackVolumeAttach:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: boot-disk
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackVolume.boot-disk
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolume/v1/iac/pulumi/module
spec:
  size: 40
  volumeType: SSD
  availabilityZone: az1
  imageId: 12345678-abcd-efgh-ijkl-123456789abc
  description: Ubuntu 22.04 boot volume
  metadata:
    os: ubuntu-22.04
    role: boot-disk
```

### Volume Restored from Snapshot

Restore a volume from an existing Cinder snapshot, useful for disaster recovery or creating test environments from production data:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: db-restore
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: staging.OpenstackVolume.db-restore
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolume/v1/iac/pulumi/module
spec:
  size: 100
  volumeType: SSD
  snapshotId: aabbccdd-1122-3344-5566-778899aabbcc
  description: Database volume restored from nightly snapshot
  metadata:
    source: nightly-snapshot
    environment: staging
```

### Bootable Volume with Foreign Key Reference

Reference an OpenMCF-managed Glance image instead of hardcoding a UUID. The `valueFrom` reference creates a DAG dependency so the image is provisioned before the volume:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolume
metadata:
  name: managed-boot
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackVolume.managed-boot
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolume/v1/iac/pulumi/module
spec:
  size: 50
  volumeType: SSD
  availabilityZone: az1
  imageId:
    valueFrom:
      kind: OpenStackImage
      name: ubuntu-base
      field: status.outputs.image_id
  description: Boot volume from managed image
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `volume_id` | `string` | UUID of the created Cinder volume. Primary output used as a foreign key by OpenStackVolumeAttach. |
| `name` | `string` | Name of the volume, derived from `metadata.name` |
| `size` | `int32` | Actual provisioned size of the volume in gigabytes |
| `volume_type` | `string` | Cinder volume type (backend storage class). Computed by Cinder if not explicitly specified. |
| `availability_zone` | `string` | Availability zone where the volume was created. Computed by Cinder if not explicitly specified. |
| `region` | `string` | OpenStack region where the volume was created |

## Related Components

- [OpenStack Instance](/docs/catalog/openstack/instance) — compute instance that volumes are typically attached to
- [OpenStack Volume Attach](/docs/catalog/openstack/volume-attach) — binds a volume to an instance as a DAG node in InfraCharts
- [OpenStack Image](/docs/catalog/openstack/image) — provides Glance images that can be referenced via `imageId` for bootable volumes
