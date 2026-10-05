# Azure Recovery Services Vault

Creates a Recovery Services vault -- the safe that classic Azure Backup data (VM and file-share backups) and Site Recovery configuration live in. Backup policies and protected items are ARM children of a vault, so one vault typically serves a whole region of workloads. The vault is free at rest -- cost accrues per protected instance and per GB of backup storage -- and several of its settings are one-way doors: redundancy locks at first protection, customer-managed-key encryption never comes off, and Locked immutability is permanent.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Recovery Services Vault** -- with its redundancy, immutability, encryption, identity, and monitoring posture
- **Resource Guard Association** -- created only when `resourceGuardId` is set: Multi-User Authorization on privileged vault operations, one guard per vault (ARM pins the association's name to `VaultProxy`)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An Azure Resource Group** -- the vault is created in it.
- **For customer-managed-key encryption**: an Azure Key Vault Key and, ideally, an Azure User Assigned Identity granted wrap/unwrap on the key before the vault deploys.

### Azure Subscription

- **The vault is free at rest** -- cost starts when items are protected (per-instance fee plus backup storage GB).
- **Redundancy locks once items are protected** -- decide `storageModeType` before the first protection, not after.
- **Deletion is guarded** -- a vault with protected items refuses to delete; these modules deliberately keep the engine-level purge switch off. Stop and delete protections first.

## Deploy

### Console

Open the deployment store, find **Azure Recovery Services Vault**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, redundancy and restore posture, and the optional immutability, identity, and encryption blocks. Start from the **Standard Backup Vault** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRecoveryServicesVault
metadata:
  name: prod-backup-vault
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: prod-backup
  name: prod-backup-vault
  storageModeType: GeoRedundant
  crossRegionRestoreEnabled: true
```

```shell
planton apply -f vault.yaml
```

This creates a Standard-SKU vault with geo-redundant backup storage and cross-region restore enabled, Microsoft-managed encryption, and every built-in alert switch at its all-on default. An Infra Job tracks the provisioning in real time.

### InfraChart

When the resource group, identity, and key are Infra Components in the same chart, wire the compliance-grade shape by reference:

```yaml
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      name: prod-backup
  name: prod-backup-vault
  identity:
    type: USER_ASSIGNED
    identityIds:
      - valueFrom:
          name: backup-cmk-identity
  encryption:
    keyId:
      valueFrom:
        name: backup-cmk-key
    infrastructureEncryptionEnabled: true
    useSystemAssignedIdentity: false
    userAssignedIdentityId:
      valueFrom:
        name: backup-cmk-identity
```

The InfraPipeline resolves the dependency graph, provisioning the identity and key (and their Key Vault grants) before the vault that encrypts with them.

## Key Configuration

These are the most important decisions when configuring an Azure Recovery Services Vault. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Decide redundancy before the first protection.** `storageModeType` updates in place -- right up until the vault protects its first item, after which Azure locks it. GeoRedundant (the default) is the production posture; LocallyRedundant roughly halves backup storage cost for workloads whose loss tolerance genuinely allows a single-region copy; ZoneRedundant replicates across in-region zones. Changing redundancy on a populated vault means a new vault and a migration of protections.

**Cross-region restore is one-way in one direction.** Enabling `crossRegionRestoreEnabled` on a geo-redundant vault is a free in-place update; disabling it REPLACES the vault. It also requires GeoRedundant storage -- validation rejects the pairing with anything else.

**Immutability: Unlocked is the trial run, Locked is forever.** Transitions run `Disabled <-> Unlocked -> Locked`. Unlocked enforces the same protections (no deleting backup data, no reducing retention) but stays reversible; leaving Locked replaces the vault, and a locked vault's backup data cannot be deleted early by anyone. Setting Locked directly works (the modules stage it through Unlocked automatically), but graduate to it only after living with the retention settings.

**CMK encryption is a ratchet with three teeth.** Once `encryption` is set: it can never be removed, `infrastructureEncryptionEnabled` can never change, and the `sku` freezes. The identity needs wrap/unwrap on the key BEFORE the vault deploys -- with a user-assigned identity the grant composes first (the reason to prefer it over system-assigned for CMK); with system-assigned there is a bootstrap hop (deploy without encryption, grant, then update). The key reference targets the versionless URI by default, so rotation propagates without touching the vault.

**The identity never downgrades.** Once the vault has an identity, removing it -- or switching from both flavors to just one -- is rejected. Adding flavors is fine. Decide the identity story once, at creation.

**Monitoring defaults are all-on; turn off deliberately.** All five alert switches default ON service-side -- an unset `monitoring` block IS the sensible posture. The three Site Recovery-related switches are new in provider v5 and only the Terraform engine can turn them OFF (the Pulumi engine's SDK predates them and fails loudly on an explicit false); job-failure and critical-operation alerts switch freely on both engines.

**Multi-user authorization is an org-structure decision.** `resourceGuardId` binds the vault to a Resource Guard so privileged operations (disabling soft delete, reducing retention) need an approval through the guard. The security value comes from the guard living in a DIFFERENT administrator's scope -- a guard the same admin controls is ceremony, not protection.

**Deleting a vault is deliberately hard.** Azure refuses to delete a vault holding protected items, and these modules keep that guard. Teardown order: flip protections off (or destroy the protection resources, which deletes their backup data), wait out soft delete if it holds items, then delete the vault. A "why won't my resource group delete" investigation usually ends at a vault.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Azure Resource Group | `resourceGroup` | `status.outputs.resource_group_name` |
| Azure User Assigned Identity | `identity.identityIds`, `encryption.userAssignedIdentityId` | `status.outputs.identity_id` |
| Azure Key Vault Key | `encryption.keyId` | `status.outputs.versionless_id` |
| Azure Data Protection Resource Guard | `resourceGuardId` | `status.outputs.resource_guard_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `recovery_services_vault_name` | The vault's name | What backup policies, backup containers, and protected items address their vault by (ARM child addressing) |
| `system_assigned_identity_principal_id` | Principal ID of the system-assigned identity, when enabled | Key Vault access policies and RBAC grants for customer-managed-key encryption |

`recovery_services_vault_id` and `resource_guard_association_id` are also exported for identification and audit.

## Common Patterns

**Everyday production vault** -- Geo-redundant storage with cross-region restore, Microsoft-managed encryption, alert defaults untouched: the first vault a region's environment needs. Start from the **Standard Backup Vault** preset.

**Compliance-grade vault** -- Immutability at Unlocked, backup data encrypted with your own Key Vault key through a user-assigned identity, public endpoint closed. For postures that must survive both ransomware and policy audits. Start from the **Immutable CMK Vault** preset.

**One vault, a region of protections** -- Deploy the vault once per region-environment, then compose Azure Backup Policy (VM), Azure Backup Policy (File Share), and their protected items against its `recovery_services_vault_name` output -- policies and protections come and go without touching the vault.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- where the vault lives; reference its `resource_group_name` output.
- [**Azure Backup Policy (VM)**](/infra-catalog/azure-backup-policy-vm) -- schedule and retention for VM backups, addressed by the vault's name output.
- [**Azure Backup Policy (File Share)**](/infra-catalog/azure-backup-policy-file-share) -- schedule and retention for file-share backups in this vault.
- [**Azure Backup Protected VM**](/infra-catalog/azure-backup-protected-vm) -- puts a VM under a policy in this vault.
- [**Azure Backup Protected File Share**](/infra-catalog/azure-backup-protected-file-share) -- puts a file share under a policy in this vault.
- [**Azure Backup Container (Storage Account)**](/infra-catalog/azure-backup-container-storage-account) -- registers a storage account with this vault ahead of file-share protection.
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed key for `encryption.keyId`; the versionless reference rotates automatically.
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the identity that unwraps the key; grantable before the vault exists.
- [**Azure Data Protection Resource Guard**](/infra-catalog/azure-data-protection-resource-guard) -- Multi-User Authorization over privileged vault operations.
