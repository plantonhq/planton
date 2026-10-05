# Azure Compute Gallery Image

Deploys a gallery image: one image definition inside an Azure Compute Gallery plus its published, region-replicated versions -- the release artifacts VMs actually deploy. The definition carries the image's permanent marketplace-style identity (publisher/offer/SKU), OS type, security posture, and recommended sizing; each version in the `versions` list names exactly one source (a disk snapshot, a VHD blob, or a managed image/VM) and the regions it replicates to. The definition is free; versions bill for the storage their regional replicas consume.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Image definition** -- the image's identity, OS type, architecture, Hyper-V generation, security posture, and advisory sizing, inside the referenced gallery
- **Image versions** -- one per entry in the `versions` list, each built from its declared source and replicated to its target regions with per-region replica counts, storage types, and optional customer-managed-key encryption
- **Azure Tags** -- your governance tags merged over the Planton-derived resource tags on both the definition and each version; a user tag with the same key wins

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Compute Gallery** -- referenced through `galleryName` as a literal name or an AzureComputeGallery ValueFromRef; the definition lives inside it.
- **A version source** (only when publishing versions) -- exactly one per version: an OS disk snapshot (`osDiskSnapshotId`, the chart-native chain), a VHD page blob (`blobUri` plus its `storageAccountId`), or a managed image or VM ARM ID (`managedImageId`).
- **A disk encryption set** (only for customer-managed-key replicas) -- referenced per target region through `diskEncryptionSetId`; not allowed with Shallow replication.

## Deploy

### Console

Open the deployment store, find **Azure Compute Gallery Image**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Linux Gen2 Trusted Launch** or **Image With Version** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureComputeGalleryImage
metadata:
  name: ubuntu-golden
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  galleryName:
    value: "platform.images"
  name: ubuntu-22-04
  region: eastus
  identifier:
    publisher: acme
    offer: ubuntu
    sku: 22-04-lts-gen2
  osType: Linux
  hyperVGeneration: V2
  trustedLaunchSupported: true
```

```shell
planton apply -f gallery-image.yaml
```

This registers the image definition -- Gen2 Linux with trusted launch supported, no versions published yet -- ready for the image pipeline's first release to land in the `versions` list. An Infra Job tracks the provisioning in real time.

### InfraChart

When a chart carries the whole image chain -- gallery, snapshot, definition, version -- wire it by reference:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  galleryName:
    valueFrom:
      kind: AzureComputeGallery
      name: platform-images
      fieldPath: status.outputs.gallery_name
  name: ubuntu-22-04
  region: eastus
  identifier:
    publisher: acme
    offer: ubuntu
    sku: 22-04-lts-gen2
  osType: Linux
  hyperVGeneration: V2
  trustedLaunchSupported: true
  versions:
    - name: "1.0.0"
      osDiskSnapshotId:
        valueFrom:
          kind: AzureDiskSnapshot
          name: golden-os-snapshot
          fieldPath: status.outputs.snapshot_id
      targetRegions:
        - name: eastus
          regionalReplicaCount: 1
```

The InfraPipeline resolves the dependency graph, deploys the gallery and snapshot first, then creates the definition and publishes the version from the resolved snapshot.

## Key Configuration

These are the most important decisions when configuring a gallery image. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The identity is forever; get Gen2 and trusted launch right on day one** -- Publisher/offer/SKU, OS type, Hyper-V generation, and the security flags are all create-only: changing any of them replaces the definition and orphans every published version. The posture that ages best for new fleets: `hyperVGeneration: V2` with `trustedLaunchSupported: true` -- consumers CHOOSE trusted launch. `trustedLaunchEnabled` FORCES it on every consumer forever; pick it only when compliance demands it. All four security flags (trusted launch supported/enabled, confidential VM supported/enabled) are mutually exclusive.

**Treat versions as immutable releases** -- A version's name, source, and replication mode are create-only; only its target regions, exclude-from-latest, end-of-life date, and tags move in place. The workflow that stays honest: the image pipeline publishes a NEW semver-named version per build, `excludeFromLatest` quarantines a bad release instantly without unpublishing it, and removing the entry unpublishes it once nothing pins it. Never rebuild a version in place under the same name.

**Replicas are deployment throughput, not durability** -- `regionalReplicaCount` sizes concurrent VM creation from that region's copy -- roughly one replica per 20 concurrent creations, more for scale-set bursts. One replica per region is fine for trickle deployments. Every replica bills storage in its region: prune target regions no fleet deploys from. Per-region `storageAccountType` is create-only in practice -- the API cannot update it in place; remove and re-add the region to change it.

**Shallow replication is a dev-loop tool with a leash** -- `replicationMode: Shallow` publishes instantly by referencing the source instead of copying it -- excellent for the image-bake inner loop and very large images. The leash: the source snapshot or blob must outlive the version (deleting it breaks deploys), Shallow versions cannot use per-region disk encryption sets, and the replica count is effectively 1. Production releases use Full, always.

**Specialized is a different product** -- `specialized: true` marks every version as carrying machine identity and user accounts from its source, deploying as clones rather than fresh machines. Most golden images are generalized (sysprepped or deprovisioned) -- leave it false unless the workflow deliberately clones prepared machines. It is create-only.

**Clearing end-of-life dates replaces the resource** -- `endOfLifeDate` (on the definition and per version) is advisory and updates in place -- but clearing a previously set date forces replacement; the provider cannot express "remove this property" to ARM. If a retirement date moves, set a new date rather than clearing it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureComputeGallery** | `galleryName` | `status.outputs.gallery_name` |
| **AzureDiskSnapshot** (snapshot-sourced versions) | `versions[].osDiskSnapshotId` | `status.outputs.snapshot_id` |
| **AzureStorageAccount** (blob-sourced versions) | `versions[].storageAccountId` | `status.outputs.storage_account_id` |
| **AzureDiskEncryptionSet** (optional, per region) | `versions[].targetRegions[].diskEncryptionSetId` | `status.outputs.disk_encryption_set_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `image_id` | The definition's ARM ID -- deploying from it resolves the latest non-excluded version | VM image references for environments that should float on the newest release |
| `version_ids` | ARM IDs of the published versions, keyed by version name | VM image references pinning production to an exact release |

The other output, `image_name`, echoes the definition's name within its gallery.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Register the identity first** -- create the definition (identity, Gen2, trusted launch supported) before the first build exists, so the pipeline has a stable target to publish into. Start from the **Linux Gen2 Trusted Launch** preset.

**Snapshot-sourced releases** -- the chart-native release chain: a prepared managed disk, an AzureDiskSnapshot of it, and a semver-named version entry sourcing that snapshot. Production VMs pin exact `version_ids`; development floats on `image_id` (latest). Start from the **Image With Version** preset.

**Quarantine, don't unpublish** -- when a release regresses, set `excludeFromLatest: true` on it: floaters stop picking it up immediately, while VMs pinned to it keep working until they migrate. Remove the version entry only when nothing references it.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the gallery and image live in
- [**Azure Compute Gallery**](/infra-catalog/azure-compute-gallery) -- the gallery the definition lives in, referenced by its `gallery_name` output
- [**Azure Disk Snapshot**](/infra-catalog/azure-disk-snapshot) -- the prepared-OS-disk source for snapshot-built versions
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- holds the VHD page blob for blob-sourced versions
- [**Azure Disk Encryption Set**](/infra-catalog/azure-disk-encryption-set) -- customer-managed-key encryption for a region's replicas
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- deploys from a pinned version ID or the definition's latest
