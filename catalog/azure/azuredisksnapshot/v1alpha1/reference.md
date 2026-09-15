# AzureDiskSnapshot

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureDiskSnapshotSpec** defines a managed disk snapshot -- a
point-in-time copy of a disk used for backup, cloning, and as the
source of gallery image versions
(AzureComputeGalleryImage.versions[].os_disk_snapshot_id).

A snapshot bills only for the storage it holds: INCREMENTAL
snapshots (incremental_enabled) store just the delta since the
previous snapshot of the same disk on standard storage -- the right
default for backup chains; full snapshots store the whole disk.

The provider's own schema does not tie create_option to its source
fields; Azure validates the pairing at create time. The working
pairs: "Copy" reads source_resource_id (a disk or another
snapshot); "Import" reads source_uri (a VHD blob) with
storage_account_id carrying the read grant.

## Example

```yaml
# Deep-shape example for docs and offline validation: an incremental
# Copy-mode snapshot of a managed disk with a private network posture
# and legacy ADE encryption settings carried from the source.
# References are literal values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDiskSnapshot
metadata:
  name: test-disk-snapshot
  id: test-disk-snapshot
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  name: app-disk-snap
  region: eastus
  createOption: Copy
  sourceResourceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Compute/disks/app-disk
  incrementalEnabled: true
  networkAccessPolicy: AllowPrivate
  diskAccessId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Compute/diskAccesses/app-disk-access
  publicNetworkAccessEnabled: false
  encryptionSettings:
    diskEncryptionKey:
      secretUrl: https://app-vault.vault.azure.net/secrets/app-disk-dek/0000000000000000
      sourceVaultId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.KeyVault/vaults/app-vault
    keyEncryptionKey:
      keyUrl: https://app-vault.vault.azure.net/keys/app-disk-kek/0000000000000000
      sourceVaultId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.KeyVault/vaults/app-vault
  tags:
    backupChain: app-disk
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.createOption` | `string` | yes |  |  |
| `spec.sourceResourceId` | `string \| valueFrom` |  |  | AzureManagedDisk (`status.outputs.disk_id`) |
| `spec.sourceUri` | `string` |  |  |  |
| `spec.storageAccountId` | `string \| valueFrom` |  |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.incrementalEnabled` | `bool` |  |  |  |
| `spec.diskSizeGb` | `int32` |  |  |  |
| `spec.networkAccessPolicy` | `string` |  |  |  |
| `spec.diskAccessId` | `string \| valueFrom` |  |  |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  |  |  |
| `spec.encryptionSettings` | `AzureDiskSnapshotEncryptionSettings` |  |  |  |
| `spec.encryptionSettings.diskEncryptionKey` | `AzureDiskSnapshotDiskEncryptionKey` | yes |  |  |
| `spec.encryptionSettings.diskEncryptionKey.secretUrl` | `string` | yes |  |  |
| `spec.encryptionSettings.diskEncryptionKey.sourceVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.encryptionSettings.keyEncryptionKey` | `AzureDiskSnapshotKeyEncryptionKey` |  |  |  |
| `spec.encryptionSettings.keyEncryptionKey.keyUrl` | `string` | yes |  |  |
| `spec.encryptionSettings.keyEncryptionKey.sourceVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the snapshot lives in. Can be a literal
string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the snapshot.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The snapshot's name -- up to 80 characters of letters, numbers,
dashes, and underscores.

**ForceNew**: changing this destroys and recreates the snapshot.

- rule: Snapshot names allow up to 80 letters, numbers, dashes, and underscores
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the snapshot is created in, e.g. "eastus".
Snapshots are regional -- create the snapshot where its source
lives.

**ForceNew**: changing this destroys and recreates the snapshot.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.createOption

`string` · required

How the snapshot is created: "Copy" (from a managed disk or
another snapshot via source_resource_id) or "Import" (from a VHD
blob via source_uri + storage_account_id).

- rule: {"required":true,"string":{"in":["Copy","Import"]}}

### spec.sourceResourceId

`string | valueFrom`

The source managed disk (or snapshot) for create_option "Copy".
Can be a literal ARM ID or a reference to an AzureManagedDisk
output.

**Create-time-only**: a snapshot's creation data is immutable
history, and Azure never returns the source on reads (live-proven
at the v5 provider pin), so both engines deliberately ignore
in-place edits to this field -- editing it does NOT destroy and
recreate the snapshot (that would silently delete a backup
artifact). To capture a different disk, create a NEW snapshot
resource. This is also what makes adopting an existing snapshot
(import) plan clean.

