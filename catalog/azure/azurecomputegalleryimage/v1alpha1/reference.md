# AzureComputeGalleryImage

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureComputeGalleryImageSpec** defines a gallery image: one image
definition inside an Azure Compute Gallery (its marketplace-style
identity, OS type, security posture, and recommended sizing) plus
the published versions of that image, each replicated to its own
target regions. VMs and scale sets deploy from a version's ARM ID
({image_id}/versions/{version}) or from the definition's ID to get
the latest version.

The definition itself is free; each version bills for the storage
its regional replicas consume. Versions are the image owner's own
release artifacts and live in the definition's manifest -- remove a
version entry to unpublish it.

## Example

```yaml
# Deep-shape example for docs and offline validation: a Gen2
# trusted-launch Linux image with recommended sizing, disk-type
# exclusions, and one snapshot-sourced version replicated to two
# regions (one with a customer-managed-key disk encryption set).
# References are literal values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureComputeGalleryImage
metadata:
  name: test-gallery-image
  id: test-gallery-image
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  galleryName:
    value: platform.images
  name: ubuntu-base
  region: eastus
  identifier:
    publisher: acme
    offer: ubuntu
    sku: 22-04-lts-gen2
  osType: Linux
  hyperVGeneration: V2
  trustedLaunchEnabled: true
  acceleratedNetworkSupportEnabled: true
  diskTypesNotAllowed:
    - Standard_LRS
  minRecommendedVcpuCount: 2
  maxRecommendedVcpuCount: 16
  minRecommendedMemoryInGb: 4
  maxRecommendedMemoryInGb: 64
  endOfLifeDate: "2030-01-01T00:00:00Z"
  releaseNoteUri: https://acme.example/images/ubuntu-base/releases
  description: Hardened Ubuntu 22.04 Gen2 base image
  versions:
    - name: 1.0.0
      osDiskSnapshotId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Compute/snapshots/ubuntu-base-1-0-0
      replicationMode: Full
      targetRegions:
        - name: eastus
          regionalReplicaCount: 2
          storageAccountType: Standard_ZRS
        - name: westeurope
          regionalReplicaCount: 1
          diskEncryptionSetId:
            value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Compute/diskEncryptionSets/images-des
      tags:
        release: stable
  tags:
    costCenter: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.galleryName` | `string \| valueFrom` | yes |  | AzureComputeGallery (`status.outputs.gallery_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.identifier` | `AzureComputeGalleryImageIdentifier` | yes |  |  |
| `spec.identifier.publisher` | `string` | yes |  |  |
| `spec.identifier.offer` | `string` | yes |  |  |
| `spec.identifier.sku` | `string` | yes |  |  |
| `spec.osType` | `string` | yes |  |  |
| `spec.specialized` | `bool` |  |  |  |
| `spec.architecture` | `string` |  |  |  |
| `spec.hyperVGeneration` | `string` |  |  |  |
| `spec.trustedLaunchSupported` | `bool` |  |  |  |
| `spec.trustedLaunchEnabled` | `bool` |  |  |  |
| `spec.confidentialVmSupported` | `bool` |  |  |  |
| `spec.confidentialVmEnabled` | `bool` |  |  |  |
| `spec.acceleratedNetworkSupportEnabled` | `bool` |  |  |  |
| `spec.hibernationEnabled` | `bool` |  |  |  |
| `spec.diskControllerTypeNvmeEnabled` | `bool` |  |  |  |
| `spec.diskTypesNotAllowed` | `[]string` |  |  |  |
| `spec.endOfLifeDate` | `string` |  |  |  |
| `spec.eula` | `string` |  |  |  |
| `spec.privacyStatementUri` | `string` |  |  |  |
| `spec.releaseNoteUri` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.purchasePlan` | `AzureComputeGalleryImagePurchasePlan` |  |  |  |
| `spec.purchasePlan.name` | `string` | yes |  |  |
| `spec.purchasePlan.publisher` | `string` |  |  |  |
| `spec.purchasePlan.product` | `string` |  |  |  |
| `spec.minRecommendedVcpuCount` | `int32` |  |  |  |
| `spec.maxRecommendedVcpuCount` | `int32` |  |  |  |
| `spec.minRecommendedMemoryInGb` | `int32` |  |  |  |
| `spec.maxRecommendedMemoryInGb` | `int32` |  |  |  |
| `spec.versions` | `[]AzureComputeGalleryImageVersion` |  |  |  |
| `spec.versions[].name` | `string` | yes |  |  |
| `spec.versions[].targetRegions` | `[]AzureComputeGalleryImageVersionTargetRegion` | yes |  |  |
| `spec.versions[].targetRegions[].name` | `string` | yes |  |  |
| `spec.versions[].targetRegions[].regionalReplicaCount` | `int32` | yes |  |  |
| `spec.versions[].targetRegions[].diskEncryptionSetId` | `string \| valueFrom` |  |  | AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`) |
| `spec.versions[].targetRegions[].excludeFromLatestEnabled` | `bool` |  |  |  |
| `spec.versions[].targetRegions[].storageAccountType` | `string` |  |  |  |
| `spec.versions[].blobUri` | `string` |  |  |  |
| `spec.versions[].storageAccountId` | `string \| valueFrom` |  |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.versions[].osDiskSnapshotId` | `string \| valueFrom` |  |  | AzureDiskSnapshot (`status.outputs.snapshot_id`) |
| `spec.versions[].managedImageId` | `string \| valueFrom` |  |  |  |
| `spec.versions[].replicationMode` | `string` |  |  |  |
| `spec.versions[].excludeFromLatest` | `bool` |  |  |  |
| `spec.versions[].deletionOfReplicatedLocationsEnabled` | `bool` |  |  |  |
| `spec.versions[].endOfLifeDate` | `string` |  |  |  |
| `spec.versions[].tags` | `map<string, string>` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the image's gallery lives in. Can be a
literal string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the image.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.galleryName

`string | valueFrom` · required

The Compute Gallery the image definition lives in. Can be a
literal gallery name or a reference to an AzureComputeGallery
output.

**ForceNew**: changing this destroys and recreates the image.

- references: AzureComputeGallery (`status.outputs.gallery_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureComputeGallery, name: <that resource's name>, fieldPath: status.outputs.gallery_name}} -- a bare string does not parse

