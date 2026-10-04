# AzureBackupProtectedFileShare

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureBackupProtectedFileShareSpec** defines a protected file
share (ARM: Microsoft.RecoveryServices/vaults/{vault}/.../
protectedItems/AzureFileShare;{system-name}) -- the binding that
puts one Azure Files share under a backup policy's protection.
Creating it only REGISTERS protection: the first backup runs on the
policy's schedule, not immediately.

**The share's storage account must already be registered with the
vault** (AzureBackupContainerStorageAccount) -- Azure discovers
protectable shares only inside registered accounts, and the create
fails loudly when the account is not registered. Wire
source_storage_account_id to the REGISTRATION's echoed output (the
default reference) so the registration always deploys first.

**Destroying this resource stops protection AND deletes the backup
data** (the modules keep the engines' default destroy behavior).
Vault soft delete -- always on since Azure's secure-by-default
policy -- may retain the deleted item's data for 14 days.

Everything except backup_policy_id is fixed at creation (the
provider's own contract); re-pointing the policy updates in place.

## Example

```yaml
# Offline-plan test manifest. The protected item is five references --
# the seam worth proving offline is the wire map itself (all five
# resolve to plain strings) and the provider's Inquire-then-protect
# addressing.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBackupProtectedFileShare
metadata:
  name: test-backup-protected-file-share
  org: test-org
  env: dev
spec:
  resourceGroup:
    value: test-rg
  recoveryVaultName:
    value: test-backup-vault
  sourceStorageAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/data-rg/providers/Microsoft.Storage/storageAccounts/appfiles
  sourceFileShareName:
    value: team-share
  backupPolicyId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.RecoveryServices/vaults/test-backup-vault/backupPolicies/daily-share-policy
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.recoveryVaultName` | `string \| valueFrom` | yes |  | AzureRecoveryServicesVault (`status.outputs.recovery_services_vault_name`) |
| `spec.sourceStorageAccountId` | `string \| valueFrom` | yes |  | AzureBackupContainerStorageAccount (`status.outputs.storage_account_id`), AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.sourceFileShareName` | `string \| valueFrom` | yes |  | AzureStorageShare (`status.outputs.share_name`) |
| `spec.backupPolicyId` | `string \| valueFrom` | yes |  | AzureBackupPolicyFileShare (`status.outputs.backup_policy_id`) |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the protecting VAULT lives in (NOT
necessarily the storage account's group). Can be a literal
resource-group name or a reference to an AzureResourceGroup's
name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.recoveryVaultName

`string | valueFrom` · required

The Recovery Services vault that protects the share, by NAME (ARM
addresses protected items as children of a vault). Fixed at
creation.

- references: AzureRecoveryServicesVault (`status.outputs.recovery_services_vault_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureRecoveryServicesVault, name: <that resource's name>, fieldPath: status.outputs.recovery_services_vault_name}} -- a bare string does not parse

### spec.sourceStorageAccountId

`string | valueFrom` · required

The storage account holding the share, by ARM ID. The DEFAULT
reference targets the account's vault REGISTRATION
(AzureBackupContainerStorageAccount echoes the account ID it
registered) so the registration deploys before this protection --
the reference carries both the value and the dependency edge. For
accounts registered outside the catalog, pass the account's ARM
ID as a literal (or a valueFrom naming AzureStorageAccount, a declared
candidate whose storage_account_id output fills in). Fixed at creation.

- references: AzureBackupContainerStorageAccount (`status.outputs.storage_account_id`), AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureBackupContainerStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.sourceFileShareName

`string | valueFrom` · required

The file share to protect, by NAME (3-63 lowercase letters,
digits and hyphens -- the share's own naming rule). The share
must live in the storage account above. Fixed at creation.

- references: AzureStorageShare (`status.outputs.share_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageShare, name: <that resource's name>, fieldPath: status.outputs.share_name}} -- a bare string does not parse

### spec.backupPolicyId

`string | valueFrom` · required

The file-share backup policy that governs schedule and retention,
by ARM ID. The policy must live in the protecting vault.
Re-pointing to a different policy updates in place -- the spec's
ONLY updatable field.

- references: AzureBackupPolicyFileShare (`status.outputs.backup_policy_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureBackupPolicyFileShare, name: <that resource's name>, fieldPath: status.outputs.backup_policy_id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureBackupProtectedFileShare, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.backup_protected_file_share_id` | `string` | The Azure Resource Manager ID of the protected item. Azure names the item by the share's SYSTEM name (not its friendly name). Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.RecoveryServices/vaults/{vault}/backupFabrics/Azure/protectionContainers/StorageContainer;storage;{sa-rg};{sa-name}/protectedItems/AzureFileShare;{system-name} |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.recoveryVaultName` | AzureRecoveryServicesVault | `status.outputs.recovery_services_vault_name` |
| `spec.sourceStorageAccountId` | AzureBackupContainerStorageAccount | `status.outputs.storage_account_id` |
| `spec.sourceStorageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.sourceFileShareName` | AzureStorageShare | `status.outputs.share_name` |
| `spec.backupPolicyId` | AzureBackupPolicyFileShare | `status.outputs.backup_policy_id` |

## See Also

- [Overview](../README.md)
