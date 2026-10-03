# Azure Managed Disk

Deploys an Azure Managed Disk — the standalone block storage volume whose data outlives any one virtual machine. The disk is a first-class resource in Azure's own model: it has its own lifecycle, SKU, encryption, and network posture, and a VM **attaches** it rather than containing it. The attachment lives on the VM side (an **AzureVirtualMachine**'s `dataDiskAttachments` reference this disk's `disk_id` output with a LUN and caching mode), so the disk survives VM replacement untouched — and a shared disk (`maxShares`) attaches to several VMs at once, the clustered-database seam only a standalone disk can express.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Managed Disk** -- the block volume with its origin (empty, cloned from a snapshot/disk, stamped from an image, imported from a VHD, restored from a backup recovery point, or a direct-upload target), storage SKU, size, and optional zone pinning
- **Performance configuration** -- fixed per-size tiers on standard/premium SKUs, or independently dialed IOPS/throughput on Premium SSD v2 and Ultra; optional premium tier override and on-demand bursting
- **Encryption posture** -- always encrypted at rest; optionally bound to a referenced AzureDiskEncryptionSet for customer-managed keys, with a second guest-state set slot for confidential-VM customer-key profiles
- **Network export posture** -- who can reach the disk's SAS export endpoint (AllowAll / AllowPrivate via a disk-access resource / DenyAll), plus the public-network dial
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

The VM-side attachment is NOT created here — which VM mounts this disk, at which LUN, with which caching, lives on the referenced AzureVirtualMachine.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the disk will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **For non-EMPTY origins**: the source the origin consumes — a snapshot/disk ARM ID (COPY), an image or gallery version (FROM_IMAGE), a VHD blob and its storage account (IMPORT), or a recovery point (RESTORE).
- **For customer-managed keys**: an AzureDiskEncryptionSet in the SAME region, whose identity holds wrap/unwrap access on the vault key.

## Deploy

### Console

Open the deployment store, find **Azure Managed Disk**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Premium Data Disk** preset in the [Presets](#presets) tab for a zone-pinned empty premium volume.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureManagedDisk
metadata:
  name: orders-db-data
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  name: orders-db-data
  storageAccountType: PREMIUM_LRS
  createOption: EMPTY
  diskSizeGb: 512
  zone: "1"
```

```shell
planton apply -f managed-disk.yaml
```

This creates a 512 GiB Premium SSD data volume pinned to zone 1 — name the data, not the VM: the volume will outlive every machine it attaches to. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the disk to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  diskEncryptionSetId:
    valueFrom:
      kind: AzureDiskEncryptionSet
      name: prod-des
      fieldPath: status.outputs.disk_encryption_set_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and encryption set first, then provisions the disk with the resolved values — and the VM that attaches it deploys after, referencing this disk's `disk_id`.

## Key Configuration

These are the most important decisions when configuring a Managed Disk. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Origin** -- `createOption` is the disk's origin story, fixed at creation: `EMPTY` (a blank volume — set `diskSizeGb`), `COPY` (clone a snapshot or disk via `sourceResourceId`), `FROM_IMAGE` (exactly one of `imageReferenceId` or `galleryImageReferenceId`), `IMPORT`/`IMPORT_SECURE` (wrap a VHD via `sourceUri` + `storageAccountId`), `RESTORE` (materialize a recovery point), or `UPLOAD` (a direct-upload target sized by `uploadSizeBytes`, exact to the byte).

**Storage SKU** -- `storageAccountType` is the fundamental performance/redundancy choice: `STANDARD_LRS` (HDD) for cold data, `STANDARD_SSD_*` for light workloads, `PREMIUM_LRS`/`PREMIUM_ZRS` for production (fixed per-size tiers, credit bursting), and `PREMIUM_V2_LRS`/`ULTRA_SSD_LRS` where size, IOPS (`diskIopsReadWrite`), and throughput (`diskMbpsReadWrite`) are dialed independently. ZRS variants replicate across zones and cannot also be zone-pinned.

**Size** -- `diskSizeGb` (1-65536) only ever INCREASES; growing an attached disk may briefly detach it or deallocate the VM. On fixed-tier SKUs, size buys performance — sometimes the resize IS the performance fix.

**Sharing** -- `maxShares` (2-10) lets several VMs attach the disk simultaneously; the workload must arbitrate writes (WSFC, Pacemaker, or a cluster-aware filesystem). The read-only dial pair budgets read-only mounts on dialed SKUs.

**Encryption** -- always on; `diskEncryptionSetId` binds the disk to your own Key Vault key through a referenced AzureDiskEncryptionSet. Confidential-VM OS disks set `securityType` (the customer-key profile additionally requires `secureVmDiskEncryptionSetId` and a FROM_IMAGE or IMPORT_SECURE origin); `trustedLaunchEnabled` is the boot-hardening alternative — a disk is one or the other.

**Export posture** -- `networkAccessPolicy` governs the SAS export endpoint (not VM attachment): `ALLOW_PRIVATE` pins export to a disk-access resource's private endpoints (`diskAccessId`), `DENY_ALL` disables export entirely.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureDiskEncryptionSet** (optional) | `diskEncryptionSetId`, `secureVmDiskEncryptionSetId` | `status.outputs.disk_encryption_set_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `disk_id` | Azure Resource Manager ID of the disk | AzureVirtualMachine `dataDiskAttachments[].managedDiskId` and `osManagedDiskId` — the attachment seam |
| `disk_name` | Name of the disk | Automation scripts, inventory |
| `disk_size_gb` | The disk's ACTUAL size in GiB — inherited from the source for COPY/FROM_IMAGE disks that omitted a size | Capacity planning, monitoring thresholds |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Premium data disk** -- a zone-pinned empty Premium SSD volume for a database or application data — the common shape. Start from the **Premium Data Disk** preset.

**Dialed performance** -- a Premium SSD v2 volume with IOPS and throughput dialed independently of size: high performance on modest capacity. Start from the **Premium SSD v2 with Dialed Performance** preset.

**Snapshot clone** -- a new disk cloned from a snapshot (restore drills, environment cloning, cross-zone moves), inheriting the source's size. Start from the **Snapshot Clone** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the disk is created
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- attaches this disk by its `disk_id` output (with a LUN and caching mode), or boots from it as a golden OS disk
- [**Azure Disk Encryption Set**](/infra-catalog/azure-disk-encryption-set) -- provides customer-managed keys for encryption at rest and confidential guest state