### spec.name

`string` · required

The image definition's name -- up to 80 characters of letters,
numbers, dots, dashes, and underscores.

**ForceNew**: changing this destroys and recreates the image.

- rule: Image names allow up to 80 letters, numbers, dots, dashes, and underscores
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the image definition is created in, e.g.
"eastus". Each version replicates to its own target regions; the
definition's region is where the metadata lives.

**ForceNew**: changing this destroys and recreates the image.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.identifier

`AzureComputeGalleryImageIdentifier` · required

The image's marketplace-style identity. The publisher/offer/SKU
triple must be unique within the gallery; VMs match images by it.

**ForceNew**: changing any part destroys and recreates the image.

- rule: {"required":true}

### spec.identifier.publisher

`string` · required

The image publisher (up to 128 characters of letters, numbers,
dots, dashes, and underscores; must not end with a dot).

- rule: publisher allows up to 128 letters, numbers, dots, dashes, and underscores, and must not end with a dot
- rule: {"required":true}

### spec.identifier.offer

`string` · required

The image offer (up to 64 characters, same character rules as
publisher).

- rule: offer allows up to 64 letters, numbers, dots, dashes, and underscores, and must not end with a dot
- rule: {"required":true}

### spec.identifier.sku

`string` · required

The image SKU (up to 64 characters, same character rules as
publisher).

- rule: sku allows up to 64 letters, numbers, dots, dashes, and underscores, and must not end with a dot
- rule: {"required":true}

### spec.osType

`string` · required

The operating system of every version of this image: "Linux" or
"Windows".

**ForceNew**: changing this destroys and recreates the image.

- rule: {"required":true,"string":{"in":["Linux","Windows"]}}

### spec.specialized

`bool`

Whether versions of this image are SPECIALIZED (carry machine
identity and user accounts from their source, deploy as clones)
rather than generalized (sysprepped/deprovisioned, deploy as
fresh machines). Most golden images are generalized -- leave this
false unless the workflow clones prepared machines.

