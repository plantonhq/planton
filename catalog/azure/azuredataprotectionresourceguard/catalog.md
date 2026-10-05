# Azure Data Protection Resource Guard

Creates a Data Protection Resource Guard -- the approval gate behind Multi-User Authorization (MUA) for Azure Backup vaults: once a vault references a guard, privileged vault operations (disabling soft delete, deleting protection, reducing retention) require an approval through the guard before they execute. The guard's protection comes entirely from separation of scope -- place it where a DIFFERENT administrator controls it than the vaults it guards, so an attacker who compromises the vault's scope still cannot approve their own destructive operations. The guard itself is a free configuration object, and one guard serves many vaults.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Resource Guard** -- the Multi-User Authorization gate (`Microsoft.DataProtection/resourceGuards`), with its critical-operation exclusion list
- **Azure Tags** -- Planton-derived resource tags (organization, environment, resource ID) merged under any user tags

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureResourceGroup** -- ideally one a DIFFERENT administrator controls than the vaults the guard will protect; that separation IS the security model.

### Azure Subscription

- **Plan who owns the guard's scope before deploying** -- a guard in the same scope as its vaults is a speed bump, not a control.
- **Vaults opt in by reference** -- creating the guard changes nothing until a vault references its ARM ID.
- **The guard is free** -- it is a pure governance object.

## Deploy

### Console

Open the deployment store, find **Azure Data Protection Resource Guard**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **MUA Guard** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataProtectionResourceGuard
metadata:
  name: backup-mua-guard
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: security-team-rg
  name: backup-mua-guard
```

```shell
planton apply -f resource-guard.yaml
```

This creates the strongest-posture guard: no exclusions, so every critical vault operation requires an approval through it -- in a resource group the security team (not the backup admins) controls. An Infra Job tracks the provisioning in real time.

### InfraChart

When the guard's resource group is deployed in the same InfraPipeline, ValueFromRef wires the reference -- though in the MUA model the guard usually lives in ANOTHER team's scope, deployed separately from the vaults it protects:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: security-team-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph -- resource group first, then the guard.

## Key Configuration

These are the most important decisions when configuring a Resource Guard. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The guard is only as strong as its scope separation** -- Multi-User Authorization works because the approver is SOMEONE ELSE. Deploy the guard into a resource group (or better, a subscription) administered by a different person or team than the vaults it protects. A guard the vault's admin also controls adds a click, not a control.

**Empty exclusion list is the strong posture** -- with no exclusions, EVERY critical vault operation requires an approval through the guard. Add entries to `vaultCriticalOperationExclusionList` only for operations the org has deliberately decided to leave ungated, and treat each entry as a standing security decision worth a comment in the manifest. The list updates in place.

**Some operations can never be excluded** -- Azure keeps a mandatory always-guarded set and rejects any exclusion list naming one of them at create time: the operations that would disarm the guard itself (`backupconfig/write`, both vault families' `backupResourceGuardProxies/delete`, `backupVaults/write#reduceSoftDeleteSecurity`) and the cross-tenant vault-mapping operations. The Terraform provider does not check this; Planton's validation mirrors ARM's list so a bad manifest fails at admission instead of mid-deploy. Note the asymmetry: `backupconfig/write` is mandatory but `backupconfig/delete` is excludable, and `backupResourceGuardProxies/write` is excludable while its `/delete` sibling is not.

**One guard serves the estate** -- vaults reference the guard by ARM ID; there is no per-vault guard object. One well-placed guard per environment (or per compliance boundary) keeps the approval path singular and auditable.

**Deleting the guard removes the gate** -- the guard deletes cleanly even while vaults reference it, and with it goes the approval requirement. Treat guard deletion as itself a privileged operation: gate it with resource locks or pipeline approvals, because Azure will not gate the guard with itself.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `resource_guard_id` | The ARM ID of the guard | AzureRecoveryServicesVault's `resourceGuardId` -- putting a vault's privileged operations under the guard's approval |
| `resource_guard_name` | The guard's name | Operational tooling and audit |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Everything guarded** -- no exclusions, deployed in the security team's scope; every critical operation on referencing vaults needs an approval. Start from the **MUA Guard** preset.

**One guard per compliance boundary** -- a single guard referenced by every backup vault in the environment, keeping the approval path singular and its audit trail in one place.

**Deliberately ungated operations** -- a short, commented exclusion list for operations the org has decided do not need a second approver, reviewed like any other standing security exception.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the guard's home scope, owned by a different administrator than the vaults
- [**Azure Recovery Services Vault**](/infra-catalog/azure-recovery-services-vault) -- references the guard's ARM ID through its `resourceGuardId` field
- [**Azure Data Protection Backup Vault**](/infra-catalog/azure-data-protection-backup-vault) -- the modern vault family whose privileged operations (like disabling soft delete) MUA exists to gate
