# Azure Data Protection Backup Vault

Creates a Data Protection backup vault -- the safe that modern Azure Backup data lives in: managed disks, blob storage, AKS clusters, MySQL and PostgreSQL flexible servers, and Data Lake storage. Backup policies and backup instances are ARM children of a vault, and the vault is free at rest -- cost accrues per protected instance and per GB of backup storage. Three settings are one-way doors the provider replaces the vault to walk back: cross-region restore once enabled, immutability once Locked, and soft delete once AlwaysOn.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Data Protection Backup Vault** -- with its datastore tier, redundancy, soft-delete, immutability, and managed-identity posture
- **Customer-Managed-Key Encryption** (optional) -- backup data encrypted with your own Key Vault key, when the `encryption` block is set
- **Azure Tags** -- Planton-derived resource tags (organization, environment, resource ID) merged under any user tags

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureResourceGroup** -- the vault is created in it; reference its `resource_group_name` output.
- **For CMK encryption (optional)** -- an AzureKeyVaultKey (reference its versionless ID output); the vault's SYSTEM-assigned identity needs wrap/unwrap access on the key before enabling.

### Azure Subscription

- **Pick the right vault generation first** -- the classic Recovery Services vault protects IaaS VMs and Azure Files shares; THIS vault protects the modern datasource family. A workload belongs to exactly one family.
- **Datastore tier and redundancy are fixed at creation** -- changing either replaces the vault; there is no in-place migration.
- **A vault refuses deletion while backup instances remain inside it** -- and Azure's delete call returns before the vault is fully gone (the modules poll until the name frees up), so budget a few extra minutes in teardown automation.

## Deploy

### Console

Open the deployment store, find **Azure Data Protection Backup Vault**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Backup Vault** or **Immutable CMK Vault** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataProtectionBackupVault
metadata:
  name: prod-backup-vault
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-backup-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-backup-vault
  datastoreType: VaultStore
  redundancy: GeoRedundant
  crossRegionRestoreEnabled: true
  identity:
    type: SYSTEM_ASSIGNED
```

```shell
planton apply -f backup-vault.yaml
```

This creates the everyday production vault: the standard vault-store tier on geo-redundant backup storage with cross-region restore, soft delete at its 14-day default, Microsoft-managed encryption, and a system-assigned identity ready for datasource grants. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the whole protection stack as one chart, ValueFromRef wires the vault to its resource group -- and CMK encryption to a Key Vault key -- deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-backup-rg
      fieldPath: status.outputs.resource_group_name
  identity:
    type: SYSTEM_ASSIGNED
  encryption:
    keyId:
      valueFrom:
        kind: AzureKeyVaultKey
        name: backup-cmk
        fieldPath: status.outputs.versionless_id
```

The InfraPipeline resolves the dependency graph -- resource group and key first, then the vault, then its policies and instances.

## Key Configuration

These are the most important decisions when configuring a backup vault. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Datastore tier and redundancy are create-time decisions** -- both replace the vault when changed. `VaultStore` + `GeoRedundant` is the production posture; drop to `LocallyRedundant` only for dev/test where losing backups with the region is acceptable. A wrong choice means a new vault and re-protecting everything.

**Cross-region restore is a one-way door in one direction** -- enabling `crossRegionRestoreEnabled` on a geo-redundant vault is an in-place update; disabling it replaces the vault. It only exists on GeoRedundant storage (the spec rejects the pairing otherwise). Enable it when paired-region restore justifies the storage premium; treat the decision as permanent.

**Immutability: Unlocked is the trial run, Locked is forever** -- `Unlocked` blocks backup deletion but can itself be turned off; run production vaults there until retention settings have survived a few review cycles. `Locked` is permanent: leaving it replaces the vault, and Azure will not shorten retention or delete backups inside it.

**Soft delete has its own ratchet** -- `On` (the default) retains deleted backup data for `retentionDurationInDays` (14-180), making deletion recoverable. `AlwaysOn` makes that permanent -- the setting can never leave AlwaysOn without replacing the vault. `Off` exists for dev/test churn, and disabling soft delete is exactly the class of privileged operation a Resource Guard gates in production. Note the window governs how long DELETED backups linger, not how long backups are kept -- retention lives on policies.

**CMK encryption can never be removed** -- once the `encryption` block is applied, customer-managed-key encryption is on for the vault's lifetime; there is no path back to Microsoft-managed keys. The KEY rotates freely -- the reference's versionless default means new key versions propagate automatically. Azure unwraps the key with the vault's SYSTEM-assigned identity (hardcoded), so a user-assigned-only identity cannot serve encryption.

**The identity is the protection principal** -- every backup instance's datasource grants ("Disk Backup Reader", "Storage Account Backup Contributor", AKS trusted access) bind to the vault's `system_assigned_identity_principal_id`. Deploy the vault with `identity.type: SYSTEM_ASSIGNED` from day one; adding it later is an update, but no instance can be created before the grants exist.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureKeyVaultKey** (CMK encryption) | `encryption.keyId` | `status.outputs.versionless_id` |
| **AzureUserAssignedIdentity** (user-assigned identity) | `identity.identityIds` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `backup_vault_id` | The ARM ID of the vault | AzureDataProtectionBackupPolicy and AzureDataProtectionBackupInstance reference their vault by it |
| `system_assigned_identity_principal_id` | The vault identity's principal ID | AzureRoleAssignment grants on datasources; Key Vault access for CMK encryption |
| `backup_vault_name` | The vault's name | Operational tooling and audit |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard production vault** -- vault-store tier on geo-redundant storage with cross-region restore, default soft delete, Microsoft-managed encryption; the right starting point for a region's disk, blob, AKS, and database backups. Start from the **Standard Backup Vault** preset.

**Compliance-grade vault** -- trial-run (Unlocked) immutability, a 30-day soft-delete window, and backup data encrypted with your own Key Vault key; for postures that must survive both ransomware and a compromised administrator. Start from the **Immutable CMK Vault** preset.

**Multi-user authorization** -- for backups that must survive a compromised administrator, pair the vault with an Azure Data Protection Resource Guard living in a DIFFERENT administrator's scope. The guard's value comes entirely from that separation.

**Dev/test churn vault** -- `softDelete: Off` and LocallyRedundant storage for estates that create and destroy protection frequently, trading the safety nets for clean, fast teardown.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- where the vault is created
- [**Azure Data Protection Backup Policy**](/infra-catalog/azure-data-protection-backup-policy) -- the schedules and retention contracts living on this vault
- [**Azure Data Protection Backup Instance**](/infra-catalog/azure-data-protection-backup-instance) -- the bindings that put datasources under this vault's protection
- [**Azure Data Protection Resource Guard**](/infra-catalog/azure-data-protection-resource-guard) -- multi-user authorization gating this vault's destructive operations
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed key behind the encryption block
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- the datasource grants the vault's identity carries
