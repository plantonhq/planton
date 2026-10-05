# Overview

The **Azure Virtual Machine Scale Set API Resource** provides a consistent and standardized interface for deploying and managing Azure Virtual Machine Scale Sets (VMSS) within our infrastructure. A scale set is a fleet of identical VMs managed as one resource: the instance template (image, size, OS profile, disks, network), the fleet controls (instance count, zones, upgrade policy, spot economics, instance repair), and the orchestration mode.

## Purpose

We developed this API resource to make fleet compute a first-class, composable building block. ARM has exactly one scale-set resource type with an orchestration-mode property, and this kind models it that way — one kind, an explicit mode, and validation that tells the truth about what each mode supports:

- **FLEXIBLE** (the default, and Azure's recommendation for new workloads) spreads instances like a resilient VM group: per-OS patch orchestration, mixed-SKU profiles, spot/on-demand priority mixing, and standalone VMs can attach to the set
- **UNIFORM** is the classic mode behind large stateless fleets: identical instances, overprovisioning, automatic OS image upgrades, spot restore, scale-in policy, gallery applications, and trusted launch

## Key Features

- **Consistent Interface**: Aligns with our existing APIs for deploying cloud infrastructure across multiple providers
- **One Kind, Two Modes**: An explicit `orchestration_mode` discriminator with mode-gated validation — a UNIFORM-only capability on a FLEXIBLE set fails at validation with a message naming the contract, never at deploy time
- **Explicit OS Discriminator**: `os_profile.linux` XOR `os_profile.windows`, each carrying its OS's authentication and management surface (SSH-first for Linux; password, WinRM, unattend content, and hotpatching for Windows)
- **Template vs Reference, Honestly Drawn**: Per-instance NICs and disks are inline templates (they live and die with each instance); everything shared is referenced — subnets, load-balancer pools and NAT rules (by the load balancer's name-keyed map outputs), NSGs, user-assigned identities, public IP prefixes, and Key Vaults
- **ARM-Faithful Validation**: The full conditional matrix is enforced up front — rolling upgrades require a rolling policy and health monitoring, spot fields pair with spot priority, `sku_name: Mix` pairs with a SKU profile, zone balance requires zones, confidential-disk encryption requires vTPM and secure boot
- **Spot Economics**: Eviction policy, max bid price, UNIFORM spot restore, and FLEXIBLE priority mixing (a guaranteed on-demand base plus a spot ratio above it)
- **Production Upgrade Machinery**: Rolling upgrades in health-checked batches, automatic OS image upgrades (UNIFORM), automatic instance repair, and termination notifications

## Use Cases

- **Stateless Web Fleets**: Ephemeral-OS-disk instances behind an AzureLoadBalancer pool, rolling-upgraded in health-checked batches
- **Spot Batch Processing**: Deeply discounted evictable capacity with DELETE eviction and (on FLEXIBLE sets) a priority mix guaranteeing an on-demand base
- **Windows Server Fleets**: UNIFORM fleets with automatic OS upgrades, unattend bootstrap, WinRM management, and Azure Hybrid Benefit licensing
- **Capacity-Resilient Fleets**: FLEXIBLE mixed-SKU profiles (`sku_name: Mix`) that fall back across VM sizes when a region runs dry
- **Network Virtual Appliance Fleets**: IP-forwarding NIC templates with static routes pointed at the fleet
- **Attached Standalone VMs**: A FLEXIBLE set's `scale_set_id` output is what an AzureVirtualMachine's `availability.virtual_machine_scale_set_id` references for scale-set-managed fault spreading of an individually-managed VM

## Future Enhancements

Future updates will include:

- **Autoscale Rules**: A first-class Azure Monitor autoscale resource so instance counts follow demand instead of a fixed number
- **First-Class Dedicated Host Groups and Capacity Reservations**: Placement references currently take plain ARM IDs; dedicated kinds will make them referenceable
- **Referenceable Application Gateway Pools**: Application Gateway pool memberships currently take plain ARM IDs; they become referenceable once the Application Gateway exports per-pool IDs
- **Comprehensive Documentation**: Expanded guidance for image lifecycle management with Shared Image Galleries

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
