# Azure Backup Protected VM

Registers one virtual machine under a backup policy's protection in a Recovery Services vault. Creating it only registers protection -- the first backup runs at the policy's next scheduled window, not immediately -- and destroying it stops protection AND deletes the backup data (the engines' default destroy behavior, kept deliberately). Cost follows protection: a per-instance protected fee plus backup storage, both on the vault's bill.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Protected item** -- the vault-side registration binding the VM to the policy (ARM: `.../protectedItems/VM;iaasvmcontainerv2;{vm-rg};{vm-name}`; ARM derives the item's name from the VM's group and name), carrying the optional disk filters and protection posture

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Recovery Services vault** -- referenced by name through `recoveryVaultName`; `resourceGroup` names the vault's group, not necessarily the VM's. A VM can be protected by exactly one vault.
- **A VM backup policy in the same vault** (AzureBackupPolicyVm) -- referenced by ARM ID through `backupPolicyId`.
- **The virtual machine to protect** -- referenced by ARM ID through `sourceVmId`; the VM must live in the vault's region.

## Deploy

### Console

Open the deployment store, find **Azure Backup Protected VM**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Protect a VM** or **Selective Disk Protection** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBackupProtectedVm
metadata:
  name: protect-app-vm
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  recoveryVaultName:
    value: "acme-prod-vault"
  sourceVmId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.Compute/virtualMachines/app-vm"
  backupPolicyId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.RecoveryServices/vaults/acme-prod-vault/backupPolicies/daily-backup-policy"
```

```shell
planton apply -f protected-vm.yaml
```

This registers the VM under the policy with all disks backed up and an Azure-managed protection posture -- the item reads `IRPending` until the policy's first scheduled backup runs. An Infra Job tracks the provisioning in real time.

### InfraChart

When a chart provisions the VM alongside its protection, wire the references so nothing ships unprotected:

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
  sourceVmId:
    valueFrom:
      kind: AzureVirtualMachine
      name: app-vm
      fieldPath: status.outputs.vm_id
  backupPolicyId:
    valueFrom:
      kind: AzureBackupPolicyVm
      name: daily-backup-policy
      fieldPath: status.outputs.backup_policy_id
```

The InfraPipeline resolves the dependency graph, deploys the vault, policy, and VM first, then registers the protection with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a protected VM. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The first backup is not at create time** -- Creating the protection registers the VM and nothing more: the item sits in `IRPending` (initial replication pending) until the policy's next scheduled window runs the first full backup. A VM protected at 09:00 against a 23:00 policy is unprotected for fourteen hours. When onboarding critical machines, trigger the first backup manually with `az backup protection backup-now` instead of waiting for the schedule.

**Destroy means "delete my backups" -- decide if that is what you want** -- Destroying this resource stops protection and deletes the backup data. Two escape hatches exist as engine-level features rather than spec fields: `vm_backup_stop_protection_and_retain_data_on_destroy` keeps the data and stops backups; the `suspend` variant keeps it under an immutable vault's rules. The vault itself refuses to delete while protected items remain -- teardown order is protections first, vault last.

**Destroy can report success before -- or without -- the delete landing** -- The protection delete is asynchronous beyond what the IaC engines poll: after a clean destroy, the item can keep answering reads for minutes, and in a measured case the vault ran no delete job at all -- the item survived active while the destroy exited clean. A surviving item then blocks its policy's delete (`BMSUserErrorPolicyObjectInUse`) and the vault's delete (`BMSUserErrorVaultDeletionNotAllowed`). If a teardown wedges this way, disable protection by hand -- `az backup protection disable --delete-backup-data true --yes` with the friendly VM name for both `--container-name` and `--item-name` (the full semicolon container ID fails from the CLI) -- and the rest of the chain deletes normally.

**Soft delete holds backup data for 14 days** -- Deleting protection soft-deletes the recovery points: they hold vault storage (unbilled) and the VM's registration for 14 days. Re-protecting the same VM inside that window collides with the ghost -- the provider recovers it automatically only when its `recover_soft_deleted_backup_protected_vm` feature is on; otherwise undelete manually (`az backup protection undelete`) or wait out the window.

**Disk filters: exclude for savings, include for surgical protection** -- `excludeDiskLuns` backs up everything except the listed data disks -- right for rebuildable content (caches, scratch, tempdb) and databases with native backup tooling, where full-disk backup pays twice for the same bytes. `includeDiskLuns` inverts it: the OS disk plus only the listed disks. The two are mutually exclusive (validated at manifest time), and the OS disk is always backed up either way.

**protectionState is a dial, not a status display** -- Leave it unset in normal operation; Azure manages the posture, and transient states read back as Protected. `ProtectionStopped` halts backups while keeping existing data -- the "pause" for a VM being decommissioned gradually. `BackupsSuspended` does the same but only works while the vault is immutable (Unlocked or Locked) -- Azure enforces that against the live vault at apply, so flipping a mutable vault's protection to BackupsSuspended fails at deploy time, not manifest time.

**One VM, one vault -- Azure's rule** -- Re-pointing `backupPolicyId` at a different policy in the same vault updates in place (the supported re-policy path). Re-pointing `sourceVmId` replaces the protection; moving to a different VAULT means delete-protection-here, wait out soft delete, protect-there. Plan vault topology before mass onboarding, not after.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureRecoveryServicesVault** | `recoveryVaultName` | `status.outputs.recovery_services_vault_name` |
| **AzureVirtualMachine** | `sourceVmId` | `status.outputs.vm_id` |
| **AzureBackupPolicyVm** | `backupPolicyId` | `status.outputs.backup_policy_id` |

### What This Kind Provides

This kind is the end of the backup chain: its only output, `backup_protected_vm_id`, is the protected item's ARM ID, and no downstream Infra Component consumes it. Restores run from the vault in the portal or CLI (`az backup restore`), not through references to this binding.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Protected from day one** -- a chart provisioning the VM and its protection together, all wired by reference, so no machine ships unprotected while someone remembers to onboard it. Start from the **Protect a VM** preset.

**Selective disk protection** -- a database VM's OS disk protected while the data disks (covered by SQL dumps or WAL archiving) are excluded via `excludeDiskLuns`. You stop paying for the same bytes twice; the excluded disks come back from their native tooling. Rehearse that combined restore before trusting it. Start from the **Selective Disk Protection** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the vault's resource group, where the protected item lives
- [**Azure Recovery Services Vault**](/infra-catalog/azure-recovery-services-vault) -- the vault that protects the VM
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- the machine being protected, referenced by its `vm_id` output
- [**Azure Backup Policy (VM)**](/infra-catalog/azure-backup-policy-vm) -- the schedule and retention the VM binds to
