# Azure Backup Policy (VM)

Creates an Azure Backup policy for IaaS virtual machines -- the schedule and layered retention rules that govern VM backups in a Recovery Services vault. The policy itself is a free configuration object; cost follows the VMs protected under it and their backup storage. The schedule's frequency decides which retention layers are legal (Hourly and Daily require `retentionDaily`; Weekly requires `retentionWeekly` instead), and Hourly schedules exist only on the V2 enhanced policy generation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **VM backup policy** -- an ARM child of its vault (`.../vaults/{vault}/backupPolicies/{name}`) carrying the policy generation (V1 or V2), the schedule, the daily/weekly/monthly/yearly retention ladder, instant-restore settings, optional archive tiering, and the snapshot consistency class

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Recovery Services vault** -- referenced by name through `recoveryVaultName`; the policy is an ARM child of the vault, and `resourceGroup` names the vault's group. The vault reference is fixed at creation.

## Deploy

### Console

Open the deployment store, find **Azure Backup Policy (VM)**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Daily Backup Policy** or **Hourly Enhanced Policy** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBackupPolicyVm
metadata:
  name: daily-backup-policy
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  recoveryVaultName:
    value: "acme-prod-vault"
  name: daily-backup-policy
  policyType: V2
  backup:
    frequency: Daily
    time: "23:00"
  retentionDaily:
    count: 30
```

```shell
planton apply -f policy.yaml
```

This creates a V2 policy that backs up its VMs nightly at 23:00 UTC and keeps each backup for 30 days -- ready for AzureBackupProtectedVm bindings to attach VMs to it. An Infra Job tracks the provisioning in real time.

### InfraChart

When a chart provisions the vault and its policies together, wire the references so the InfraPipeline orders them:

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
  name: daily-backup-policy
  policyType: V2
  backup:
    frequency: Daily
    time: "23:00"
  retentionDaily:
    count: 30
```

The InfraPipeline resolves the dependency graph, deploys the resource group and vault first, then creates the policy inside the resolved vault.

## Key Configuration

These are the most important decisions when configuring a VM backup policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Start V2 unless you have a reason not to** -- `policyType` defaults to V1 (the provider's own default), but V2 -- the enhanced generation -- is what new designs should state explicitly: it unlocks hourly schedules, zonally resilient instant restore, and instant-restore retention beyond 5 days, at no extra cost class. The catch: `policyType` is fixed at creation, and a policy replacement re-binds every protected VM under it -- decide the generation before VMs bind, not after.

**The retention ladder is your backup bill** -- Backup storage cost is retention math: dailies + weeklies + monthlies + yearlies, each a recovery point at incremental storage prices. The classic 30/12/12/7 shape suits most workloads; resist keeping everything daily for years. For multi-year retention, add `tieringPolicy` (`TierRecommended` is the safe default; `TierAfter` archives every point past a configured age) so aged points move to archive-tier prices.

**Azure's daily floor is 1 or 7+** -- The service rejects 2-6 days of daily retention at create time, a rule that surfaces nowhere in the portal until it fails. The manifest validation front-loads it: `retentionDaily.count` is 1 (a single rolling daily) or 7 and above.

**Instant restore must be shorter than daily retention** -- Azure keeps a few days of snapshots next to the VM for restores without a vault round-trip. `instantRestoreRetentionDays` must be strictly less than `retentionDaily.count`, or Azure fails with `BMSUserErrorInstantRPRetentionExceedsVaultedRetention` -- an error naming no field. V2 defaults the window to 7 days when unset, so a V2 policy with `retentionDaily.count: 7` and no explicit instant value is undeployable: set the instant window to 1-6, or keep dailies longer than 7. The manifest validation rejects the colliding shapes before Azure does. V1 caps the window at 5 days.

**One time dial for everything** -- `backup.time` sets the backup start AND the retention timestamps of every layer (the provider wires them together), and must land on the hour or half past. Pick a low-traffic window in the policy's `timezone` -- VM snapshots briefly elevate IO.

**Hourly is a window, not a metronome** -- An Hourly V2 schedule runs inside a window: `time` starts it, `hourDuration` bounds it (4-24 hours, a multiple of the interval), `hourInterval` spaces backups within it (4, 6, 8, or 12). Interval 4 with a 12-hour window gives three backups a day, all inside working hours. Weekly schedules instead name `weekdays` and drop the daily retention layer entirely.

**Two retention grammars, mutually exclusive** -- Monthly and yearly layers pick their kept backup either by week-of-month (`weeks` + `weekdays`: "First Sunday") or by month days (`days` and/or `includeLastDays`); mixing the forms in one layer fails validation. `includeLastDays` exists because months differ in length -- it is the only correct way to say "the last day".

**Changing a policy changes every VM under it** -- A policy update applies to all protected VMs at the next scheduled backup; there is no per-VM override. Shortening retention deletes existing recovery points beyond the new horizon on the service side -- treat retention reductions as a deliberate, announced operation, and remember vault immutability blocks them outright.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureRecoveryServicesVault** | `recoveryVaultName` | `status.outputs.recovery_services_vault_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `backup_policy_id` | Azure Resource Manager ID of the policy | AzureBackupProtectedVm's `backupPolicyId` -- how VMs bind to this schedule |

The other output, `backup_policy_name`, echoes the configured name back for reference.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Everyday production policy** -- one nightly backup with the grandfather-father-son ladder (30 dailies, 12 weekly Sundays, 12 first-Sunday monthlies, 7 yearlies) on V2. The right default for most VMs and usually the first policy a vault needs. Start from the **Daily Backup Policy** preset.

**Low-RPO enhanced policy** -- a backup every 4 hours inside a 12-hour working-day window, a week of instant-restore snapshots, and age-based archive tiering. For VMs whose data loses real money by the hour. Hourly points multiply retention math -- `retentionDaily.count: 30` means 30 days of hourly points, so watch the storage line the first month. Start from the **Hourly Enhanced Policy** preset.

**One policy, many VMs** -- a single policy serving every VM with the same recovery objective. Fewer policies means retention changes happen in one reviewable place -- balanced against the blast radius of that one change reaching every VM under it.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the vault's resource group, where the policy lives
- [**Azure Recovery Services Vault**](/infra-catalog/azure-recovery-services-vault) -- the vault the policy is a child of
- [**Azure Backup Protected VM**](/infra-catalog/azure-backup-protected-vm) -- binds an individual VM to this policy via its `backup_policy_id` output
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- the workloads the policy ultimately protects, through their protection bindings