- references: AzureManagedDisk (`status.outputs.disk_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureManagedDisk, name: <that resource's name>, fieldPath: status.outputs.disk_id}} -- a bare string does not parse

### spec.sourceUri

`string`

The source VHD blob URI for create_option "Import".

**Create-time-only**: like source_resource_id, the URI is immutable
creation data Azure never returns on reads -- both engines ignore
in-place edits (no destroy+recreate); importing a different VHD is
a NEW snapshot resource.

### spec.storageAccountId

`string | valueFrom`

The storage account holding source_uri (the read grant for
"Import"). Can be a literal ARM ID or a reference to an
AzureStorageAccount output. The snapshot READS the VHD out of the
account and lives in its own resource group, so the reference is
access, not placement, on a diagram.

**ForceNew**: changing this destroys and recreates the snapshot.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.incrementalEnabled

`bool`

Whether the snapshot is INCREMENTAL: stores only the delta since
the previous snapshot of the same disk, on standard storage --
dramatically cheaper for backup chains, and the required form for
some consumers (e.g. cross-region copy). Gallery image versions
accept both forms.

**ForceNew**: changing this destroys and recreates the snapshot.

### spec.diskSizeGb

`int32` · optional (explicit presence)

The snapshot's size in GB. Unset inherits the source's size
(Azure computes it); set it only to create a LARGER snapshot from
a smaller source. Updatable in place (grow only).

- rule: {"int32":{"gte":1}}

### spec.networkAccessPolicy

`string`

Network access policy for the snapshot's data plane: "AllowAll"
(the provider default when unset), "AllowPrivate" (only through
the disk_access_id private endpoint), or "DenyAll" (no
export/read access at all). Updatable in place.

- rule: {"string":{"in":["","AllowAll","AllowPrivate","DenyAll"]}}

### spec.diskAccessId

`string | valueFrom`

The disk-access resource whose private endpoint serves the
snapshot when network_access_policy is "AllowPrivate". Plain ARM
ID: disk-access resources are not modeled as a Planton kind.
Updatable in place.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the snapshot's data plane accepts public network access.
Unset means the provider default, true. Set false alongside
AllowPrivate/DenyAll for a fully private posture. Updatable in
place.

### spec.encryptionSettings

`AzureDiskSnapshotEncryptionSettings`

Legacy Azure Disk Encryption settings carried over from an
ADE-encrypted source (the Key Vault secret holding the disk
encryption key, and optionally the key-encryption key wrapping
it). Only for sources using in-guest ADE -- platform/CMK
encryption needs no settings here. ONE-WAY: once set, removing
the block destroys and recreates the snapshot (Azure cannot
disable encryption in place).

### spec.encryptionSettings.diskEncryptionKey

`AzureDiskSnapshotDiskEncryptionKey` · required

The Key Vault SECRET holding the disk encryption key.

- rule: {"required":true}

### spec.encryptionSettings.diskEncryptionKey.secretUrl

`string` · required

The secret's URL (the Key Vault secret identifier).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.encryptionSettings.diskEncryptionKey.sourceVaultId

`string | valueFrom` · required

The Key Vault holding the secret. Can be a literal ARM ID or a
reference to an AzureKeyVault output. The snapshot reads its key out
of the vault and lives in its own resource group, so the reference is
access, not placement, on a diagram.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.encryptionSettings.keyEncryptionKey

`AzureDiskSnapshotKeyEncryptionKey`

The Key Vault KEY wrapping the disk encryption key (KEK), when
the source used one.

### spec.encryptionSettings.keyEncryptionKey.keyUrl

`string` · required

The key's URL (the Key Vault key identifier).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.encryptionSettings.keyEncryptionKey.sourceVaultId

`string | valueFrom` · required

The Key Vault holding the key. Can be a literal ARM ID or a
reference to an AzureKeyVault output. Access, not placement, on a
diagram, for the same reason as the disk encryption key's vault.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Tags to apply to the snapshot, merged over the Planton-derived
metadata tags (user values win on key conflicts).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDiskSnapshot, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.snapshot_id` | `string` | The snapshot's Azure Resource Manager ID -- what disks restore from and gallery image versions build from. |
| `status.outputs.snapshot_name` | `string` | The snapshot's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.sourceResourceId` | AzureManagedDisk | `status.outputs.disk_id` |
| `spec.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.encryptionSettings.diskEncryptionKey.sourceVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |
| `spec.encryptionSettings.keyEncryptionKey.sourceVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureComputeGalleryImage | `spec.versions[].osDiskSnapshotId` | `status.outputs.snapshot_id` |

## See Also

- [Overview](../README.md)
