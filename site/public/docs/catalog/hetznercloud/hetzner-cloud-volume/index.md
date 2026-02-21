---
title: "Hetzner Cloud Volume"
description: "Hetzner Cloud Volume deployment documentation"
icon: "package"
order: 100
componentName: "hetznercloudvolume"
---

# Hetzner Cloud Volume

Provisions a Hetzner Cloud block storage volume with an optional server attachment. Volumes persist independently of any server, making them the standard mechanism for data that must survive server replacement — databases, application state, and uploaded files. Size can be increased online after creation but can never be decreased.

## What Gets Created

- **Block Storage Volume** — an `hcloud_volume` resource with the specified size, location, optional filesystem format (ext4 or xfs), labels computed from resource metadata, and optional delete protection.
- **Volume Attachment** (when `serverId` is set) — an `hcloud_volume_attachment` resource that attaches the volume to the specified server with optional automount. Created only when the `serverId` field is present. The volume and server must be in the same location.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config
- **A server in the same location** if attaching the volume via `serverId` — either pre-existing or managed as a `HetznerCloudServer` component

## Quick Start

Create a file `volume.yaml`:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudVolume
metadata:
  name: my-volume
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudVolume.my-volume
spec:
  size: 10
  location: fsn1
```

Deploy:

```shell
openmcf apply -f volume.yaml
```

This creates a 10 GB raw (unformatted) volume in Falkenstein, unattached to any server.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `size` | `int` | Volume size in GB. Can be increased after creation but never decreased — the Hetzner Cloud API rejects size reductions. | `gte: 10`, `lte: 10240` |
| `location` | `string` | Hetzner Cloud location where the volume is stored. Known locations: `fsn1`, `nbg1`, `hel1`, `ash`, `hil`, `sin`. The volume can only be attached to servers in the same location. Changing this forces volume replacement (data loss). | `min_len: 1` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `format` | `enum` | unspecified (raw) | Filesystem format applied at creation time. Valid values: `ext4`, `xfs`. When unspecified, the volume is created as a raw block device that must be formatted manually. This is a create-time-only setting — changing it after creation has no effect. |
| `serverId` | `StringValueOrRef` | unset | Server to attach the volume to. Accepts a literal Hetzner Cloud server ID (as a string) or a reference to a `HetznerCloudServer` resource via `valueFrom`. The server must be in the same location. When set, creates an `hcloud_volume_attachment` resource. |
| `automount` | `bool` | `false` | Automatically mount the volume on the server after attaching. Only meaningful when `serverId` is set. This is a create-time-only setting. |
| `deleteProtection` | `bool` | `false` | Prevent accidental deletion of the volume via the Hetzner Cloud API. Must be disabled before the volume can be deleted. |

## Examples

### Unattached Volume with ext4

A formatted volume created without a server attachment. Useful for pre-provisioning storage that will be attached to a server later.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudVolume
metadata:
  name: app-data
  org: acme-corp
  env: staging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: storage
    pulumi.openmcf.org/stack.name: staging.HetznerCloudVolume.app-data
spec:
  size: 50
  location: fsn1
  format: ext4
```

### Volume Attached to a Server

A volume attached to a server with automount enabled. The `serverId` uses a `valueFrom` reference to a `HetznerCloudServer` component, ensuring the server is created before the volume attachment.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudVolume
metadata:
  name: db-data
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: databases
    pulumi.openmcf.org/stack.name: production.HetznerCloudVolume.db-data
spec:
  size: 100
  location: fsn1
  format: ext4
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: db-primary
      fieldPath: status.outputs.server_id
  automount: true
```

### Production Volume with Protection

A large XFS volume for high-throughput workloads with delete protection enabled. XFS is suited for large sequential writes — log aggregation, media storage, and backup repositories.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudVolume
metadata:
  name: media-storage
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: media-platform
    pulumi.openmcf.org/stack.name: production.HetznerCloudVolume.media-storage
    role: media
spec:
  size: 500
  location: fsn1
  format: xfs
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: media-processor
      fieldPath: status.outputs.server_id
  automount: true
  deleteProtection: true
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `volume_id` | `string` | Hetzner Cloud numeric ID of the created volume. Available for external reference by scripts or monitoring systems. |
| `linux_device` | `string` | The Linux device path for the volume on the attached server (e.g., `/dev/disk/by-id/scsi-0HC_Volume_12345678`). Stable across reboots. Empty if the volume is not attached. Suitable for `/etc/fstab` entries. |

## Related Components

- [HetznerCloudServer](/docs/catalog/hetznercloud/hetzner-cloud-server) — Compute instance the volume attaches to, referenced via `serverId`
