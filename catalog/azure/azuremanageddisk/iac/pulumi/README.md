# AzureManagedDisk Pulumi Module

## Overview

This Pulumi module provisions an Azure Managed Disk using the Azure
Classic provider (`pulumi-azure` v6). It creates a single
`compute.ManagedDisk` -- the standalone block storage volume whose data
outlives any one virtual machine -- with the configured SKU, origin
(create option), size, performance dials, encryption, and network
posture.

The VM-side attachment is deliberately NOT here: AzureVirtualMachine's
`data_disk_attachments` owns which VM mounts this disk, at which LUN,
with which caching -- so the disk survives VM replacement untouched, and
a shared disk (`max_shares`) can attach to several VMs at once.

Spec-level validation enforces the same rules ARM does -- each create
option's required source fields, the SKU gates on the performance dials,
the encryption pairings, and the network-posture pairing -- so the module
maps fields without re-validating them. The module is behaviorally
identical to the Terraform module: the same spec deploys the same disk
on either engine.

Lifecycle notes worth knowing before operating this resource:

- Name, region, zone, create option (and its source fields), logical
  sector size, security profile, and performance-plus are the disk's
  identity -- changing any of them replaces the disk **and its data**.
- `disk_size_gb` can only INCREASE. Growing an attached disk may briefly
  detach it or deallocate the VM, except where Azure supports live
  resize.
- Changing `tier` or the SKU on an attached disk deallocates the VM for
  the change and restarts it after.
- The performance dials, network posture, and tags update in place.

## Resources Created

- `compute.ManagedDisk` -- the disk itself

## Inputs

The module receives an `AzureManagedDiskIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the disk's ARM identity (references resolved to literals by the platform)
- `target.spec.storage_account_type` -- the SKU (`STANDARD_LRS` through `ULTRA_SSD_LRS`; PremiumV2/Ultra unlock the independent performance dials)
- `target.spec.create_option` + its source fields -- the disk's origin: `disk_size_gb` for EMPTY, `source_resource_id` for COPY/RESTORE, `source_uri` + `storage_account_id` for IMPORT/IMPORT_SECURE, exactly one of `image_reference_id` / `gallery_image_reference_id` for FROM_IMAGE, `upload_size_bytes` for UPLOAD
- `target.spec.os_type` / `target.spec.hyper_v_generation` -- OS-carrying disks only; unset for data disks
- `target.spec.zone` -- optional zone pin for a zonal disk (unset for regional or ZRS disks)
- `target.spec.disk_iops_read_write` / `disk_mbps_read_write` / `disk_iops_read_only` / `disk_mbps_read_only` -- the PremiumV2/Ultra performance dials (the read-only pair budgets a shared disk's read-only mounts and requires `max_shares`)
- `target.spec.tier` / `on_demand_bursting_enabled` -- classic-Premium performance decoupling and bursting
- `target.spec.max_shares` -- the shared-disk attach limit (2-10) for clustered workloads
- `target.spec.logical_sector_size` -- 512 or 4096, PremiumV2/Ultra only
- `target.spec.disk_encryption_set_id` / `secure_vm_disk_encryption_set_id` / `security_type` / `trusted_launch_enabled` -- the encryption and security surface (mutually exclusive sets; pairings enforced by the spec)
- `target.spec.network_access_policy` / `disk_access_id` / `public_network_access_enabled` -- the export posture
- `target.spec.optimized_frequent_attach_enabled` / `performance_plus_enabled` / `edge_zone` -- specialty toggles
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials, resolved by the shared provider builder (static client secret, keyless web identity, or ambient chain)

Optional enums are mapped to their ARM strings only when explicitly set
(e.g. `ALLOW_PRIVATE` becomes `AllowPrivate`), so an unspecified spec
field and Azure's default deploy identically on both engines.

## Outputs

| Output | Description |
|--------|-------------|
| `disk_id` | Full ARM ID of the disk -- the join key AzureVirtualMachine's `data_disk_attachments` references |
| `disk_name` | The disk's name as deployed |
| `disk_size_gb` | The disk's ACTUAL size in GiB (inherited from the source when the spec omitted it) |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
