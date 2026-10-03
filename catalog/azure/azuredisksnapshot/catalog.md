# Azure Disk Snapshot

Deploys a managed disk snapshot -- a point-in-time copy of a disk used for backup before risky changes, disk cloning, and as the source of gallery image versions. One resource, two creation modes: "Copy" captures a managed disk (or another snapshot), "Import" wraps a VHD blob already sitting in a storage account. This kind is the single deliberate artifact, not a backup strategy -- scheduled, retention-managed protection belongs to the Recovery Services kinds.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Disk snapshot** -- the point-in-time copy, with its creation mode and source, incremental mode, optional size override, network access posture, and optional legacy ADE encryption settings
- **Azure Tags** -- Planton-derived metadata tags merged with your `tags` map (your values win on key conflicts)

Both engines deliberately ignore in-place edits to the source fields (`sourceResourceId`, `sourceUri`): a snapshot's creation data is immutable history, so editing the source is a no-op rather than a destroy-and-recreate that would silently delete the backup artifact.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A source to capture** -- "Copy" mode needs a managed disk or an existing snapshot (`sourceResourceId`); "Import" mode needs a page-blob VHD's URL (`sourceUri`) and the storage account holding it (`storageAccountId`, which carries the read grant).
- **Same region as the source** -- snapshots are regional; create the snapshot where its source disk lives (`region`).
- **A disk-access resource (only for private posture)** -- `networkAccessPolicy: AllowPrivate` serves the snapshot through a disk-access resource's private endpoint; supply its ARM ID in `diskAccessId` (disk-access resources are not modeled as a Planton kind).

## Deploy

### Console

Open the deployment store, find **Azure Disk Snapshot**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Incremental Disk Backup** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDiskSnapshot
metadata:
  name: orders-db-pre-upgrade
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: acme-prod-rg
  name: orders-db-pre-upgrade
  region: eastus
  createOption: Copy
  sourceResourceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.Compute/disks/orders-db-data
  incrementalEnabled: true
```

```shell
planton apply -f disk-snapshot.yaml
```

This creates an incremental point-in-time copy of the `orders-db-data` disk in `acme-prod-rg`, storing only the delta on standard storage. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the snapshot to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  sourceResourceId:
    valueFrom:
      kind: AzureManagedDisk
      name: orders-db-data
      fieldPath: status.outputs.disk_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and disk first, then captures the snapshot from the resolved disk ID.

## Key Configuration

These are the most important decisions when configuring a disk snapshot. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Creation mode and its source pair** -- `createOption` picks the working pair, and the schema deliberately does not enforce it (the provider's own schema doesn't either -- Azure validates the pairing at create time): "Copy" reads `sourceResourceId`, "Import" reads `sourceUri` plus `storageAccountId`. A manifest with `createOption: Copy` and no source validates offline and then fails at Azure -- only the live create catches a missing source, so get the pair right in the manifest.

**Incremental or full, fixed at creation** -- `incrementalEnabled: true` stores only the delta since the disk's previous snapshot, on standard storage, and is the required form for some consumers such as cross-region copy; full snapshots store the whole disk. The mode is ForceNew -- a chain started full stays full -- so make incremental the reflex and document the exception. One caveat: the first incremental snapshot of a disk stores the full disk; the savings start at the second.

**The source is immutable history** -- both engines ignore in-place edits to `sourceResourceId` and `sourceUri`, because Azure never returns a snapshot's source on reads and a destroy-and-recreate would delete the very backup you took while re-snapshotting the disk's CURRENT state. To capture a different disk -- or a fresh copy of the same disk -- create a NEW snapshot resource with a new name. The same design makes adopting an existing snapshot plan clean: the missing source in state is expected, not drift.

**Network posture: snapshots are exfiltration surfaces** -- the snapshot's data plane supports export (anyone with the right role can generate a download SAS). The defaults (`AllowAll`, public access on) are fine for build artifacts; for snapshots of sensitive disks set `networkAccessPolicy: AllowPrivate` with a `diskAccessId` private endpoint -- or `DenyAll` when nothing should ever export it -- and set `publicNetworkAccessEnabled: false` alongside. All three dials update in place.

**Size inherits from the source** -- leave `diskSizeGb` unset and Azure computes it from the source; set it only to create a LARGER snapshot from a smaller source. It grows in place but never shrinks.

**ADE settings are a one-way door** -- `encryptionSettings` exists only for sources encrypted with legacy in-guest Azure Disk Encryption: it records the Key Vault secret (and optional key-encryption key) a restored disk needs to boot. Platform-managed and customer-managed-key encryption need nothing here. Once set, removing the block destroys and recreates the snapshot -- Azure cannot disable encryption in place.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureManagedDisk** (Copy mode) | `sourceResourceId` | `status.outputs.disk_id` |
| **AzureStorageAccount** (Import mode) | `storageAccountId` | `status.outputs.storage_account_id` |
| **AzureKeyVault** (only with `encryptionSettings`) | `encryptionSettings.diskEncryptionKey.sourceVaultId`, `encryptionSettings.keyEncryptionKey.sourceVaultId` | `status.outputs.key_vault_id` |

`diskAccessId` accepts only a plain ARM ID -- disk-access resources are not modeled as a Planton kind.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `snapshot_id` | The snapshot's Azure Resource Manager ID | A gallery image version's `osDiskSnapshotId` (the golden-image chain), a new AzureManagedDisk's `sourceResourceId` in COPY mode, cross-region copy jobs |
| `snapshot_name` | The snapshot's name | Automation scripts, inventory |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Pre-change backup** -- one deliberate incremental copy of a disk before risky maintenance; a snapshot bills only for the storage it holds, so delete it once the change settles. Start from the **Incremental Disk Backup** preset.

**Golden-image handoff** -- snapshot a prepared OS disk, then publish the `snapshot_id` as a gallery image version. The snapshot is the pipeline's durable artifact between the build VM and the image.

**VHD import on-ramp** -- wrap a page-blob VHD (built on-prem, in another cloud, or exported from another subscription) into a managed snapshot, then clone disks from it. Start from the **Import VHD** preset.

**What this is NOT** -- if you find yourself cron-ing snapshot manifests, you wanted a backup policy: scheduled, retention-managed VM protection belongs to the Recovery Services kinds such as Azure Backup Policy (VM).

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the snapshot lives in
- [**Azure Managed Disk**](/infra-catalog/azure-managed-disk) -- the Copy-mode source, and the consumer that clones new disks from this snapshot's `snapshot_id`
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- holds the VHD blob and carries the read grant for Import mode
- [**Azure Compute Gallery Image**](/infra-catalog/azure-compute-gallery-image) -- builds image versions from this snapshot's `snapshot_id`
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- holds the ADE secrets referenced by `encryptionSettings` for legacy-encrypted sources
- [**Azure Backup Policy (VM)**](/infra-catalog/azure-backup-policy-vm) -- the scheduled, retention-managed alternative when one deliberate copy is not what you wanted
