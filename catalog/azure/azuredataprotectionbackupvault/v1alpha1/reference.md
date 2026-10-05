# AzureDataProtectionBackupVault

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureDataProtectionBackupVaultSpec** defines a Data Protection
backup vault (ARM: Microsoft.DataProtection/backupVaults) -- the
safe that MODERN Azure Backup data lives in: managed disks, blob
storage, AKS clusters, MySQL/PostgreSQL flexible servers and Data
Lake storage. Backup policies and backup instances are ARM children
of a vault. This is the newer generation alongside the classic
Recovery Services vault (which serves VM and file-share backup).

A vault is FREE at rest -- cost accrues per protected instance and
per GB of backup storage, not for the vault object itself.

**Three settings are one-way doors** (the provider replaces the
vault to walk them back): cross-region restore once enabled,
immutability once Locked, and soft delete once AlwaysOn. Each is
documented on its field.

**Deletion outlives the API's answer**: Azure's delete call returns
before the vault is fully gone, so the provider polls until the
name is actually free (its own workaround for the service bug).
Expect destroy to take a little longer than the API suggests.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: geo-redundant
# storage with cross-region restore, a widened soft-delete window, the
# Unlocked immutability posture, a system-assigned identity, the
# composed customer-managed-key arm with the versionless key URI, and
# user tags merged over the derived ones. The one-way doors (Locked,
# AlwaysOn) stay offline-only by design -- they are permanent on a live
# vault.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataProtectionBackupVault
metadata:
  name: test-data-protection-backup-vault
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: test-backup-vault
  datastoreType: VaultStore
  redundancy: GeoRedundant
  crossRegionRestoreEnabled: true
  retentionDurationInDays: 30
  softDelete: "On"
  immutability: Unlocked
  identity:
    type: SYSTEM_ASSIGNED
  encryption:
    keyId:
      value: https://test-kv.vault.azure.net/keys/vault-cmk
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.datastoreType` | `string` | yes |  |  |
| `spec.redundancy` | `string` | yes |  |  |
| `spec.crossRegionRestoreEnabled` | `bool` |  |  |  |
| `spec.retentionDurationInDays` | `int32` |  | `14` |  |
| `spec.softDelete` | `string` |  | `On` |  |
| `spec.immutability` | `string` |  | `Disabled` |  |
| `spec.identity` | `AzureDataProtectionBackupVaultIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.encryption` | `AzureDataProtectionBackupVaultEncryption` |  |  |  |
| `spec.encryption.keyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the vault lives in, e.g. "eastus". Backup data
is stored in this region (and its pair, under geo-redundant
storage). Changing the region replaces the vault.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the vault is created in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The vault's name, unique within its resource group: 2-50
characters, letters, digits and hyphens (the provider's own
rule). Backup policies and backup instances address the vault
through its ARM ID. Changing the name replaces the vault.

- rule: {"required":true,"string":{"pattern":"^[-a-zA-Z0-9]{2,50}$"}}

### spec.datastoreType

`string` · required

The datastore tier backups in this vault land on (the wire
values). "VaultStore" is the standard tier and the right choice
for almost everything; "OperationalStore" serves
operational-tier-only datasources; "ArchiveStore" is the
long-term cold tier. Fixed at creation -- changing it replaces
the vault. NOTE: policies choose per-rule stores independently;
this setting is the vault's storage-settings tier.

- rule: {"required":true,"string":{"in":["VaultStore","OperationalStore","ArchiveStore"]}}

### spec.redundancy

`string` · required

How the vault's backup storage replicates (the wire values).
GeoRedundant keeps a copy in the paired region (the production
posture, and the only redundancy cross-region restore works
with); LocallyRedundant is the cheapest; ZoneRedundant replicates
across availability zones in-region. Fixed at creation --
changing it replaces the vault.

- rule: {"required":true,"string":{"in":["GeoRedundant","LocallyRedundant","ZoneRedundant"]}}

### spec.crossRegionRestoreEnabled

`bool`