**ForceNew**: changing this destroys and recreates the image.

### spec.architecture

`string`

The CPU architecture of the image: "x64" or "Arm64". Unset means
the provider default, "x64".

**ForceNew**: changing this destroys and recreates the image.

- rule: {"string":{"in":["","x64","Arm64"]}}

### spec.hyperVGeneration

`string`

The Hyper-V generation VMs deploying this image use: "V1" (BIOS
boot) or "V2" (UEFI -- required for trusted launch and
confidential VMs). Unset means the provider default, "V1"; prefer
"V2" for new images.

**ForceNew**: changing this destroys and recreates the image.

- rule: {"string":{"in":["","V1","V2"]}}

### spec.trustedLaunchSupported

`bool`

Marks the image as SUPPORTING trusted launch: consumers may
deploy it with or without trusted launch. At most one of the four
security flags can be set -- they are mutually exclusive security
postures.

**ForceNew**: changing this destroys and recreates the image.

### spec.trustedLaunchEnabled

`bool`

Marks the image as REQUIRING trusted launch (secure boot + vTPM)
on every VM deployed from it. Requires Hyper-V generation V2. At
most one of the four security flags can be set.

**ForceNew**: changing this destroys and recreates the image.

### spec.confidentialVmSupported

`bool`

Marks the image as SUPPORTING confidential VMs: consumers may
deploy it as a confidential VM. At most one of the four security
flags can be set.

**ForceNew**: changing this destroys and recreates the image.

### spec.confidentialVmEnabled

`bool`

Marks the image as REQUIRING confidential-VM deployment. At most
one of the four security flags can be set.

**ForceNew**: changing this destroys and recreates the image.

### spec.acceleratedNetworkSupportEnabled

`bool`

Marks the image as supporting accelerated networking; consumers
see it reflected on the image definition.

**ForceNew**: changing this destroys and recreates the image.

### spec.hibernationEnabled

`bool`

Marks the image as supporting hibernation on VMs deployed from
it.

**ForceNew**: changing this destroys and recreates the image.

### spec.diskControllerTypeNvmeEnabled

`bool`

Marks the image as supporting the NVMe disk controller (in
addition to SCSI) on VMs deployed from it.

**ForceNew**: changing this destroys and recreates the image.

### spec.diskTypesNotAllowed

`[]string`

Disk storage types VMs deploying this image may NOT use:
"Standard_LRS" and/or "Premium_LRS". Updatable in place.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["Standard_LRS","Premium_LRS"]}}}}

### spec.endOfLifeDate

`string`

When the image definition reaches end of life (RFC3339, e.g.
"2030-01-01T00:00:00Z") -- an advisory signal to consumers, not
an enforcement. Updatable in place; CLEARING a previously set
date destroys and recreates the image (the provider forces
replacement on removal).

- rule: end_of_life_date must be an RFC3339 timestamp, e.g. "2030-01-01T00:00:00Z"

### spec.eula

`string`

The end-user license agreement text or URL for the image.

**ForceNew**: changing this destroys and recreates the image.

### spec.privacyStatementUri

`string`

A privacy statement URI shown to image consumers.

**ForceNew**: changing this destroys and recreates the image.

### spec.releaseNoteUri

`string`

A release-notes URI for the image. Updatable in place.

### spec.description

`string`

A description of the image shown in the portal and returned by
the API. Updatable in place.

### spec.purchasePlan

`AzureComputeGalleryImagePurchasePlan`

The marketplace purchase plan the image's source carries, if any
(bring-your-own-subscription images built from plan-bearing
marketplace sources must declare it).

**ForceNew**: changing any part destroys and recreates the image.

### spec.purchasePlan.name

`string` · required

The purchase plan's name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.purchasePlan.publisher

`string`

The purchase plan's publisher.

### spec.purchasePlan.product

`string`

The purchase plan's product.

### spec.minRecommendedVcpuCount

`int32` · optional (explicit presence)

The minimum vCPU count recommended for VMs deploying this image
(1-80). Advisory sizing guidance shown to consumers. Updatable in
place.

- rule: {"int32":{"lte":80,"gte":1}}

