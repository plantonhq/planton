# AzureBackupContainerStorageAccount

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureBackupContainerStorageAccountSpec** registers a storage
account with a Recovery Services vault as a backup container (ARM:
Microsoft.RecoveryServices/vaults/{vault}/backupFabrics/Azure/
protectionContainers/StorageContainer;storage;{sa-rg};{sa-name}).
Registration is the prerequisite for protecting any of the
account's file shares: ONE registration per storage-account-and-
vault pair, then each share gets its own
AzureBackupProtectedFileShare binding. Registration itself is free
and moves no data.

**Every field is fixed at creation** (changing any replaces the
registration -- ARM has no update on protection containers).
While registered, Azure Backup places a RESOURCE LOCK on the
storage account (protecting the backups' source); unregistering
removes it. Unregistering REFUSES while any of the account's
shares are still protected -- destroy the protections first (the
GUIDE's teardown recipe).

## Example

```yaml
# Offline-plan test manifest. The registration is three references --
# the seam worth proving offline is the wire map itself (all three
# resolve to plain strings) and ARM's derived container addressing.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBackupContainerStorageAccount
metadata:
  name: test-backup-container-storage-account
  org: test-org
  env: dev
spec:
  resourceGroup:
    value: test-rg
  recoveryVaultName:
    value: test-backup-vault
  storageAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/data-rg/providers/Microsoft.Storage/storageAccounts/appfiles
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.recoveryVaultName` | `string \| valueFrom` | yes |  | AzureRecoveryServicesVault (`status.outputs.recovery_services_vault_name`) |
| `spec.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the registering VAULT lives in (NOT
necessarily the storage account's group). Can be a literal
resource-group name or a reference to an AzureResourceGroup's
name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.recoveryVaultName

`string | valueFrom` · required

The Recovery Services vault the storage account registers with,
by NAME (ARM addresses backup containers as children of a vault).

- references: AzureRecoveryServicesVault (`status.outputs.recovery_services_vault_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureRecoveryServicesVault, name: <that resource's name>, fieldPath: status.outputs.recovery_services_vault_name}} -- a bare string does not parse

### spec.storageAccountId

`string | valueFrom` · required

The storage account to register, by ARM ID. The account must live
in the vault's region (Azure Files backup is regional).

The registration REGISTERS this account with the vault and is an ARM
child of the vault, never a thing inside the account, so on a diagram
the reference is access, not placement.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureBackupContainerStorageAccount, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.backup_container_id` | `string` | The Azure Resource Manager ID of the backup container registration. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.RecoveryServices/vaults/{vault}/backupFabrics/Azure/protectionContainers/StorageContainer;storage;{sa-rg};{sa-name} |
| `status.outputs.storage_account_id` | `string` | The registered storage account's ARM ID, echoed from the spec after reference resolution. Protected file shares reference THIS output for their source_storage_account_id so the registration deploys first -- the reference carries both the value and the dependency edge (the provider docs' own wiring pattern). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.recoveryVaultName` | AzureRecoveryServicesVault | `status.outputs.recovery_services_vault_name` |
| `spec.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureBackupProtectedFileShare | `spec.sourceStorageAccountId` | `status.outputs.storage_account_id` |

## See Also

- [Overview](../README.md)
