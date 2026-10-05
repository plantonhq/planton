# AzureRecoveryServicesVault

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureRecoveryServicesVaultSpec** defines a Recovery Services vault
(ARM: Microsoft.RecoveryServices/vaults) -- the safe that classic
Azure Backup data (VM and file-share backups) and Site Recovery
configuration live in. Backup policies and protected items are ARM
children of a vault; one vault typically serves a whole
region-of-workloads.

A vault is FREE at rest -- cost accrues per protected instance and
per GB of backup storage, not for the vault object itself.

**Deleting a vault fails while protected items remain inside it**
(Azure's own guard; the provider purges them first only when the
engine-level `purge_protected_items_from_vault_on_destroy` feature
is enabled, which these modules deliberately leave OFF). Stop and
delete the protections first, then delete the vault.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: the geo-redundant
# default with cross-region restore, an Unlocked immutability posture,
# customer-managed-key encryption with the versionless key URI and
# infrastructure encryption, a system-assigned identity, the two
# engine-portable monitoring switches, the composed Resource Guard
# association, and user tags merged over the derived ones. The three
# v5-only monitoring switches stay unset by design (canonical manifests
# never exercise PARITY-EXCEPTION fields).
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRecoveryServicesVault
metadata:
  name: test-recovery-services-vault
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: test-backup-vault
  sku: Standard
  storageModeType: GeoRedundant
  crossRegionRestoreEnabled: true
  publicNetworkAccessEnabled: false
  immutability: Unlocked
  identity:
    type: SYSTEM_ASSIGNED
  encryption:
    keyId:
      value: https://test-kv.vault.azure.net/keys/vault-cmk
    infrastructureEncryptionEnabled: true
  monitoring:
    alertsForAllJobFailuresEnabled: true
    alertsForCriticalOperationFailuresEnabled: false
  resourceGuardId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DataProtection/resourceGuards/test-guard
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.sku` | `string` |  | `Standard` |  |
| `spec.storageModeType` | `string` |  | `GeoRedundant` |  |
| `spec.crossRegionRestoreEnabled` | `bool` |  |  |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.immutability` | `string` |  |  |  |
| `spec.identity` | `AzureRecoveryServicesVaultIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.encryption` | `AzureRecoveryServicesVaultEncryption` |  |  |  |
| `spec.encryption.keyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.encryption.infrastructureEncryptionEnabled` | `bool` |  |  |  |
| `spec.encryption.useSystemAssignedIdentity` | `bool` |  | `true` |  |
| `spec.encryption.userAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.monitoring` | `AzureRecoveryServicesVaultMonitoring` |  |  |  |
| `spec.monitoring.alertsForAllJobFailuresEnabled` | `bool` |  | `true` |  |
| `spec.monitoring.alertsForAllFailoverIssuesEnabled` | `bool` |  |  |  |
| `spec.monitoring.alertsForAllReplicationIssuesEnabled` | `bool` |  |  |  |
| `spec.monitoring.alertsForCriticalOperationFailuresEnabled` | `bool` |  | `true` |  |
| `spec.monitoring.emailNotificationsForSiteRecoveryEnabled` | `bool` |  |  |  |
| `spec.resourceGuardId` | `string \| valueFrom` |  |  | AzureDataProtectionResourceGuard (`status.outputs.resource_guard_id`) |
| `spec.classicVmwareReplicationEnabled` | `bool` |  |  |  |
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
characters, letters, digits and hyphens, starting with a letter
(the provider's own rule). Backup policies and protected items
address the vault by this name. Changing the name replaces the
vault.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z][-a-zA-Z0-9]{1,49}$"}}

### spec.sku

`string` · optional (explicit presence)

The vault's SKU (the wire values). "Standard" and "RS0" are the
only values ARM accepts, and they price identically -- "RS0" is
the legacy tier-style name (the provider pins its tier to
"Standard"). Unspecified applies "Standard". NOTE: once
customer-managed-key encryption is enabled the SKU can no longer
be changed (the provider's own update guard).

- default: `Standard`
- rule: {"string":{"in":["Standard","RS0"]}}

### spec.storageModeType

`string` · optional (explicit presence)

How the vault's backup storage replicates (the provider's
storage_mode_type). GeoRedundant keeps a copy in the paired
region (the default and the recommended posture for production);
LocallyRedundant is the cheapest; ZoneRedundant replicates across
availability zones in-region. Unspecified applies GeoRedundant.
Switch it ONLY while the vault protects nothing -- Azure locks
redundancy once items are protected (an apply-time contract the
service enforces).