### spec.maxRecommendedVcpuCount

`int32` · optional (explicit presence)

The maximum vCPU count recommended for VMs deploying this image
(1-80); must be >= the minimum when both are set. Updatable in
place.

- rule: {"int32":{"lte":80,"gte":1}}

### spec.minRecommendedMemoryInGb

`int32` · optional (explicit presence)

The minimum memory in GB recommended for VMs deploying this image
(1-640). Updatable in place.

- rule: {"int32":{"lte":640,"gte":1}}

### spec.maxRecommendedMemoryInGb

`int32` · optional (explicit presence)

The maximum memory in GB recommended for VMs deploying this image
(1-640); must be >= the minimum when both are set. Updatable in
place.

- rule: {"int32":{"lte":640,"gte":1}}

### spec.versions

`[]AzureComputeGalleryImageVersion`

The image's published versions -- the release artifacts VMs
actually deploy. Each version names its source (exactly one of a
blob, a disk snapshot, or a managed image/VM) and the regions it
replicates to. Versions share the definition's namespace: names
must be unique. Add an entry to publish a version; remove it to
unpublish.

- rule: exactly one of blob_uri, os_disk_snapshot_id, and managed_image_id must be set -- the version has one source
- rule: blob_uri and storage_account_id are set together -- the blob's storage account carries the read grant
- rule: disk_encryption_set_id cannot be used in target_regions when replication_mode is Shallow

### spec.versions[].name

`string` · required

The version's name: three dot-separated numeric segments (e.g.
"1.2.0", each segment up to 10 digits), or the literal "latest"
or "recent" (the provider's own validator admits all three
forms).

**ForceNew**: changing this destroys and recreates the version.

- rule: Version names are three dot-separated numeric segments (e.g. "1.2.0"), or the literal "latest" or "recent"
- rule: {"required":true}

### spec.versions[].targetRegions

`[]AzureComputeGalleryImageVersionTargetRegion` · required

The regions this version replicates to, each with its own replica
count and storage type. At least one region is required; the
version is deployable in exactly these regions. Regions can be
added and removed in place.

- rule: {"repeated":{"minItems":"1"}}

### spec.versions[].targetRegions[].name

`string` · required

The region's name, e.g. "eastus". Updatable: adding and removing
regions replicates and unreplicates in place.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.versions[].targetRegions[].regionalReplicaCount

`int32` · required

How many replicas of the version this region keeps (throughput
scaling for concurrent VM deployments). Updatable in place.

- rule: {"required":true,"int32":{"gte":1}}

### spec.versions[].targetRegions[].diskEncryptionSetId

`string | valueFrom`

Encrypts this region's replicas with a customer-managed key
through a disk encryption set. Can be a literal ARM ID or a
reference to an AzureDiskEncryptionSet output. Not allowed with
Shallow replication.

**ForceNew**: changing this destroys and recreates the version.

