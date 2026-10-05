# Azure Backup Policy (File Share)

Creates an Azure Backup policy for Azure Files shares -- the schedule and layered retention rules that govern file-share backups in a Recovery Services vault. The policy itself is a free configuration object; cost follows the shares protected under it and their backup storage. File-share schedules run Daily or Hourly only (there is no Weekly schedule), and the `backupTier` choice decides whether backups live as snapshots inside the storage account or are additionally copied into the vault.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **File-share backup policy** -- an ARM child of its vault (`.../vaults/{vault}/backupPolicies/{name}`) carrying the schedule, the daily/weekly/monthly/yearly retention ladder, the timezone, and the snapshot or vault-standard tier choice. ARM carries no tags on backup policies.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Recovery Services vault** -- referenced by name through `recoveryVaultName`; the policy is an ARM child of the vault, and `resourceGroup` names the vault's group. The vault reference is fixed at creation.

## Deploy

### Console

Open the deployment store, find **Azure Backup Policy (File Share)**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Daily Snapshot Policy** or **Hourly Vault-Standard Policy** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBackupPolicyFileShare
metadata:
  name: daily-snapshot-policy
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  recoveryVaultName:
    value: "acme-prod-vault"
  name: daily-snapshot-policy
  backup:
    frequency: Daily
    time: "23:00"
  retentionDaily:
    count: 30
```

```shell
planton apply -f policy.yaml
```

This creates a snapshot-tier policy that backs up its shares nightly at 23:00 UTC and keeps each backup for 30 days -- ready for AzureBackupProtectedFileShare bindings to attach shares to it. An Infra Job tracks the provisioning in real time.

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
  name: daily-snapshot-policy
  backup:
    frequency: Daily
    time: "23:00"
  retentionDaily:
    count: 30
```

The InfraPipeline resolves the dependency graph, deploys the resource group and vault first, then creates the policy inside the resolved vault.

## Key Configuration

These are the most important decisions when configuring a file-share backup policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Snapshot or vault-standard is THE decision** -- `backupTier: snapshot` (the default) keeps backups as share snapshots inside the storage account: restores are fast, but the backups share the account's fate -- delete or compromise the account and the "backups" go with it. `vault-standard` additionally copies backups into the vault, which is what "backup" usually means to an auditor. Pick vault-standard for anything whose loss would matter; keep snapshot for dev shares where speed beats durability.

**The snapshot budget is 200, and retention math spends it** -- Azure Files holds at most 200 snapshots per share, and the retention ladder spends that budget: 30 dailies + 12 weeklies + 12 monthlies + 5 yearlies is 59 kept points -- fine; 200 dailies alone maxes it out. The bounds are already shorter than VM policies (daily and weekly cap at 200, monthly at 120, yearly at 10) -- design inside them.

**Daily names a time; Hourly configures a window** -- A Daily schedule sets `backup.time` ("HH:mm", on the hour or half past). An Hourly schedule has no `time` field at all: `backup.hourly` opens a window at `startTime`, bounds it with `windowDuration` (4-24 hours), and spaces backups by `interval` (4, 6, 8, or 12 hours). Interval 4 with a 12-hour window gives three backups inside business hours and none at night. The manifest validation enforces the shape-to-frequency pairing before anything reaches Azure.

**vault-standard's snapshot retention is a strict bound** -- `snapshotRetentionInDays` (vault-standard only) keeps local snapshots alongside the vaulted copies for fast operational restores. It must be strictly less than `retentionDaily.count` -- the provider rejects equality. Five local days against thirty vaulted dailies is the everyday shape.

**Two retention grammars, mutually exclusive** -- Monthly and yearly layers pick their kept backup either by week-of-month (`weeks` + `weekdays` together: "First Sunday") or by month days (`days` and/or `includeLastDays`); mixing the two forms in one layer fails validation. `retentionDaily` is always required -- it is the base layer both frequencies retain into.

**Changing a policy changes every share under it** -- A policy update applies to all protected shares at the next scheduled backup; there is no per-share override. Shortening retention deletes existing recovery points beyond the new horizon on the service side -- treat retention reductions as a deliberate, announced operation.

**The policy is step one of three** -- A policy alone protects nothing. The full chain: register the share's storage account with the vault (AzureBackupContainerStorageAccount), then bind each share to this policy (AzureBackupProtectedFileShare). One policy can serve many shares across many registered accounts in the same vault.

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
| `backup_policy_id` | Azure Resource Manager ID of the policy | AzureBackupProtectedFileShare's `backupPolicyId` -- how shares bind to this schedule |

The other output, `backup_policy_name`, echoes the configured name back for reference.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Everyday nightly policy** -- one nightly snapshot backup with a grandfather-father-son ladder (30 dailies, 12 weekly Sundays, 12 first-Sunday monthlies, 5 yearlies). The right default for most shares and usually the first policy a vault needs. Start from the **Daily Snapshot Policy** preset.

**Durable low-RPO policy** -- backups every 4 hours inside a business-hours window, copied into the vault (vault-standard) so they survive storage-account deletion or ransomware, with five days of local snapshots for fast restores. Hourly backups multiply kept points quickly -- watch the 200-snapshot budget when extending the ladder. Start from the **Hourly Vault-Standard Policy** preset.

**One policy, many shares** -- a single policy serving every share with the same recovery objective, across all registered accounts in the vault. Fewer policies means retention changes happen in one reviewable place -- balanced against the blast radius of that one change reaching every share under it.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the vault's resource group, where the policy lives
- [**Azure Recovery Services Vault**](/infra-catalog/azure-recovery-services-vault) -- the vault the policy is a child of
- [**Azure Backup Container (Storage Account)**](/infra-catalog/azure-backup-container-storage-account) -- registers each storage account whose shares the policy will protect
- [**Azure Backup Protected File Share**](/infra-catalog/azure-backup-protected-file-share) -- binds an individual share to this policy via its `backup_policy_id` output
