---
title: "Volume Attach"
description: "Volume Attach deployment documentation"
icon: "package"
order: 100
componentName: "openstackvolumeattach"
---

# OpenStack Volume Attach

Attaches an OpenStack Cinder volume to a compute instance, making the volume appear as a block device (e.g., `/dev/vdb`) inside the instance. This is a join resource that connects two independently managed resources -- a volume and an instance -- and makes the attachment relationship explicit in the deployment graph.

## What Gets Created

When you deploy an OpenStackVolumeAttach resource, OpenMCF provisions:

- **Volume Attachment** -- an `openstack_compute_volume_attach_v2` resource that connects a Cinder volume to a compute instance via the Nova API. The volume transitions from "available" to "in-use" state, and the hypervisor presents the block device to the instance at the specified (or auto-assigned) device path.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **A compute instance** in running state (created via OpenStackInstance or referenced by UUID)
- **A Cinder volume** in "available" state (created via OpenStackVolume or referenced by UUID)
- **Same availability zone** for both the volume and instance (required by most OpenStack deployments)

## Quick Start

Create a file `volume-attach.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: my-attach
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackVolumeAttach.my-attach
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolumeattach/v1/iac/pulumi/module
spec:
  instanceId: 3b4c5d6e-7f8a-9b0c-1d2e-3f4a5b6c7d8e
  volumeId: a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

Deploy:

```shell
openmcf apply -f volume-attach.yaml
```

This attaches the specified Cinder volume to the compute instance. Nova automatically assigns the next available device path (e.g., `/dev/vdb`).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `instanceId` | `StringValueOrRef` | UUID of the compute instance to attach the volume to. Can reference an OpenStackInstance resource via `valueFrom`. | Required |
| `volumeId` | `StringValueOrRef` | UUID of the Cinder volume to attach. Can reference an OpenStackVolume resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `device` | `string` | auto-assigned | Device path where the volume appears inside the instance (e.g., `/dev/vdb`, `/dev/vdc`). If omitted, Nova selects the next available device. |
| `region` | `string` | provider default | Overrides the region from the provider config for this attachment. ForceNew: changing the region recreates the attachment. |

All fields on this resource are ForceNew. Any change to the spec recreates the attachment (detach + reattach).

## Examples

### Basic Attachment with Literal IDs

Attach a volume to an instance using their UUIDs directly:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: data-attach
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackVolumeAttach.data-attach
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolumeattach/v1/iac/pulumi/module
spec:
  instanceId: 3b4c5d6e-7f8a-9b0c-1d2e-3f4a5b6c7d8e
  volumeId: a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

### Attachment with Explicit Device Path

Specify an exact device path for the volume instead of letting Nova auto-assign one. Useful when the application expects the disk at a known path:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: db-data-attach
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackVolumeAttach.db-data-attach
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolumeattach/v1/iac/pulumi/module
spec:
  instanceId: 3b4c5d6e-7f8a-9b0c-1d2e-3f4a5b6c7d8e
  volumeId: a1b2c3d4-e5f6-7890-abcd-ef1234567890
  device: /dev/vdb
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding UUIDs. The FK resolver middleware resolves these references at deployment time:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: app-data-attach
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackVolumeAttach.app-data-attach
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolumeattach/v1/iac/pulumi/module
spec:
  instanceId:
    valueFrom:
      kind: OpenStackInstance
      name: app-server
      field: status.outputs.instance_id
  volumeId:
    valueFrom:
      kind: OpenStackVolume
      name: app-data
      field: status.outputs.volume_id
  device: /dev/vdc
```

### Multiple Volumes on One Instance

Attach several volumes to the same instance, each at a different device path. Deploy each attachment as a separate resource:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: db-data-disk
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackVolumeAttach.db-data-disk
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolumeattach/v1/iac/pulumi/module
spec:
  instanceId:
    valueFrom:
      kind: OpenStackInstance
      name: db-server
      field: status.outputs.instance_id
  volumeId:
    valueFrom:
      kind: OpenStackVolume
      name: db-data
      field: status.outputs.volume_id
  device: /dev/vdb
---
apiVersion: openstack.openmcf.org/v1
kind: OpenStackVolumeAttach
metadata:
  name: db-wal-disk
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackVolumeAttach.db-wal-disk
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackvolumeattach/v1/iac/pulumi/module
spec:
  instanceId:
    valueFrom:
      kind: OpenStackInstance
      name: db-server
      field: status.outputs.instance_id
  volumeId:
    valueFrom:
      kind: OpenStackVolume
      name: db-wal
      field: status.outputs.volume_id
  device: /dev/vdc
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Terraform resource identifier for the attachment. Format: `{instance_id}/{volume_id}` |
| `instance_id` | `string` | UUID of the compute instance the volume is attached to |
| `volume_id` | `string` | UUID of the Cinder volume that was attached |
| `device` | `string` | Device path where the volume appears inside the instance (e.g., `/dev/vdb`). Computed by OpenStack if not explicitly specified in the spec. |
| `region` | `string` | OpenStack region where the attachment was created |

## Related Components

- [OpenStack Instance](/docs/catalog/openstack/instance) -- the compute instance that the volume attaches to
- [OpenStack Volume](/docs/catalog/openstack/volume) -- the Cinder block storage volume being attached