- references: AzureDiskEncryptionSet (`status.outputs.disk_encryption_set_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDiskEncryptionSet, name: <that resource's name>, fieldPath: status.outputs.disk_encryption_set_id}} -- a bare string does not parse

### spec.versions[].targetRegions[].excludeFromLatestEnabled

`bool`

Excludes this REGION from "latest" resolution (region-scoped
sibling of the version-level exclude_from_latest). Updatable in
place.

### spec.versions[].targetRegions[].storageAccountType

`string`

The storage account type for this region's replicas:
"Standard_LRS" (the provider default when unset), "Premium_LRS",
or "Standard_ZRS". Azure cannot UPDATE this in place and the
provider cannot force replacement for it (region-list membership
changes in place) -- changing it on an existing region surfaces
Azure's own error; remove and re-add the region instead.

- rule: {"string":{"in":["","Premium_LRS","Standard_LRS","Standard_ZRS"]}}

### spec.versions[].blobUri

`string`

Source: a page blob (VHD) URI. Exactly one of blob_uri,
os_disk_snapshot_id, and managed_image_id must be set; blob_uri
requires storage_account_id.

**ForceNew**: changing this destroys and recreates the version.

- rule: blob_uri must be an http or https URL

### spec.versions[].storageAccountId

`string | valueFrom`

The storage account holding blob_uri. Required with blob_uri,
forbidden otherwise. Can be a literal ARM ID or a reference to an
AzureStorageAccount output. The version is BUILT from the blob and
lives in its gallery, so the reference is access, not placement, on a
diagram.

**ForceNew**: changing this destroys and recreates the version.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.versions[].osDiskSnapshotId

`string | valueFrom`

Source: an OS disk snapshot. Exactly one of blob_uri,
os_disk_snapshot_id, and managed_image_id must be set. Can be a
literal ARM ID or a reference to an AzureDiskSnapshot output.

**ForceNew**: changing this destroys and recreates the version.

- references: AzureDiskSnapshot (`status.outputs.snapshot_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDiskSnapshot, name: <that resource's name>, fieldPath: status.outputs.snapshot_id}} -- a bare string does not parse

### spec.versions[].managedImageId

`string | valueFrom`

Source: a legacy managed image's ARM ID, or a VM's ARM ID to
capture directly. Exactly one of blob_uri, os_disk_snapshot_id,
and managed_image_id must be set. Plain ARM ID: legacy managed
images are superseded by the gallery itself and are not modeled
as a Planton kind.

**ForceNew**: changing this destroys and recreates the version.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.versions[].replicationMode

`string`

The replication mode: "Full" (durable copies per region -- the
provider default when unset) or "Shallow" (references the source
directly: instant publish, source must live on; used for dev/test
and very large images). Shallow versions cannot use per-region
disk encryption sets.

**ForceNew**: changing this destroys and recreates the version.

- rule: {"string":{"in":["","Full","Shallow"]}}

### spec.versions[].excludeFromLatest

`bool`

Excludes this version from "latest" resolution: consumers pinned
to the definition's latest version skip it (e.g. while it bakes
or after a regression). Updatable in place.

### spec.versions[].deletionOfReplicatedLocationsEnabled

`bool`

Whether deleting this version is allowed to also delete its
replicated copies in all target regions.

**ForceNew**: changing this destroys and recreates the version.

### spec.versions[].endOfLifeDate

`string`

When this version reaches end of life (RFC3339) -- advisory to
consumers. Updatable in place; CLEARING a previously set date
destroys and recreates the version (the provider forces
replacement on removal).

- rule: end_of_life_date must be an RFC3339 timestamp, e.g. "2030-01-01T00:00:00Z"

### spec.versions[].tags

`map<string, string>`

Tags to apply to the version, merged over the Planton-derived
metadata tags (user values win on key conflicts).

### spec.tags

`map<string, string>`

Tags to apply to the image definition, merged over the
Planton-derived metadata tags (user values win on key conflicts).

## Validation Rules

- `gallery_image_security_flags_exclusive`: at most one of trusted_launch_supported, trusted_launch_enabled, confidential_vm_supported, and confidential_vm_enabled can be set -- they are mutually exclusive security postures
- `gallery_image_vcpu_range_ordered`: max_recommended_vcpu_count must be greater than or equal to min_recommended_vcpu_count
- `gallery_image_memory_range_ordered`: max_recommended_memory_in_gb must be greater than or equal to min_recommended_memory_in_gb
- `gallery_image_version_names_unique`: versions names must be unique -- the name is the version's identity under the image

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureComputeGalleryImage, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.image_id` | `string` | The image definition's Azure Resource Manager ID. VMs deploying from it get the image's latest (non-excluded) version. |
| `status.outputs.image_name` | `string` | The image definition's name within its gallery. |
| `status.outputs.version_ids` | `map<string, string>` | The ARM IDs of the image's published versions, keyed by version name ({image_id}/versions/{name}). VMs pin to an exact release through these. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.galleryName` | AzureComputeGallery | `status.outputs.gallery_name` |
| `spec.versions[].targetRegions[].diskEncryptionSetId` | AzureDiskEncryptionSet | `status.outputs.disk_encryption_set_id` |
| `spec.versions[].storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.versions[].osDiskSnapshotId` | AzureDiskSnapshot | `status.outputs.snapshot_id` |

## See Also

- [Overview](../README.md)
