# Azure Backup Protected File Share

Puts one Azure Files share under a backup policy's protection in a Recovery Services vault -- the last link in the backup chain (vault, registration, policy, protection). Creating it only registers protection: the first backup runs at the policy's next scheduled run, not immediately. The share's storage account must already be registered with the vault, and destroying the binding stops protection AND deletes the backup data, subject to the vault's 14-day soft-delete hold.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Protected file share** -- an ARM child of the vault (`.../protectionContainers/StorageContainer;storage;{sa-rg};{sa-name}/protectedItems/AzureFileShare;{system-name}`); Azure names the item by the share's SYSTEM name, which differs from its friendly name

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **The storage account registered with the vault** (AzureBackupContainerStorageAccount) -- Azure discovers protectable shares only inside registered accounts; the create fails with `fileshare not found in protectable or protected fileshares, make sure Storage Account ... is registered` when it is not, and that error costs minutes of discovery time before it surfaces. `sourceStorageAccountId`'s default reference wires through the registration's echoed output so the ordering is automatic.
- **A file-share backup policy in the SAME vault** (AzureBackupPolicyFileShare) -- referenced by ARM ID through `backupPolicyId`.
- **The share itself** (AzureStorageShare) -- referenced by name through `sourceFileShareName`; it must live in the registered account.

## Deploy

### Console

Open the deployment store, find **Azure Backup Protected File Share**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Protect a File Share** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBackupProtectedFileShare
metadata:
  name: protect-app-files
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  recoveryVaultName:
    value: "acme-prod-vault"
  sourceStorageAccountId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.Storage/storageAccounts/acmeprodfiles"
  sourceFileShareName:
    value: "app-files"
  backupPolicyId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.RecoveryServices/vaults/acme-prod-vault/backupPolicies/daily-snapshot-policy"
```

```shell
planton apply -f protected-share.yaml
```

This binds the share to the policy's schedule and retention -- protection is registered immediately, and the first recovery point lands at the policy's next scheduled run. An Infra Job tracks the provisioning in real time.

### InfraChart

When a chart provisions the whole backup chain, wire the references so the InfraPipeline orders vault, registration, policy, and protection:

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
  sourceStorageAccountId:
    valueFrom:
      kind: AzureBackupContainerStorageAccount
      name: files-backup-registration
      fieldPath: status.outputs.storage_account_id
  sourceFileShareName:
    valueFrom:
      kind: AzureStorageShare
      name: app-files
      fieldPath: status.outputs.share_name
  backupPolicyId:
    valueFrom:
      kind: AzureBackupPolicyFileShare
      name: daily-snapshot-policy
      fieldPath: status.outputs.backup_policy_id
```

The InfraPipeline resolves the dependency graph, deploys the registration and policy first, then binds the share with the resolved values.

## Key Configuration

These are the most important decisions when configuring a protected file share. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Wire the account through its registration, not directly** -- `sourceStorageAccountId`'s default reference targets the AzureBackupContainerStorageAccount registration's echoed `storage_account_id` output. The value is identical to the account's own ARM ID, but the reference edge guarantees the registration deploys before this protection -- reference the account directly and a fresh chart deploy can race the discovery pass and fail. For accounts registered outside the catalog, pass the ARM ID as a literal.

**Only the policy re-points in place** -- `backupPolicyId` is the spec's single updatable field: moving the share to a different policy in the same vault is an in-place update. Everything else -- vault, account, share name -- is identity: changing any of them replaces the protection with a new protected item, and the old item's backup data follows the vault's soft-delete rules.

**The first backup is not at create time** -- Creating the binding registers protection; the first recovery point lands at the policy's next scheduled run. A share protected at 09:00 against a 23:00 daily policy has no restore point until 23:00 -- run an on-demand backup outside Planton if you need one sooner.

**Destroy deletes the data, and soft delete holds it** -- Destroying this resource stops protection and deletes the backup data. Vault soft delete holds a deleted item with recovery points for 14 days: it still counts against the vault, blocks unregistering the storage account, and blocks vault deletion. Teardown order is protections, then registration, then vault. A protection destroyed before its first backup ever ran deletes outright, and the delete is asynchronous beyond what the IaC engines poll -- reads on the item can keep answering briefly after a destroy the engine already reported successful.

**Budget for the 80-minute operation class** -- Protection configuration is a long-running ARM operation; the provider allows 80 minutes for create, update, and delete. Normal runs finish in minutes -- budget pipelines for the class, not the average.

**Azure renames the share internally** -- The protected item's ARM ID carries the share's system name (`AzureFileShare;{system-name}`), not its friendly name. When correlating with `az backup item list`, match by the friendly-name field, not the ID segment.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureRecoveryServicesVault** | `recoveryVaultName` | `status.outputs.recovery_services_vault_name` |
| **AzureBackupContainerStorageAccount** | `sourceStorageAccountId` | `status.outputs.storage_account_id` |
| **AzureStorageShare** | `sourceFileShareName` | `status.outputs.share_name` |
| **AzureBackupPolicyFileShare** | `backupPolicyId` | `status.outputs.backup_policy_id` |

### What This Kind Provides

This kind is the end of the backup chain: its only output, `backup_protected_file_share_id`, is the protected item's ARM ID, and no downstream Infra Component consumes it. Restore operations happen through the vault, not through references to this binding.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Restore-ready from day one** -- an app chart provisioning the share, the account's registration, and this protection in one deploy, all wired by reference. New environments come up already under the backup policy instead of waiting for someone to remember. Start from the **Protect a File Share** preset.

**One binding per share, one policy for many** -- protection is per share, so an account with five shares needs five bindings -- but all five can point at the same `backupPolicyId`. Give a share its own policy only when its recovery objective genuinely differs.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the vault's resource group, where the protection lives
- [**Azure Recovery Services Vault**](/infra-catalog/azure-recovery-services-vault) -- the vault that protects the share
- [**Azure Backup Container (Storage Account)**](/infra-catalog/azure-backup-container-storage-account) -- the account registration this binding wires through
- [**Azure Storage Share**](/infra-catalog/azure-storage-share) -- the file share being protected
- [**Azure Backup Policy (File Share)**](/infra-catalog/azure-backup-policy-file-share) -- the schedule and retention the share binds to
