# Azure Backup Container (Storage Account)

Registers a storage account with a Recovery Services vault as a backup container -- the one-time prerequisite that lets the account's file shares be protected. Registration is free and moves no data; cost starts when shares are protected. One registration exists per storage-account-and-vault pair, and every field is fixed at creation -- ARM has no update on protection containers, so changing any field replaces the registration.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Backup container registration** -- an ARM child of the vault (`.../vaults/{vault}/backupFabrics/Azure/protectionContainers/StorageContainer;storage;{sa-rg};{sa-name}`); ARM derives the registration's own name from the storage account's group and name
- **A DoNotDelete resource lock on the storage account** -- placed by Azure Backup for as long as the account is registered, protecting the backups' source; removed at unregister, never by hand

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Recovery Services vault** -- referenced by name through `recoveryVaultName`; the registration lives inside it, in the vault's resource group (`resourceGroup` names the VAULT's group, not necessarily the storage account's).
- **A storage account in the vault's region** -- referenced by ARM ID through `storageAccountId`. Azure Files backup is regional: cross-region registration fails at apply time with a service error, and nothing in the manifest can pre-check it.

## Deploy

### Console

Open the deployment store, find **Azure Backup Container (Storage Account)**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Register a Storage Account** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBackupContainerStorageAccount
metadata:
  name: files-backup-registration
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  recoveryVaultName:
    value: "acme-prod-vault"
  storageAccountId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.Storage/storageAccounts/acmeprodfiles"
```

```shell
planton apply -f registration.yaml
```

This registers the storage account with the vault as a backup container -- no data moves, no cost accrues, and the account's file shares become protectable. An Infra Job tracks the provisioning in real time.

### InfraChart

When a chart provisions the vault, the storage account, and the registration together, wire all three references so the InfraPipeline orders them:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  recoveryVaultName:
    valueFrom:
      kind: AzureRecoveryServicesVault
      name: prod-vault
      fieldPath: status.outputs.recovery_services_vault_name
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: prod-files
      fieldPath: status.outputs.storage_account_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, vault, and storage account first, then registers the account with the resolved values.

## Key Configuration

These are the most important decisions when configuring a backup container registration. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One registration per account-and-vault pair -- never per share** -- Registering an already-registered account fails with "already exists": the registration is a singleton on the (vault, storage account) pair. Deploy one AzureBackupContainerStorageAccount per account, then any number of AzureBackupProtectedFileShare bindings ride it. A chart provisioning several shares in one account points them all at the same registration.

**Wire protections through the registration, not the account** -- AzureBackupProtectedFileShare's `sourceStorageAccountId` should reference this resource's `storage_account_id` output (its default reference does exactly that), not the storage account directly. The value is identical, but the reference edge is what guarantees the registration deploys before the protection -- reference the account directly and a fresh chart deploy can race: the protection's discovery pass finds an unregistered account and fails.

**The resource lock blocks share deletes too** -- The DoNotDelete lock Azure Backup places on the account covers the account's children: deleting ANY file share in a registered account fails with `ScopeLocked` naming the account -- even a share that was never protected. If automation tears down shares alongside this registration, the unregister must run first; a share delete that hits `ScopeLocked` is an ordering problem, not a stuck lock.

**Teardown runs strictly backwards** -- Unregistering refuses while any of the account's shares are still protected. The order is always: destroy the AzureBackupProtectedFileShare bindings, then this registration, then (if ever) the vault. Vault soft delete can hold a deleted protection's data for 14 days, and a held item can delay the unregister -- if unregister fails right after destroying protections, the soft-deleted items are why; undelete-and-purge them or wait out the window.

**Everything is fixed at creation** -- All three fields are ForceNew. Pointing the registration at a different vault or account is a replace: the old registration unregisters (subject to the protected-shares refusal above) and a new one is created.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureRecoveryServicesVault** | `recoveryVaultName` | `status.outputs.recovery_services_vault_name` |
| **AzureStorageAccount** | `storageAccountId` | `status.outputs.storage_account_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `storage_account_id` | The registered account's ARM ID, echoed after reference resolution | AzureBackupProtectedFileShare's `sourceStorageAccountId` -- referencing the echo carries both the value and the deploy-order edge, guaranteeing the registration exists before protection begins |

The other output, `backup_container_id`, is the registration's own ARM ID -- nothing downstream consumes it; it identifies the registration in the vault.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Register-then-protect in one chart** -- a storage account, its registration, and one AzureBackupProtectedFileShare per share, all in a single InfraChart. The registration node makes the ordering automatic: shares wire their `sourceStorageAccountId` through the registration's echo output. Start from the **Register a Storage Account** preset.

**Registering an existing account** -- the registration with a literal `storageAccountId` pointing at an account provisioned outside the chart. Same singleton rule applies; the account must already live in the vault's region.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the vault's resource group, where the registration lives
- [**Azure Recovery Services Vault**](/infra-catalog/azure-recovery-services-vault) -- the vault the account registers with
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the account being registered for backup
- [**Azure Backup Protected File Share**](/infra-catalog/azure-backup-protected-file-share) -- per-share protection bindings that wire through this registration's `storage_account_id` output
- [**Azure Backup Policy (File Share)**](/infra-catalog/azure-backup-policy-file-share) -- the schedule-and-retention policy those protections attach to
