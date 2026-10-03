# AzureVirtualMachine

## Overview

`AzureVirtualMachine` provisions an Azure Virtual Machine: the compute
instance itself — its size, image, OS profile, OS disk, identity,
placement, and security posture.

The VM is deliberately just the machine. Everything it composes with is
referenced, never created here — matching Azure's own model, where a VM
is a compute shell wired to first-class resources:

- **Network presence** is one or more referenced `AzureNetworkInterface`
  resources (`network_interface_ids`, at least one; the first is
  primary). Public IPs, NSG filtering, and subnet placement all live on
  the NIC.
- **Data volumes** are referenced `AzureManagedDisk` resources mounted at
  a LUN with a caching mode (`data_disk_attachments`) — the data outlives
  the machine.
- **Identities** are referenced `AzureUserAssignedIdentity` resources or
  the VM's own system-assigned principal (surfaced in the outputs);
  grants are composed with `AzureRoleAssignment`.

Only the OS disk is inline: it is born and dies with the VM by
definition — unless the VM boots from an existing referenced OS disk
(`os_managed_disk_id`), the disk-swap/golden-disk path, in which case the
OS profile carries no authentication fields (the disk already contains
its users; spec-level validation enforces the pairing).

## Key Features

- **Explicit OS choice** — `os_profile` carries exactly one of `linux`
  or `windows`, each with its own authentication contract (SSH-first for
  Linux with password auth disabled by default; password + WinRM/unattend
  for Windows) and its own patch-mode vocabulary, mirroring ARM's
  separate per-OS surfaces. The module deploys the matching ARM resource.
- **Exactly one image source** — marketplace coordinates
  (`source_image_reference`; `latest` resolves at creation only), a
  custom/gallery image ID (`source_image_id`), or an existing OS disk
  (`os_managed_disk_id`).
- **Full compute surface** — spot capacity (eviction policy + max bid),
  placement (zone XOR availability set, proximity placement, capacity
  reservations, dedicated hosts, Flexible-VMSS attach), trusted launch
  (secure boot + vTPM), encryption at host, confidential-VM guest-state
  encryption, platform patch orchestration, boot diagnostics (managed
  storage by default), gallery applications, scheduled-event
  notifications, marketplace plans, per-OS license types (Azure Hybrid
  Benefit / Linux BYOS), Key Vault certificate installation, NVMe disk
  controller, Ultra-SSD/hibernation capabilities, and edge zones.
- **`custom_data` vs `user_data` handled honestly** — cloud-init
  `custom_data` is secret material, delivered once at first boot;
  `user_data` is IMDS-readable and updatable, never secret.
- **Composable** — prerequisites are `AzureResourceGroup` and
  `AzureNetworkInterface`; disks, identities, and vault certificates are
  optional referenced seams.

## Prerequisites

| Kind | Why |
|------|-----|
| `AzureResourceGroup` | The VM is created inside a referenced resource group |
| `AzureNetworkInterface` | At least one referenced NIC gives the VM its network presence |

Optional referenced kinds: `AzureManagedDisk` (data disks or an existing
OS disk), `AzureUserAssignedIdentity` (user-assigned identity),
`AzureKeyVault` (certificate installation).

## Outputs

| Output | Description |
|--------|-------------|
| `vm_id` | ARM ID of the VM (what role assignments, diagnostics, and backup scope to) |
| `vm_name` | VM name within the resource group |
| `virtual_machine_guid` | The 128-bit GUID Azure assigns the VM (licensing/inventory key) |
| `private_ip_address` | Primary private IP across the attached NICs (convenience echo) |
| `public_ip_address` | Primary public IP across the attached NICs (empty for private-only VMs) |
| `computer_name` | The OS hostname the VM booted with |
| `system_assigned_identity_principal_id` | Principal ID of the system-assigned identity (when enabled) — what `AzureRoleAssignment` grants reference |

## Related Resources

- **`AzureNetworkInterface`** — the VM's network presence: subnet
  placement, public IPs, NSG filtering, accelerated networking
- **`AzureManagedDisk`** — first-class data volumes that outlive the VM,
  and the golden-disk boot source
- **`AzureUserAssignedIdentity`** — shareable managed identities attached
  via `identity.identity_ids`
- **`AzureRoleAssignment`** — grants against the VM's system-assigned
  principal or its user-assigned identities
- **`AzureKeyVault`** — certificates installed onto the VM at
  provisioning time (`secrets`)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
