---
title: "Standard Volume Attachment"
description: "This preset attaches a Cinder volume to a compute instance. The volume appears as a block device inside the instance (e.g., `/dev/vdb`). The device path is auto-assigned by Nova -- add `device` to..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "volume-attach"
componentTitle: "Volume Attach"
provider: "openstack"
icon: "package"
order: 1
---

# Standard Volume Attachment

This preset attaches a Cinder volume to a compute instance. The volume appears as a block device inside the instance (e.g., `/dev/vdb`). The device path is auto-assigned by Nova -- add `device` to request a specific path.

## When to Use

- Attaching data volumes to running instances
- InfraCharts where volume creation and attachment are separate DAG nodes
- Any instance that needs additional block storage beyond its root disk

## Key Configuration Choices

- **Join resource** -- binds a volume to an instance with no additional configuration
- **Auto-assigned device** -- Nova picks the next available device path (e.g., `/dev/vdb`, `/dev/vdc`)
- **ForceNew** -- all fields are immutable; changing either reference recreates the attachment

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<instance-id>` | ID of the compute instance to attach the volume to | OpenStack console or `OpenStackInstance` status outputs |
| `<volume-id>` | ID of the Cinder volume to attach | OpenStack console or `OpenStackVolume` status outputs |
