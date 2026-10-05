# Azure Data Protection Backup Policy

Creates a Data Protection backup policy -- WHEN backups run (ISO-8601 repeating intervals) and HOW LONG they are kept (a default retention plus optional named rules that tag specific backups, like first-of-week, for longer keeps). One kind covers the six datasource types as variants: blob storage, managed disks, AKS clusters, MySQL flexible servers, PostgreSQL flexible servers, and Data Lake storage -- the variant block you set IS the datasource type. The policy is a free configuration object and it is immutable: every field is fixed at creation, so every change replaces the policy.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions exactly ONE of:

- **Blob Storage Backup Policy** -- operational (continuous, in-account) and/or vault (scheduled) retention tiers -- the only dual-tier variant
- **Disk Backup Policy** -- scheduled incremental snapshots on the operational tier
- **Kubernetes Cluster Backup Policy** -- scheduled AKS backups on the operational tier
- **MySQL / PostgreSQL Flexible Server Backup Policy** -- scheduled full backups on the vault tier
- **Data Lake Storage Backup Policy** -- scheduled ADLS backups on the vault tier

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureDataProtectionBackupVault** -- the policy is an ARM child of it; reference its `backup_vault_id` output.

### Azure Subscription

- **The policy is free** -- cost follows the protected instances and their backup storage.
- **Schedules are ISO-8601 repeating intervals** -- `R/2024-01-01T00:00:00+00:00/P1D` means daily from that instant; durations like `P7D`, `P4M`, `P10Y` set retention.
- **Time zones are Windows-style names** -- "India Standard Time", not "Asia/Kolkata"; the MySQL and Data Lake variants validate the value strictly.

## Deploy

### Console

Open the deployment store, find **Azure Data Protection Backup Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Daily Disk Backup**, **AKS Cluster Backup**, or **Blob Dual-Tier Backup** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataProtectionBackupPolicy
metadata:
  name: daily-disk-policy
  org: acme-corp
  env: prod
spec:
  vaultId:
    valueFrom:
      kind: AzureDataProtectionBackupVault
      name: prod-backup-vault
      fieldPath: status.outputs.backup_vault_id
  name: daily-disk-policy
  disk:
    backupRepeatingTimeIntervals:
      - R/2024-01-01T02:00:00+00:00/P1D
    defaultRetentionDuration: P7D
    retentionRules:
      - name: weekly
        duration: P90D
        criteria:
          absoluteCriteria: FirstOfWeek
        priority: 25
```

```shell
planton apply -f backup-policy.yaml
```

This creates a disk policy on the vault: daily incremental snapshots at 02:00 UTC, seven days of default retention, and the first backup of each week kept 90 days. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the vault, its policies, and their instances as one chart, ValueFromRef wires the policy to the vault deployed in the same InfraPipeline:

```yaml
spec:
  vaultId:
    valueFrom:
      kind: AzureDataProtectionBackupVault
      name: prod-backup-vault
      fieldPath: status.outputs.backup_vault_id
```

The InfraPipeline resolves the dependency graph -- vault first, then this policy, then the instances that bind to it.

## Key Configuration

These are the most important decisions when configuring a backup policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The policy is immutable -- plan replacements, not edits** -- the provider ships no update path at all: changing a schedule or retention value REPLACES the policy, and the backup instances bound to it re-bind to the replacement (the policy binding is the instance's only in-place-updatable field). Name policies by their contract (`daily-p7d`) rather than by their consumer, so a changed contract reads as the new object it actually is.

**Match the variant to the datasource family** -- each variant maps to exactly one Azure datasource type, and the stores differ by design: disks and AKS retain on the OPERATIONAL store (snapshots near the source), databases and Data Lake on the VAULT store (isolated copies), and blob storage is the only dual-tier variant. The spec's store vocabularies accept only what each datasource's service supports today -- a rejected store value is the service's boundary, not a catalog gap.

**Read the schedule grammar once, carefully** -- `R/2024-01-01T02:00:00+00:00/P1D`: from that instant, repeat daily. The date anchors the phase (backups run at 02:00 UTC because the anchor says so); the `P` duration sets the cadence (`P1D` daily, `P1W` weekly, `PT4H` every four hours).

**Retention rules are tags, not filters** -- a named rule (criteria + duration + priority) TAGS matching backups at creation time: the first backup of the week tagged `weekly` lives out that rule's duration. When several rules match one backup, the LOWEST priority number wins, and the unnamed default layer catches everything untagged. The Data Lake variant is the exception: its rules have no priority field -- ORDER in the list is priority.

**Blob's two tiers are different products** -- the operational tier (`operationalDefaultRetentionDuration` alone) is continuous point-in-time restore inside the storage account: no schedule, no vault copy, no named rules. The vault tier (`vaultDefaultRetentionDuration` plus `backupRepeatingTimeIntervals`) is scheduled copies into the vault that survive account deletion. Configure both for defense in depth; configure only operational when in-account restore is all you need. A blob policy needs at least one tier -- one with neither protects nothing, and the spec rejects it.

**Retention is the cost lever** -- the policy is free, but every duration in it multiplies backup storage. A `P10Y` rule on a fat datasource is a decade-long storage commitment; size the default short and reserve long keeps for narrowly-tagged rules (FirstOfMonth, FirstOfYear).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureDataProtectionBackupVault** | `vaultId` | `status.outputs.backup_vault_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `backup_policy_id` | The ARM ID of the policy | AzureDataProtectionBackupInstance's `backupPolicyId` -- what every instance binds to |
| `backup_policy_name` | The policy's name, unique on its vault | Operational tooling and audit |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Everyday disk protection** -- daily incremental snapshots at 02:00 UTC, a week of default retention, the first backup of each week kept 90 days; the right starting point for VM data disks and standalone managed disks. Start from the **Daily Disk Backup** preset.

**AKS cluster protection** -- backups every four hours with two weeks of default retention and the first backup of each day kept eight weeks. Start from the **AKS Cluster Backup** preset.

**Blob defense in depth** -- continuous point-in-time restore inside the storage account (30 days) PLUS daily vaulted copies that survive account deletion (90 days, first of each month kept a year). Start from the **Blob Dual-Tier Backup** preset.

**One policy, many instances** -- policies are per-datasource-type, not per-resource: bind every production disk to the same disk policy and the retention contract stays auditable in one place.

## Works With

- [**Azure Data Protection Backup Vault**](/infra-catalog/azure-data-protection-backup-vault) -- the vault the policy lives on
- [**Azure Data Protection Backup Instance**](/infra-catalog/azure-data-protection-backup-instance) -- binds a datasource to this policy's schedule and retention
- [**Azure Managed Disk**](/infra-catalog/azure-managed-disk) / [**Azure Storage Account**](/infra-catalog/azure-storage-account) / [**Azure AKS Cluster**](/infra-catalog/azure-aks-cluster) -- the datasources instances put under this policy
- [**Azure MySQL Flexible Server**](/infra-catalog/azure-mysql-flexible-server) / [**Azure PostgreSQL Flexible Server**](/infra-catalog/azure-postgresql-flexible-server) -- the database datasources
