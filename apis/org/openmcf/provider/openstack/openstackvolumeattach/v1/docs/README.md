# OpenStackVolumeAttach: Research & Design Documentation

## OpenStack Compute Volume Attachment

Volume attachments are managed through the Nova Compute API (not Cinder). When you attach a volume, Nova requests Cinder to transition the volume from "available" to "in-use" state, and then instructs the hypervisor to present the block device to the instance.

### Key Concepts

- **Volume state**: Must be "available" to attach, transitions to "in-use"
- **Device path**: Auto-assigned by Nova if not specified (e.g., /dev/vdb, /dev/vdc)
- **AZ constraint**: Volume and instance must be in the same availability zone (in most deployments)
- **Single-attach**: By default, a volume can only be attached to one instance at a time

### Terraform Resource: `openstack_compute_volume_attach_v2`

The Terraform provider has 6 fields. We selected 4 for the 80/20 set:

| Included | Excluded | Reason for Exclusion |
|----------|----------|---------------------|
| instance_id | multiattach | SAN shared volumes, niche |
| volume_id | tag | PCI device tagging, niche |
| device | vendor_options | TF-specific workaround |
| region | | |

### Join Resource Pattern

VolumeAttach follows the same "join resource" pattern as:
- **RouterInterface** (Router + Subnet): Two required FKs, no CEL validations
- **FloatingIpAssociate** (FloatingIp + Port): Two required FKs, optional plain fields

The pattern is: two required `StringValueOrRef` FKs + optional plain fields + region.

### Pulumi Resource: `openstack.compute.VolumeAttach`

The Pulumi SDK uses `compute.NewVolumeAttach()` (in the compute package, not blockstorage). The `VolumeAttachArgs` struct has `InstanceId` and `VolumeId` as required `pulumi.StringInput` fields, plus optional `Device` and `Region`.

Note: There is also a `blockstorage.VolumeAttach` resource in the Pulumi SDK, but that maps to the Cinder-level admin API (`openstack_blockstorage_volume_attach_v3`). We use the compute variant which is the tenant-facing Nova API.

## Design Decisions

1. **`compute` not `blockstorage` Pulumi package** -- The TF resource is `openstack_compute_volume_attach_v2` (Nova), not `openstack_blockstorage_volume_attach_v3` (Cinder admin). The Pulumi `compute.VolumeAttach` maps to the correct Nova resource.

2. **`device` as plain string** -- Device paths are literal values (e.g., "/dev/vdb"), not resource references. No FK needed.

3. **No `multiattach` field** -- Multi-attach requires special volume types and is a niche SAN use case. The 80/20 principle excludes it.

4. **No CEL validations** -- Only required FKs and optional plain strings. Same as RouterInterface (simplest join pattern).