Whether backup data can be restored in the paired region
(cross-region restore). Requires GeoRedundant redundancy.
Enabling is an in-place update; ONCE ENABLED IT CANNOT BE
DISABLED without replacing the vault (the provider's one-way
ForceNew).

### spec.retentionDurationInDays

`int32` · optional (explicit presence)

How many days soft-deleted backup data is retained before
permanent removal (this is the SOFT-DELETE retention window --
how long backups survive a deletion -- NOT how long backups are
kept; backup retention lives on policies). 14-180 days.
Unspecified applies 14 (the provider's default).

- default: `14`
- rule: {"int32":{"lte":180,"gte":14}}

### spec.softDelete

`string` · optional (explicit presence)

The vault's soft-delete posture (the wire values). "On" (the
default) retains deleted backup data for the retention window;
"AlwaysOn" locks soft delete on PERMANENTLY -- leaving AlwaysOn
replaces the vault (the provider's one-way ForceNew); "Off"
disables it (privileged when a Resource Guard governs the vault).
Unspecified applies On.

- default: `On`
- rule: {"string":{"in":["On","Off","AlwaysOn"]}}

### spec.immutability

`string` · optional (explicit presence)

The vault's immutability posture (the wire values). Immutability
stops backup data being deleted or retention being reduced --
ransomware protection. Disabled <-> Unlocked moves freely; LOCKED
IS PERMANENT (leaving Locked replaces the vault -- the provider's
one-way ForceNew). Unspecified applies Disabled.

- default: `Disabled`
- rule: {"string":{"in":["Disabled","Unlocked","Locked"]}}

### spec.identity

`AzureDataProtectionBackupVaultIdentity`

The vault's managed identity -- required for customer-managed-key
encryption (the key is always unwrapped by the SYSTEM-assigned
identity) and for granting the vault access to datasources it
backs up.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the vault (required for customer-managed-key encryption);
USER_ASSIGNED brings identities you manage;
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_data_protection_backup_vault_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the vault.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the vault, by ARM ID. Reference
AzureUserAssignedIdentity resources so access grants can be
composed before the vault is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.encryption

`AzureDataProtectionBackupVaultEncryption`

Encryption of backup data with your OWN Key Vault key
(customer-managed key) instead of Microsoft-managed keys.
Requires a system-assigned identity with wrap/unwrap access on
the key. ONCE ENABLED IT CAN NEVER BE REMOVED (Azure's own
contract -- the provider's delete is a documented no-op; only
deleting the vault removes CMK). The key itself CAN be rotated in
place.

### spec.encryption.keyId

`string | valueFrom` · required

The Key Vault KEY that encrypts the vault's backup data.
Versionless and versioned key URIs are BOTH accepted by the
provider; the versionless form (this reference's default) is the
right choice -- key rotation then propagates automatically,
without touching the vault. The vault's system-assigned identity
needs wrap/unwrap access on the key. The key is the ONLY part of
encryption that updates in place -- encryption itself can never
be removed once enabled.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Free-form tags applied to the vault, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

## Validation Rules

- `dpbv_crr_requires_geo_redundant`: cross_region_restore_enabled requires redundancy GeoRedundant -- cross-region restore only exists on geo-redundant backup storage
- `dpbv_encryption_requires_system_assigned_identity`: encryption requires the identity block with SYSTEM_ASSIGNED or SYSTEM_AND_USER_ASSIGNED -- Azure unwraps the key with the vault's system-assigned identity

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataProtectionBackupVault, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.backup_vault_id` | `string` | The Azure Resource Manager ID of the vault -- what backup policies and backup instances reference their vault by. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.DataProtection/backupVaults/{name} |
| `status.outputs.backup_vault_name` | `string` | The vault's name. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the vault's system-assigned identity, when one is enabled -- what Key Vault access policies and datasource RBAC grants bind to (customer-managed-key encryption, backup permissions on protected resources). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.encryption.keyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataProtectionBackupInstance | `spec.vaultId` | `status.outputs.backup_vault_id` |
| AzureDataProtectionBackupPolicy | `spec.vaultId` | `status.outputs.backup_vault_id` |

## See Also

- [Overview](../README.md)
