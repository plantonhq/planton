# AzureVirtualMachine Pulumi Module

## Overview

This Pulumi module provisions an Azure Virtual Machine using the Azure
Classic provider (`pulumi-azure` v6). ARM models Linux and Windows VMs as
separate management surfaces (different auth contracts, patch vocabularies,
and OS settings), so the module deploys exactly one of
`compute.LinuxVirtualMachine` or `compute.WindowsVirtualMachine` from the
spec's explicit OS discriminator (`os_profile.linux` XOR
`os_profile.windows`), plus one `compute.DataDiskAttachment` per referenced
data disk.

The VM is deliberately just the machine (matching Azure's own model):
network presence comes from referenced network interfaces (public IPs, NSG
filtering, and subnet placement live NIC-side), data volumes are referenced
managed disks realized as attachment resources (the data outlives the VM),
and only the OS disk is inline -- unless the VM boots from an existing
referenced OS disk (`os_managed_disk_id`), in which case the OS profile
carries no authentication fields (spec-level validation enforces the
pairing).

Only explicit spec choices are ever sent to ARM (absent optionals stay
nil), so an unspecified field and Azure's default deploy identically -- and
the module stays at behavioral parity with the Terraform module.

Lifecycle notes: name, region, zone, image source, admin credentials,
`custom_data`, and the security/confidential posture are the VM's identity
-- changing any of them replaces the VM (the OS disk with it; data disks
and NICs survive, which is exactly why they are referenced). Resizing
(`size`) reboots in place. Spot settings are fixed at creation.

## Resources Created

- `compute.LinuxVirtualMachine` OR `compute.WindowsVirtualMachine` --
  exactly one, from the OS discriminator
- `compute.DataDiskAttachment` -- one per `data_disk_attachments` entry,
  named `{vm-name}-lun-{n}`

## Inputs

The module receives an `AzureVirtualMachineIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name`
  / `target.spec.size` -- the VM's ARM identity (references resolved to
  literals by the platform)
- `target.spec.network_interface_ids` -- resolved ARM IDs of the attached
  NICs (at least one; the first is primary)
- `target.spec.os_profile` -- exactly one of `linux` (SSH-first
  authentication, patch mode, BYOS license type) or `windows` (password
  authentication, automatic updates, hotpatching, timezone, WinRM
  listeners, unattend content, Azure Hybrid Benefit), plus the optional
  hostname override
- `target.spec.os_disk` -- caching and storage SKU (required), optional
  size/name, ephemeral diff-disk settings, CMK encryption sets,
  confidential-VM `security_encryption_type`, Write Accelerator
- `target.spec.source_image_reference` / `source_image_id` /
  `os_managed_disk_id` -- exactly one image source: marketplace
  coordinates, a custom/gallery image ID, or an existing OS disk (the
  golden-disk path)
- `target.spec.data_disk_attachments` -- referenced first-class managed
  disks mounted at LUNs (0-63) with caching modes
- `target.spec.identity` -- system-assigned, user-assigned (resolved
  identity ARM IDs), or both
- `target.spec.spot` -- presence makes the VM a spot instance (eviction
  policy + optional max bid price)
- `target.spec.availability` -- zone XOR availability set, proximity
  placement, capacity reservation, dedicated host/group, Flexible
  scale-set attach + fault domain
- `target.spec.security` -- secure boot + vTPM (trusted launch),
  encryption at host
- `target.spec.patching` -- assessment mode, reboot setting, safety-check
  bypass (gated on the OS profile's AUTOMATIC_BY_PLATFORM patch mode)
- `target.spec.boot_diagnostics` / `gallery_applications` /
  `termination_notification` / `os_image_notification` / `plan` /
  `secrets` -- presence-enabled feature blocks
- `target.spec.custom_data` (secret, first boot only) / `user_data`
  (IMDS-readable, updatable, never secret)
- `target.spec.extensions_time_budget`, `provision_vm_agent`,
  `allow_extension_operations`, `disk_controller_type`,
  `additional_capabilities`, `edge_zone`
- `target.spec.tags` -- user tags, merged over the metadata-derived tags
  (user wins on collision)
- `provider_config` -- Azure credentials, resolved by the shared provider
  builder (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `vm_id` | Full ARM ID of the VM |
| `vm_name` | The VM's name as deployed |
| `virtual_machine_guid` | The 128-bit unique GUID Azure assigns the VM |
| `private_ip_address` | Primary private IP across the attached NICs |
| `public_ip_address` | Primary public IP across the attached NICs (empty for private-only VMs) |
| `computer_name` | The OS hostname the VM booted with |
| `system_assigned_identity_principal_id` | Principal ID of the system-assigned identity (empty string when the identity type doesn't include it) |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