- default: `GeoRedundant`
- rule: {"string":{"in":["GeoRedundant","LocallyRedundant","ZoneRedundant"]}}

### spec.crossRegionRestoreEnabled

`bool`

Whether backup data can be restored in the paired region
(cross-region restore). Requires geo-redundant storage. Enabling
is an in-place update; DISABLING it replaces the vault (the
provider's one-way ForceNew).

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the vault's endpoints answer the public internet.
Unspecified applies true (the provider's default). Set false for
the private-endpoint-only posture.

- default: `true`

### spec.immutability

`string`

The vault's immutability posture (the wire values). Immutability
stops backup data being deleted or retention being reduced --
ransomware protection. Transitions are one-way at the end:
Disabled <-> Unlocked -> Locked, and LOCKED IS PERMANENT (leaving
Locked replaces the vault). Setting Locked directly is staged by
the provider through Unlocked automatically. Unspecified leaves
the service default (Disabled).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Locked","Unlocked","Disabled"]}}

### spec.identity

`AzureRecoveryServicesVaultIdentity`

The vault's managed identity -- required for customer-managed-key
encryption and used for private-endpoint integrations. NOTE: once
an identity is set it must never be removed or switched to the
opposite flavor alone (the provider rejects the downgrade --
Azure's CMK guidance).

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the vault (the common choice for CMK); USER_ASSIGNED brings
identities you manage (grantable Key Vault access BEFORE the
vault exists); SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_recovery_services_vault_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the vault.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the vault, by ARM ID. Reference
AzureUserAssignedIdentity resources so Key Vault grants can be
composed before the vault is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.encryption

`AzureRecoveryServicesVaultEncryption`

Encryption of backup data with your OWN Key Vault key
(customer-managed key) instead of Microsoft-managed keys. Requires
the identity block. Once enabled it can never be disabled, and
infrastructure_encryption_enabled can never change afterwards
(the provider's own update guards).

- rule: set use_system_assigned_identity: false when user_assigned_identity_id is set -- the key is unwrapped by exactly one identity
- rule: user_assigned_identity_id is required when use_system_assigned_identity is false

### spec.encryption.keyId

`string | valueFrom` · required

The Key Vault KEY that encrypts the vault's backup data.
Versionless and versioned key URIs are BOTH accepted by the
provider; the versionless form (this reference's default) is the
right choice -- key rotation then propagates automatically,
without touching the vault. The vault's identity needs
wrap/unwrap access on the key.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.encryption.infrastructureEncryptionEnabled

`bool`

Whether backup data is encrypted a SECOND time at the
infrastructure layer (double encryption). Decide at creation --
once encryption is enabled this can NEVER change (the provider's
own update guard). ARM requires an explicit choice, so this plain
bool always ships on the wire.

### spec.encryption.useSystemAssignedIdentity

`bool` · optional (explicit presence)

Whether the vault's SYSTEM-assigned identity unwraps the key.
Unspecified applies true. Set false to use a user-assigned
identity instead (user_assigned_identity_id then becomes
required).

- default: `true`

### spec.encryption.userAssignedIdentityId

`string | valueFrom`

For use_system_assigned_identity false: the user-assigned
identity that unwraps the key, by ARM ID. Must also be attached
in the vault's identity block.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.monitoring

`AzureRecoveryServicesVaultMonitoring`

The vault's built-in Azure Monitor alert settings. Every switch
defaults to ON (the provider's and the service's own posture);
configure the block only to turn specific alert classes off.

### spec.monitoring.alertsForAllJobFailuresEnabled

`bool` · optional (explicit presence)

Alerts for ALL backup job failures (not just critical ones).
Unspecified applies true.

- default: `true`

### spec.monitoring.alertsForAllFailoverIssuesEnabled

`bool` · optional (explicit presence)

Alerts for all Site Recovery FAILOVER issues. Unspecified applies
true (the service's own default).

PARITY-EXCEPTION: this switch is new in azurerm v5 and absent
from the classic Pulumi SDK -- the Pulumi module rejects an
explicit FALSE loudly (leave it unset, or use the terraform
module to turn it off). An explicit true is wire-equivalent to
the default and passes on both engines.

### spec.monitoring.alertsForAllReplicationIssuesEnabled

`bool` · optional (explicit presence)

Alerts for all Site Recovery REPLICATION issues. Unspecified
applies true (the service's own default).

PARITY-EXCEPTION: new in azurerm v5, absent from the classic
Pulumi SDK -- the Pulumi module rejects an explicit FALSE loudly
(leave it unset, or use the terraform module to turn it off).

### spec.monitoring.alertsForCriticalOperationFailuresEnabled

`bool` · optional (explicit presence)

Alerts for critical operations (deleting backup data, disabling
protection). Unspecified applies true.

- default: `true`

### spec.monitoring.emailNotificationsForSiteRecoveryEnabled

`bool` · optional (explicit presence)

Email notifications to subscription owners for Site Recovery
events. Unspecified applies true (the service's own default).

PARITY-EXCEPTION: new in azurerm v5, absent from the classic
Pulumi SDK -- the Pulumi module rejects an explicit FALSE loudly
(leave it unset, or use the terraform module to turn it off).

### spec.resourceGuardId

`string | valueFrom`

The ARM ID of a Resource Guard to associate with the vault --
Multi-User Authorization: privileged vault operations (disabling
soft delete, reducing retention) then require an approval through
the guard, which typically lives in a DIFFERENT administrator's
scope. References an AzureDataProtectionResourceGuard's ARM-id
output (or a literal ARM ID for a guard outside the catalog, e.g.
in another administrator's subscription). Fixed at creation of the
association.

- references: AzureDataProtectionResourceGuard (`status.outputs.resource_guard_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataProtectionResourceGuard, name: <that resource's name>, fieldPath: status.outputs.resource_guard_id}} -- a bare string does not parse

### spec.classicVmwareReplicationEnabled

`bool`

Whether the vault is configured for classic VMware-to-Azure
replication (Site Recovery's legacy VMware flow). Fixed at
creation.

### spec.tags

`map<string, string>`

Free-form tags applied to the vault, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

## Validation Rules

- `rsv_crr_requires_geo_redundant`: cross_region_restore_enabled requires storage_mode_type GeoRedundant (the default) -- cross-region restore only exists on geo-redundant backup storage
- `rsv_encryption_requires_identity`: encryption requires the identity block -- the vault needs a managed identity to read your Key Vault key

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureRecoveryServicesVault, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.recovery_services_vault_id` | `string` | The Azure Resource Manager ID of the vault. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.RecoveryServices/vaults/{name} |
| `status.outputs.recovery_services_vault_name` | `string` | The vault's name -- what backup policies and protected items address their vault by (ARM child addressing). |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the vault's system-assigned identity, when one is enabled -- what Key Vault access policies and RBAC grants bind to for customer-managed-key encryption. |
| `status.outputs.resource_guard_association_id` | `string` | The ARM ID of the vault's Resource Guard association, when spec.resource_guard_id composes one. Empty otherwise. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.encryption.keyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.encryption.userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.resourceGuardId` | AzureDataProtectionResourceGuard | `status.outputs.resource_guard_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureBackupContainerStorageAccount | `spec.recoveryVaultName` | `status.outputs.recovery_services_vault_name` |
| AzureBackupPolicyFileShare | `spec.recoveryVaultName` | `status.outputs.recovery_services_vault_name` |
| AzureBackupPolicyVm | `spec.recoveryVaultName` | `status.outputs.recovery_services_vault_name` |
| AzureBackupProtectedFileShare | `spec.recoveryVaultName` | `status.outputs.recovery_services_vault_name` |
| AzureBackupProtectedVm | `spec.recoveryVaultName` | `status.outputs.recovery_services_vault_name` |

## See Also

- [Overview](../README.md)
