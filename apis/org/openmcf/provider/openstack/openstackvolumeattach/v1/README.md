# OpenStackVolumeAttach

Attach an OpenStack Cinder volume to a compute instance, managed through OpenMCF.

## Overview

`OpenStackVolumeAttach` is a "join" resource that connects a Cinder volume (persistent block storage) to a Nova instance (compute). The attachment makes the volume appear as a block device inside the instance (e.g., `/dev/vdb`).

This follows the same join-resource pattern as `OpenStackRouterInterface` (Router + Subnet) and `OpenStackFloatingIpAssociate` (FloatingIp + Port).

## When to Use

- **Add persistent storage to an instance**: Attach a data volume for databases, logs, or application state
- **InfraChart DAG visibility**: Volume and Instance are created independently; VolumeAttach connects them with explicit dependency edges
- **Detach/reattach workflow**: Volumes can be detached from one instance and attached to another

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `instance_id` | StringValueOrRef | **yes** | FK to OpenStackInstance |
| `volume_id` | StringValueOrRef | **yes** | FK to OpenStackVolume |
| `device` | string | no | Device path (e.g., "/dev/vdb"). Auto-selected if omitted |
| `region` | string | no | Region override |

## Outputs

| Output | Description |
|--------|-------------|
| `id` | Terraform resource ID |
| `instance_id` | Attached instance UUID |
| `volume_id` | Attached volume UUID |
| `device` | Device path in the instance |
| `region` | OpenStack region |

## Foreign Key Relationships

- `instance_id` -> `OpenStackInstance.status.outputs.instance_id` (required)
- `volume_id` -> `OpenStackVolume.status.outputs.volume_id` (required)
