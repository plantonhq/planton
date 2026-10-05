# AzureVirtualMachineScaleSet Pulumi Module

## Overview

This Pulumi module provisions an Azure Virtual Machine Scale Set using
the Azure Classic provider (`pulumi-azure` v6). One spec, three dispatch
branches: the module reads `orchestration_mode` and the OS profile and
creates exactly one of `compute.OrchestratedVirtualMachineScaleSet`
(FLEXIBLE — the default), `compute.LinuxVirtualMachineScaleSet`, or
`compute.WindowsVirtualMachineScaleSet` (UNIFORM), mirroring how ARM has
one scale-set resource type whose orchestration mode gates its
capabilities.

The template-versus-reference line the spec draws is realized here:
per-instance NICs, disks, and optional public IPs are stamped from
inline templates (they live and die with each instance), while
everything shared — subnets, load-balancer backend pools and NAT rules
(resolved through the load balancer's name-keyed map outputs), NSGs,
user-assigned identities, public IP prefixes, Key Vaults, and the
UNIFORM health probe — arrives as references the platform resolves to
literals before the module runs.

Spec-level validation enforces ARM's conditional matrix (mode gates,
rolling-upgrade pairings, spot pairings, confidential-disk chains), so
the module maps fields without re-validating them. The module is
behaviorally identical to the Terraform module except for one documented
divergence: the pinned SDK bridges the legacy `sku_profile` shape
without per-size ranks, so a PRIORITIZED ranked profile fails loudly
here (see the `PARITY-EXCEPTION` note in `module/flexible.go`) instead
of silently dropping ranks.

Lifecycle notes worth knowing before operating this resource:

- The orchestration mode, OS flavor, and name are the fleet's identity —
  changing any of them replaces the scale set.
- Upgrade mode MANUAL (the default) applies template changes to NEW
  instances only; ROLLING requires health monitoring (the application
  health extension, or the health probe on UNIFORM sets).
- Removing a zone from `zones` replaces the set; adding one does not.
- `single_placement_group` can go true→false but never back.

## Resources Created

- Exactly one of:
  - `compute.OrchestratedVirtualMachineScaleSet` (FLEXIBLE)
  - `compute.LinuxVirtualMachineScaleSet` (UNIFORM + Linux)
  - `compute.WindowsVirtualMachineScaleSet` (UNIFORM + Windows)

## Inputs

The module receives an `AzureVirtualMachineScaleSetIacInput` containing:

- `target.spec.region` / `resource_group` / `name` — the fleet's ARM identity (references resolved to literals by the platform)
- `target.spec.orchestration_mode` — FLEXIBLE (default) or UNIFORM; selects the dispatch branch
- `target.spec.sku_name` / `instances` / `sku_profile` — the size, count, and (FLEXIBLE, `Mix`) mixed-size profile
- `target.spec.os_profile` — exactly one of `linux` (SSH-first auth, optional password, FLEXIBLE patch modes) or `windows` (password, WinRM, unattend content, licensing, FLEXIBLE patch modes + hotpatching)
- `target.spec.os_disk` / `data_disks` — the disk templates (ephemeral OS disks, UltraSSD/PremiumV2 dialed performance, CMK and confidential encryption)
- `target.spec.network_interfaces` — NIC templates with IP configurations: subnet references, LB pool/NAT-rule references (via the LB's name-keyed map outputs), Application Gateway pool IDs, ASG IDs, per-instance public IP templates
- `target.spec.upgrade_policy` — mode, rolling batch contract, UNIFORM automatic OS upgrades, the UNIFORM LB health probe reference
- `target.spec.spot` — eviction policy, max bid, UNIFORM restore, FLEXIBLE priority mix
- `target.spec.automatic_instance_repair` / `termination_notification` / `extensions` / `extensions_time_budget` — the health and lifecycle surface (extension `protected_settings` is secret material; Key-Vault-sourced protected settings are supported)
- `target.spec.identity` — SYSTEM_ASSIGNED (UNIFORM only), USER_ASSIGNED (AzureUserAssignedIdentity references), or both
- `target.spec.security` / `boot_diagnostics` / `secrets` / `plan` / `gallery_applications` / `placement` / `additional_capabilities` — trusted launch and encryption-at-host, diagnostics, Key Vault certificates, marketplace plans, UNIFORM VM applications, placement constraints, Ultra SSD attachability
- `target.spec.zones` / `zone_balance` / `platform_fault_domain_count` — the spreading contract (fault-domain count is REQUIRED on FLEXIBLE sets)
- `target.spec.custom_data` (secret) / `user_data` / `provision_vm_agent` / `extension_operations_enabled` — boot-time data and agent posture
- `target.spec.tags` — user tags, merged over the metadata-derived tags (user wins)
- `provider_config` — Azure credentials, resolved by the shared provider builder (static client secret, keyless web identity, or ambient chain)

Optional enums are mapped to their ARM strings only when explicitly set,
and optional fields with proto defaults are presence-guarded to those
defaults, so an unspecified spec field and Azure's default deploy
identically on both engines.

## Outputs

| Output | Description |
|--------|-------------|
| `scale_set_id` | Full ARM ID of the scale set — what an AzureVirtualMachine's `availability.virtual_machine_scale_set_id` references (FLEXIBLE attach) and what autoscale/monitoring scope to |
| `scale_set_name` | The scale set's name as deployed |
| `unique_id` | The scale set's globally unique ARM-assigned identifier |
| `system_assigned_identity_principal_id` | The system-assigned identity's principal ID (UNIFORM sets with a SYSTEM_ASSIGNED identity); empty otherwise |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
