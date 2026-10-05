# AzureDataProtectionBackupPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureDataProtectionBackupPolicySpec** defines a Data Protection
backup policy (ARM: Microsoft.DataProtection/backupVaults/{vault}/
backupPolicies/{name}) -- WHEN backups run (ISO-8601 repeating
intervals) and HOW LONG they are kept (a default retention plus
optional named rules that tag specific backups -- first of day,
first of week -- for longer keeps). ONE kind covers the six
datasource types as variants: blob storage, managed disks,
Kubernetes (AKS) clusters, MySQL flexible servers, PostgreSQL
flexible servers and Data Lake storage. Exactly one variant block
is set; the block IS the datasource type.

**A policy is immutable**: every field on every variant is fixed at
creation (the provider ships no update path -- its own contract).
Changing anything replaces the policy; backup instances then
re-bind to the replacement.

**The policy itself is a free configuration object** -- cost
follows the protected instances and their backup storage.

**Retention grammar, shared by all variants**: durations are
ISO-8601 ("P7D" = 7 days, "P4M" = 4 months, "P10Y" = 10 years);
backup schedules are ISO-8601 repeating intervals
("R/2024-01-01T00:00:00+00:00/P1D" = daily from that instant).
Named retention rules carry a priority (lower wins when several
rules tag the same backup) and criteria choosing WHICH backups the
rule keeps (absolute markers like FirstOfDay/FirstOfWeek, or
calendar selectors).

## Example

```yaml
# Offline-plan test manifest. Exercises the union's deepest single
# shape: the blob variant's DUAL retention tiers (operational +
# vault), the schedule intervals the vault tier requires, and a named
# vault-tier rule carrying calendar criteria including the
# last-day-of-month encoding (days_of_month 0). The other five
# variants' wire shapes are proven by the scenario manifests and the
# per-variant offline plans.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataProtectionBackupPolicy
metadata:
  name: test-data-protection-backup-policy
  org: test-org
  env: dev
spec:
  vaultId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DataProtection/backupVaults/test-backup-vault
  name: test-blob-dual-tier-policy
  blobStorage:
    operationalDefaultRetentionDuration: P30D
    vaultDefaultRetentionDuration: P90D
    backupRepeatingTimeIntervals:
      - R/2024-01-01T00:00:00+00:00/P1D
    timeZone: UTC
    retentionRules:
      - name: month-end
        criteria:
          absoluteCriteria: FirstOfMonth
          daysOfMonth:
            - 0
            - 28
          weeksOfMonth:
            - Last
          monthsOfYear:
            - December
          scheduledBackupTimes:
            - "2024-01-01T00:00:00Z"
        lifeCycle:
          dataStoreType: VaultStore
          duration: P1Y
        priority: 20
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.vaultId` | `string \| valueFrom` | yes |  | AzureDataProtectionBackupVault (`status.outputs.backup_vault_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.blobStorage` | `AzureDataProtectionBackupPolicyBlobStorage` |  |  |  |
| `spec.blobStorage.operationalDefaultRetentionDuration` | `string` |  |  |  |
| `spec.blobStorage.vaultDefaultRetentionDuration` | `string` |  |  |  |
| `spec.blobStorage.backupRepeatingTimeIntervals` | `[]string` |  |  |  |
| `spec.blobStorage.timeZone` | `string` |  |  |  |
| `spec.blobStorage.retentionRules` | `[]AzureDataProtectionBackupPolicyBlobStorageRetentionRule` |  |  |  |
| `spec.blobStorage.retentionRules[].name` | `string` | yes |  |  |
| `spec.blobStorage.retentionRules[].criteria` | `AzureDataProtectionBackupPolicyBlobStorageCriteria` | yes |  |  |
| `spec.blobStorage.retentionRules[].criteria.absoluteCriteria` | `string` |  |  |  |
| `spec.blobStorage.retentionRules[].criteria.daysOfMonth` | `[]int32` |  |  |  |
| `spec.blobStorage.retentionRules[].criteria.daysOfWeek` | `[]string` |  |  |  |
| `spec.blobStorage.retentionRules[].criteria.monthsOfYear` | `[]string` |  |  |  |
| `spec.blobStorage.retentionRules[].criteria.scheduledBackupTimes` | `[]string` |  |  |  |
| `spec.blobStorage.retentionRules[].criteria.weeksOfMonth` | `[]string` |  |  |  |
| `spec.blobStorage.retentionRules[].lifeCycle` | `AzureDataProtectionBackupPolicyBlobStorageLifeCycle` | yes |  |  |
| `spec.blobStorage.retentionRules[].lifeCycle.dataStoreType` | `string` | yes |  |  |
| `spec.blobStorage.retentionRules[].lifeCycle.duration` | `string` | yes |  |  |
| `spec.blobStorage.retentionRules[].priority` | `int32` | yes |  |  |
| `spec.disk` | `AzureDataProtectionBackupPolicyDisk` |  |  |  |
| `spec.disk.backupRepeatingTimeIntervals` | `[]string` | yes |  |  |
| `spec.disk.defaultRetentionDuration` | `string` | yes |  |  |
| `spec.disk.retentionRules` | `[]AzureDataProtectionBackupPolicyDiskRetentionRule` |  |  |  |
| `spec.disk.retentionRules[].name` | `string` | yes |  |  |
| `spec.disk.retentionRules[].duration` | `string` | yes |  |  |
| `spec.disk.retentionRules[].criteria` | `AzureDataProtectionBackupPolicyDiskCriteria` | yes |  |  |
| `spec.disk.retentionRules[].criteria.absoluteCriteria` | `string` |  |  |  |
| `spec.disk.retentionRules[].priority` | `int32` | yes |  |  |
| `spec.disk.timeZone` | `string` |  |  |  |
| `spec.kubernetesCluster` | `AzureDataProtectionBackupPolicyKubernetesCluster` |  |  |  |
| `spec.kubernetesCluster.backupRepeatingTimeIntervals` | `[]string` | yes |  |  |
| `spec.kubernetesCluster.defaultRetentionRule` | `AzureDataProtectionBackupPolicyKubernetesClusterDefaultRetentionRule` | yes |  |  |
| `spec.kubernetesCluster.defaultRetentionRule.lifeCycles` | `[]AzureDataProtectionBackupPolicyKubernetesClusterLifeCycle` | yes |  |  |
| `spec.kubernetesCluster.defaultRetentionRule.lifeCycles[].dataStoreType` | `string` | yes |  |  |
| `spec.kubernetesCluster.defaultRetentionRule.lifeCycles[].duration` | `string` | yes |  |  |
| `spec.kubernetesCluster.retentionRules` | `[]AzureDataProtectionBackupPolicyKubernetesClusterRetentionRule` |  |  |  |
| `spec.kubernetesCluster.retentionRules[].name` | `string` | yes |  |  |
| `spec.kubernetesCluster.retentionRules[].criteria` | `AzureDataProtectionBackupPolicyKubernetesClusterCriteria` | yes |  |  |
| `spec.kubernetesCluster.retentionRules[].criteria.absoluteCriteria` | `string` |  |  |  |
| `spec.kubernetesCluster.retentionRules[].criteria.daysOfWeek` | `[]string` |  |  |  |
| `spec.kubernetesCluster.retentionRules[].criteria.monthsOfYear` | `[]string` |  |  |  |
| `spec.kubernetesCluster.retentionRules[].criteria.scheduledBackupTimes` | `[]string` |  |  |  |
| `spec.kubernetesCluster.retentionRules[].criteria.weeksOfMonth` | `[]string` |  |  |  |
| `spec.kubernetesCluster.retentionRules[].lifeCycles` | `[]AzureDataProtectionBackupPolicyKubernetesClusterLifeCycle` | yes |  |  |
| `spec.kubernetesCluster.retentionRules[].lifeCycles[].dataStoreType` | `string` | yes |  |  |
| `spec.kubernetesCluster.retentionRules[].lifeCycles[].duration` | `string` | yes |  |  |
| `spec.kubernetesCluster.retentionRules[].priority` | `int32` | yes |  |  |
| `spec.kubernetesCluster.timeZone` | `string` |  |  |  |
| `spec.mysqlFlexibleServer` | `AzureDataProtectionBackupPolicyMysqlFlexibleServer` |  |  |  |
| `spec.mysqlFlexibleServer.backupRepeatingTimeIntervals` | `[]string` | yes |  |  |
| `spec.mysqlFlexibleServer.defaultRetentionRule` | `AzureDataProtectionBackupPolicyFlexibleServerDefaultRetentionRule` | yes |  |  |
| `spec.mysqlFlexibleServer.defaultRetentionRule.lifeCycles` | `[]AzureDataProtectionBackupPolicyFlexibleServerLifeCycle` | yes |  |  |
| `spec.mysqlFlexibleServer.defaultRetentionRule.lifeCycles[].dataStoreType` | `string` | yes |  |  |
| `spec.mysqlFlexibleServer.defaultRetentionRule.lifeCycles[].duration` | `string` | yes |  |  |
| `spec.mysqlFlexibleServer.retentionRules` | `[]AzureDataProtectionBackupPolicyFlexibleServerRetentionRule` |  |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].name` | `string` | yes |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].criteria` | `AzureDataProtectionBackupPolicyFlexibleServerCriteria` | yes |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].criteria.absoluteCriteria` | `string` |  |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].criteria.daysOfWeek` | `[]string` |  |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].criteria.monthsOfYear` | `[]string` |  |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].criteria.scheduledBackupTimes` | `[]string` |  |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].criteria.weeksOfMonth` | `[]string` |  |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].lifeCycles` | `[]AzureDataProtectionBackupPolicyFlexibleServerLifeCycle` | yes |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].lifeCycles[].dataStoreType` | `string` | yes |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].lifeCycles[].duration` | `string` | yes |  |  |
| `spec.mysqlFlexibleServer.retentionRules[].priority` | `int32` | yes |  |  |
| `spec.mysqlFlexibleServer.timeZone` | `string` |  |  |  |
| `spec.postgresqlFlexibleServer` | `AzureDataProtectionBackupPolicyPostgresqlFlexibleServer` |  |  |  |
| `spec.postgresqlFlexibleServer.backupRepeatingTimeIntervals` | `[]string` | yes |  |  |
| `spec.postgresqlFlexibleServer.defaultRetentionRule` | `AzureDataProtectionBackupPolicyFlexibleServerDefaultRetentionRule` | yes |  |  |
| `spec.postgresqlFlexibleServer.defaultRetentionRule.lifeCycles` | `[]AzureDataProtectionBackupPolicyFlexibleServerLifeCycle` | yes |  |  |
| `spec.postgresqlFlexibleServer.defaultRetentionRule.lifeCycles[].dataStoreType` | `string` | yes |  |  |
| `spec.postgresqlFlexibleServer.defaultRetentionRule.lifeCycles[].duration` | `string` | yes |  |  |
| `spec.postgresqlFlexibleServer.retentionRules` | `[]AzureDataProtectionBackupPolicyFlexibleServerRetentionRule` |  |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].name` | `string` | yes |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].criteria` | `AzureDataProtectionBackupPolicyFlexibleServerCriteria` | yes |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].criteria.absoluteCriteria` | `string` |  |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].criteria.daysOfWeek` | `[]string` |  |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].criteria.monthsOfYear` | `[]string` |  |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].criteria.scheduledBackupTimes` | `[]string` |  |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].criteria.weeksOfMonth` | `[]string` |  |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].lifeCycles` | `[]AzureDataProtectionBackupPolicyFlexibleServerLifeCycle` | yes |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].lifeCycles[].dataStoreType` | `string` | yes |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].lifeCycles[].duration` | `string` | yes |  |  |
| `spec.postgresqlFlexibleServer.retentionRules[].priority` | `int32` | yes |  |  |
| `spec.postgresqlFlexibleServer.timeZone` | `string` |  |  |  |
| `spec.dataLakeStorage` | `AzureDataProtectionBackupPolicyDataLakeStorage` |  |  |  |
| `spec.dataLakeStorage.backupSchedule` | `[]string` | yes |  |  |
| `spec.dataLakeStorage.defaultRetentionDuration` | `string` | yes |  |  |
| `spec.dataLakeStorage.retentionRules` | `[]AzureDataProtectionBackupPolicyDataLakeStorageRetentionRule` |  |  |  |
| `spec.dataLakeStorage.retentionRules[].name` | `string` | yes |  |  |
| `spec.dataLakeStorage.retentionRules[].duration` | `string` | yes |  |  |
| `spec.dataLakeStorage.retentionRules[].absoluteCriteria` | `string` |  |  |  |
| `spec.dataLakeStorage.retentionRules[].daysOfWeek` | `[]string` |  |  |  |
| `spec.dataLakeStorage.retentionRules[].monthsOfYear` | `[]string` |  |  |  |
| `spec.dataLakeStorage.retentionRules[].scheduledBackupTimes` | `[]string` |  |  |  |
| `spec.dataLakeStorage.retentionRules[].weeksOfMonth` | `[]string` |  |  |  |
| `spec.dataLakeStorage.timeZone` | `string` |  |  |  |

## Field Details

### spec.vaultId

`string | valueFrom` · required

The Data Protection backup vault the policy lives in, by ARM ID
(a policy is an ARM child of its vault). Fixed at creation.

- references: AzureDataProtectionBackupVault (`status.outputs.backup_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataProtectionBackupVault, name: <that resource's name>, fieldPath: status.outputs.backup_vault_id}} -- a bare string does not parse

### spec.name

`string` · required

The policy's name, unique on the vault: 3-150 characters,
letters, digits and hyphens (the provider's own rule; the
data_lake_storage variant additionally requires the name to START
with a letter -- enforced below). Changing the name replaces the
policy.

- rule: {"required":true,"string":{"pattern":"^[-a-zA-Z0-9]{3,150}$"}}

### spec.blobStorage

`AzureDataProtectionBackupPolicyBlobStorage`

The blob-storage variant: backs up a storage account's blob
services. The only variant with TWO retention homes -- an
operational tier (continuous, in the storage account) and a
vault tier (scheduled copies in the vault); configure either or
both.

- rule: at least one of operational_default_retention_duration or vault_default_retention_duration must be set -- a blob policy without either tier protects nothing
- rule: vault_default_retention_duration requires backup_repeating_time_intervals -- the vault tier is scheduled, not continuous
- rule: retention_rules require vault_default_retention_duration -- the operational tier cannot carry named retention rules

### spec.blobStorage.operationalDefaultRetentionDuration

`string`

How long operational-tier (in-account) backups are kept, as an
ISO-8601 duration (e.g. "P30D"). Setting this enables the
operational tier.

- rule: operational_default_retention_duration must be an ISO-8601 duration, e.g. P30D

### spec.blobStorage.vaultDefaultRetentionDuration

`string`

How long vault-tier backups are kept by default, as an ISO-8601
duration (e.g. "P90D"). Setting this enables the vault tier and
requires backup_repeating_time_intervals (the schedule the vault
copies run on).

- rule: vault_default_retention_duration must be an ISO-8601 duration, e.g. P90D

### spec.blobStorage.backupRepeatingTimeIntervals

`[]string`

When vault-tier backups run: ISO-8601 repeating intervals, e.g.
"R/2024-01-01T00:00:00+00:00/P1D" (daily). Required with (and
only meaningful for) the vault tier.

- rule: {"repeated":{"items":{"cel":[{"id":"dpbp_blob_interval_iso8601_repeating","message":"each interval must be an ISO-8601 repeating interval, e.g. R/2024-01-01T00:00:00+00:00/P1D","expression":"this.matches('^R/.+/P((\\\\d+Y)?(\\\\d+M)?(\\\\d+W)?(\\\\d+D)?)(T(\\\\d+H)?(\\\\d+M)?(\\\\d+S)?)?$')"}]}}}

### spec.blobStorage.timeZone

`string`

The Windows time zone the schedule is interpreted in, e.g. "UTC"
or "Pacific Standard Time". Unspecified leaves the service
interpretation (UTC).

### spec.blobStorage.retentionRules

`[]AzureDataProtectionBackupPolicyBlobStorageRetentionRule`

Named vault-tier retention rules that keep SPECIFIC backups
longer than the default (e.g. first-of-week for 12 weeks).
Requires the vault tier.

### spec.blobStorage.retentionRules[].name

`string` · required

The rule's name (also the tag ARM stamps on the backups it
keeps).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.blobStorage.retentionRules[].criteria

`AzureDataProtectionBackupPolicyBlobStorageCriteria` · required

WHICH backups this rule keeps.

- rule: {"required":true}

### spec.blobStorage.retentionRules[].criteria.absoluteCriteria

`string`

An absolute marker (the wire values): AllBackup keeps everything;
FirstOfDay/FirstOfWeek/FirstOfMonth/FirstOfYear keep the first
backup of each period.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AllBackup","FirstOfDay","FirstOfWeek","FirstOfMonth","FirstOfYear"]}}

### spec.blobStorage.retentionRules[].criteria.daysOfMonth

`[]int32`

Days of the month, 1-28; 0 means the LAST day of the month (the
provider's own encoding).

- rule: {"repeated":{"items":{"int32":{"lte":28,"gte":0}}}}

### spec.blobStorage.retentionRules[].criteria.daysOfWeek

`[]string`

Days of the week, e.g. "Sunday".

- rule: {"repeated":{"items":{"string":{"in":["Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"]}}}}

### spec.blobStorage.retentionRules[].criteria.monthsOfYear

`[]string`

Months of the year, e.g. "January".

- rule: {"repeated":{"items":{"string":{"in":["January","February","March","April","May","June","July","August","September","October","November","December"]}}}}

### spec.blobStorage.retentionRules[].criteria.scheduledBackupTimes

`[]string`

Specific backup times as RFC-3339 timestamps (only the time of
day is honored), e.g. "2024-01-01T05:30:00Z".

### spec.blobStorage.retentionRules[].criteria.weeksOfMonth

`[]string`

Weeks of the month (the wire values).

- rule: {"repeated":{"items":{"string":{"in":["First","Second","Third","Fourth","Last"]}}}}

### spec.blobStorage.retentionRules[].lifeCycle

`AzureDataProtectionBackupPolicyBlobStorageLifeCycle` · required

HOW LONG the matched backups are kept, on which store.

- rule: {"required":true}

### spec.blobStorage.retentionRules[].lifeCycle.dataStoreType

`string` · required

The store the retention applies to. "VaultStore" is the only
value the service accepts today (the provider mirrors the service
team's own statement; the vocabulary widens if ArchiveStore ever
arrives).

- rule: {"required":true,"string":{"in":["VaultStore"]}}

### spec.blobStorage.retentionRules[].lifeCycle.duration

`string` · required

The retention duration, ISO-8601 (e.g. "P12W").

- rule: duration must be an ISO-8601 duration, e.g. P12W
- rule: {"required":true}

### spec.blobStorage.retentionRules[].priority

`int32` · required · optional (explicit presence)

The rule's priority -- when several rules tag the same backup,
the LOWEST priority number wins. Required (the provider's own
contract); modeled with explicit presence so 0 stays expressible.

- rule: {"required":true}

### spec.disk

`AzureDataProtectionBackupPolicyDisk`

The managed-disk variant: scheduled incremental snapshots kept on
the operational tier.

### spec.disk.backupRepeatingTimeIntervals

`[]string` · required

When backups run: ISO-8601 repeating intervals, e.g.
"R/2024-01-01T02:00:00+00:00/P1D" (daily at 02:00 UTC). At least
one interval is required.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"dpbp_disk_interval_iso8601_repeating","message":"each interval must be an ISO-8601 repeating interval, e.g. R/2024-01-01T02:00:00+00:00/P1D","expression":"this.matches('^R/.+/P((\\\\d+Y)?(\\\\d+M)?(\\\\d+W)?(\\\\d+D)?)(T(\\\\d+H)?(\\\\d+M)?(\\\\d+S)?)?$')"}]}}}

### spec.disk.defaultRetentionDuration

`string` · required

How long backups are kept by default, as an ISO-8601 duration
(e.g. "P7D").

- rule: default_retention_duration must be an ISO-8601 duration, e.g. P7D
- rule: {"required":true}

### spec.disk.retentionRules

`[]AzureDataProtectionBackupPolicyDiskRetentionRule`

Named retention rules that keep SPECIFIC backups longer than the
default. Disk rules select by absolute marker only (the
provider's own surface -- no calendar selectors here).

### spec.disk.retentionRules[].name

`string` · required

The rule's name (also the tag ARM stamps on the backups it
keeps).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.disk.retentionRules[].duration

`string` · required

The retention duration for matched backups, ISO-8601 (e.g.
"P90D"). Disks retain on the operational store (the provider pins
it).

- rule: duration must be an ISO-8601 duration, e.g. P90D
- rule: {"required":true}

### spec.disk.retentionRules[].criteria

`AzureDataProtectionBackupPolicyDiskCriteria` · required

WHICH backups this rule keeps.

- rule: {"required":true}

### spec.disk.retentionRules[].criteria.absoluteCriteria

`string`

An absolute marker (the wire values).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AllBackup","FirstOfDay","FirstOfWeek","FirstOfMonth","FirstOfYear"]}}

### spec.disk.retentionRules[].priority

`int32` · required · optional (explicit presence)

The rule's priority -- when several rules tag the same backup,
the LOWEST priority number wins. Required (the provider's own
contract); modeled with explicit presence so 0 stays expressible.

- rule: {"required":true}

### spec.disk.timeZone

`string`

The Windows time zone the schedule is interpreted in, e.g. "UTC".
Unspecified leaves the service interpretation (UTC).

### spec.kubernetesCluster

`AzureDataProtectionBackupPolicyKubernetesCluster`

The Kubernetes (AKS) cluster variant: scheduled cluster backups
kept on the operational tier.

### spec.kubernetesCluster.backupRepeatingTimeIntervals

`[]string` · required

When backups run: ISO-8601 repeating intervals, e.g.
"R/2024-01-01T00:00:00+00:00/PT4H" (every four hours). At least
one interval is required.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"dpbp_k8s_interval_iso8601_repeating","message":"each interval must be an ISO-8601 repeating interval, e.g. R/2024-01-01T00:00:00+00:00/PT4H","expression":"this.matches('^R/.+/P((\\\\d+Y)?(\\\\d+M)?(\\\\d+W)?(\\\\d+D)?)(T(\\\\d+H)?(\\\\d+M)?(\\\\d+S)?)?$')"}]}}}

### spec.kubernetesCluster.defaultRetentionRule

`AzureDataProtectionBackupPolicyKubernetesClusterDefaultRetentionRule` · required

The default retention applied to every backup no rule tags.

- rule: {"required":true}

### spec.kubernetesCluster.defaultRetentionRule.lifeCycles

`[]AzureDataProtectionBackupPolicyKubernetesClusterLifeCycle` · required

How long default-retained backups live, per store.

- rule: {"repeated":{"minItems":"1"}}

### spec.kubernetesCluster.defaultRetentionRule.lifeCycles[].dataStoreType

`string` · required

The store the retention applies to. "OperationalStore" is the
only value the service accepts today (the provider mirrors the
service team's own statement; the vocabulary widens when the AKS
vault tier lands).

- rule: {"required":true,"string":{"in":["OperationalStore"]}}

### spec.kubernetesCluster.defaultRetentionRule.lifeCycles[].duration

`string` · required

The retention duration, ISO-8601 (e.g. "P14D").

- rule: duration must be an ISO-8601 duration, e.g. P14D
- rule: {"required":true}

### spec.kubernetesCluster.retentionRules

`[]AzureDataProtectionBackupPolicyKubernetesClusterRetentionRule`

Named retention rules that keep SPECIFIC backups longer than the
default.

### spec.kubernetesCluster.retentionRules[].name

`string` · required

The rule's name (also the tag ARM stamps on the backups it
keeps).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.kubernetesCluster.retentionRules[].criteria

`AzureDataProtectionBackupPolicyKubernetesClusterCriteria` · required

WHICH backups this rule keeps.

- rule: {"required":true}

### spec.kubernetesCluster.retentionRules[].criteria.absoluteCriteria

`string`

An absolute marker (the wire values).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AllBackup","FirstOfDay","FirstOfWeek","FirstOfMonth","FirstOfYear"]}}

### spec.kubernetesCluster.retentionRules[].criteria.daysOfWeek

`[]string`

Days of the week, e.g. "Sunday".

- rule: {"repeated":{"items":{"string":{"in":["Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"]}}}}

### spec.kubernetesCluster.retentionRules[].criteria.monthsOfYear

`[]string`

Months of the year, e.g. "January".

- rule: {"repeated":{"items":{"string":{"in":["January","February","March","April","May","June","July","August","September","October","November","December"]}}}}

### spec.kubernetesCluster.retentionRules[].criteria.scheduledBackupTimes

`[]string`

Specific backup times as RFC-3339 timestamps (only the time of
day is honored), e.g. "2024-01-01T05:30:00Z".

### spec.kubernetesCluster.retentionRules[].criteria.weeksOfMonth

`[]string`

Weeks of the month (the wire values).

- rule: {"repeated":{"items":{"string":{"in":["First","Second","Third","Fourth","Last"]}}}}

### spec.kubernetesCluster.retentionRules[].lifeCycles

`[]AzureDataProtectionBackupPolicyKubernetesClusterLifeCycle` · required

HOW LONG the matched backups are kept, per store.

- rule: {"repeated":{"minItems":"1"}}

### spec.kubernetesCluster.retentionRules[].lifeCycles[].dataStoreType

`string` · required

The store the retention applies to. "OperationalStore" is the
only value the service accepts today (the provider mirrors the
service team's own statement; the vocabulary widens when the AKS
vault tier lands).

- rule: {"required":true,"string":{"in":["OperationalStore"]}}

### spec.kubernetesCluster.retentionRules[].lifeCycles[].duration

`string` · required

The retention duration, ISO-8601 (e.g. "P14D").

- rule: duration must be an ISO-8601 duration, e.g. P14D
- rule: {"required":true}

### spec.kubernetesCluster.retentionRules[].priority

`int32` · required · optional (explicit presence)

The rule's priority -- when several rules tag the same backup,
the LOWEST priority number wins. Required (the provider's own
contract); modeled with explicit presence so 0 stays expressible.

- rule: {"required":true}

### spec.kubernetesCluster.timeZone

`string`

The Windows time zone the schedule is interpreted in, e.g. "UTC".
Unspecified leaves the service interpretation (UTC).

### spec.mysqlFlexibleServer

`AzureDataProtectionBackupPolicyMysqlFlexibleServer`

The MySQL flexible-server variant: scheduled full backups kept on
the vault tier.

### spec.mysqlFlexibleServer.backupRepeatingTimeIntervals

`[]string` · required

When backups run: ISO-8601 repeating intervals, e.g.
"R/2024-01-01T00:00:00+00:00/P1W" (weekly). At least one interval
is required.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"dpbp_mysql_interval_iso8601_repeating","message":"each interval must be an ISO-8601 repeating interval, e.g. R/2024-01-01T00:00:00+00:00/P1W","expression":"this.matches('^R/.+/P((\\\\d+Y)?(\\\\d+M)?(\\\\d+W)?(\\\\d+D)?)(T(\\\\d+H)?(\\\\d+M)?(\\\\d+S)?)?$')"}]}}}

### spec.mysqlFlexibleServer.defaultRetentionRule

`AzureDataProtectionBackupPolicyFlexibleServerDefaultRetentionRule` · required

The default retention applied to every backup no rule tags.

- rule: {"required":true}

### spec.mysqlFlexibleServer.defaultRetentionRule.lifeCycles

`[]AzureDataProtectionBackupPolicyFlexibleServerLifeCycle` · required

How long default-retained backups live, per store.

- rule: {"repeated":{"minItems":"1"}}

### spec.mysqlFlexibleServer.defaultRetentionRule.lifeCycles[].dataStoreType

`string` · required

The store the retention applies to. "VaultStore" is the only
value the service accepts today (the provider mirrors the service
team's own statement; the vocabulary widens if ArchiveStore ever
arrives).

- rule: {"required":true,"string":{"in":["VaultStore"]}}

### spec.mysqlFlexibleServer.defaultRetentionRule.lifeCycles[].duration

`string` · required

The retention duration, ISO-8601 (e.g. "P3M").

- rule: duration must be an ISO-8601 duration, e.g. P3M
- rule: {"required":true}

### spec.mysqlFlexibleServer.retentionRules

`[]AzureDataProtectionBackupPolicyFlexibleServerRetentionRule`

Named retention rules that keep SPECIFIC backups longer than the
default.

### spec.mysqlFlexibleServer.retentionRules[].name

`string` · required

The rule's name (also the tag ARM stamps on the backups it
keeps).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.mysqlFlexibleServer.retentionRules[].criteria

`AzureDataProtectionBackupPolicyFlexibleServerCriteria` · required

WHICH backups this rule keeps.

- rule: {"required":true}

### spec.mysqlFlexibleServer.retentionRules[].criteria.absoluteCriteria

`string`

An absolute marker (the wire values).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AllBackup","FirstOfDay","FirstOfWeek","FirstOfMonth","FirstOfYear"]}}

### spec.mysqlFlexibleServer.retentionRules[].criteria.daysOfWeek

`[]string`

Days of the week, e.g. "Sunday".

- rule: {"repeated":{"items":{"string":{"in":["Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"]}}}}

### spec.mysqlFlexibleServer.retentionRules[].criteria.monthsOfYear

`[]string`

Months of the year, e.g. "January".

- rule: {"repeated":{"items":{"string":{"in":["January","February","March","April","May","June","July","August","September","October","November","December"]}}}}

### spec.mysqlFlexibleServer.retentionRules[].criteria.scheduledBackupTimes

`[]string`

Specific backup times as RFC-3339 timestamps (only the time of
day is honored), e.g. "2024-01-01T05:30:00Z".

### spec.mysqlFlexibleServer.retentionRules[].criteria.weeksOfMonth

`[]string`

Weeks of the month (the wire values).

- rule: {"repeated":{"items":{"string":{"in":["First","Second","Third","Fourth","Last"]}}}}

### spec.mysqlFlexibleServer.retentionRules[].lifeCycles

`[]AzureDataProtectionBackupPolicyFlexibleServerLifeCycle` · required

HOW LONG the matched backups are kept, per store.

- rule: {"repeated":{"minItems":"1"}}

### spec.mysqlFlexibleServer.retentionRules[].lifeCycles[].dataStoreType

`string` · required

The store the retention applies to. "VaultStore" is the only
value the service accepts today (the provider mirrors the service
team's own statement; the vocabulary widens if ArchiveStore ever
arrives).

- rule: {"required":true,"string":{"in":["VaultStore"]}}

### spec.mysqlFlexibleServer.retentionRules[].lifeCycles[].duration

`string` · required

The retention duration, ISO-8601 (e.g. "P3M").

- rule: duration must be an ISO-8601 duration, e.g. P3M
- rule: {"required":true}

### spec.mysqlFlexibleServer.retentionRules[].priority

`int32` · required · optional (explicit presence)

The rule's priority -- when several rules tag the same backup,
the LOWEST priority number wins. Required (the provider's own
contract); modeled with explicit presence so 0 stays expressible.

- rule: {"required":true}

### spec.mysqlFlexibleServer.timeZone

`string`

The Windows time zone the schedule is interpreted in (the
provider validates this variant against the Windows time-zone
catalog). Unspecified leaves the service interpretation (UTC).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Afghanistan Standard Time","Alaskan Standard Time","Aleutian Standard Time","Altai Standard Time","Arab Standard Time","Arabian Standard Time","Arabic Standard Time","Argentina Standard Time","Astrakhan Standard Time","Atlantic Standard Time","AUS Central Standard Time","Aus Central W. Standard Time","AUS Eastern Standard Time","Azerbaijan Standard Time","Azores Standard Time","Bahia Standard Time","Bangladesh Standard Time","Belarus Standard Time","Bougainville Standard Time","Canada Central Standard Time","Cape Verde Standard Time","Caucasus Standard Time","Cen. Australia Standard Time","Central America Standard Time","Central Asia Standard Time","Central Brazilian Standard Time","Central Europe Standard Time","Central European Standard Time","Central Pacific Standard Time","Central Standard Time","Central Standard Time (Mexico)","Chatham Islands Standard Time","China Standard Time","Cuba Standard Time","Dateline Standard Time","E. Africa Standard Time","E. Australia Standard Time","E. Europe Standard Time","E. South America Standard Time","Easter Island Standard Time","Eastern Standard Time","Eastern Standard Time (Mexico)","Egypt Standard Time","Ekaterinburg Standard Time","Fiji Standard Time","FLE Standard Time","Georgian Standard Time","GMT Standard Time","Greenland Standard Time","Greenwich Standard Time","GTB Standard Time","Haiti Standard Time","Hawaiian Standard Time","India Standard Time","Iran Standard Time","Israel Standard Time","Jordan Standard Time","Kaliningrad Standard Time","Kamchatka Standard Time","Korea Standard Time","Libya Standard Time","Line Islands Standard Time","Lord Howe Standard Time","Magadan Standard Time","Magallanes Standard Time","Marquesas Standard Time","Mauritius Standard Time","Mid-Atlantic Standard Time","Middle East Standard Time","Montevideo Standard Time","Morocco Standard Time","Mountain Standard Time","Mountain Standard Time (Mexico)","Myanmar Standard Time","N. Central Asia Standard Time","Namibia Standard Time","Nepal Standard Time","New Zealand Standard Time","Newfoundland Standard Time","Norfolk Standard Time","North Asia East Standard Time","North Asia Standard Time","North Korea Standard Time","Omsk Standard Time","Pacific SA Standard Time","Pacific Standard Time","Pacific Standard Time (Mexico)","Pakistan Standard Time","Paraguay Standard Time","Qyzylorda Standard Time","Romance Standard Time","Russia Time Zone 10","Russia Time Zone 11","Russia Time Zone 3","Russian Standard Time","SA Eastern Standard Time","SA Pacific Standard Time","SA Western Standard Time","Saint Pierre Standard Time","Sakhalin Standard Time","Samoa Standard Time","Sao Tome Standard Time","Saratov Standard Time","SE Asia Standard Time","Singapore Standard Time","South Africa Standard Time","South Sudan Standard Time","Sri Lanka Standard Time","Sudan Standard Time","Syria Standard Time","Taipei Standard Time","Tasmania Standard Time","Tocantins Standard Time","Tokyo Standard Time","Tomsk Standard Time","Tonga Standard Time","Transbaikal Standard Time","Turkey Standard Time","Turks And Caicos Standard Time","Ulaanbaatar Standard Time","US Eastern Standard Time","US Mountain Standard Time","UTC","UTC-02","UTC-08","UTC-09","UTC-11","UTC+12","UTC+13","Venezuela Standard Time","Vladivostok Standard Time","Volgograd Standard Time","W. Australia Standard Time","W. Central Africa Standard Time","W. Europe Standard Time","W. Mongolia Standard Time","West Asia Standard Time","West Bank Standard Time","West Pacific Standard Time","Yakutsk Standard Time","Yukon Standard Time"]}}

### spec.postgresqlFlexibleServer

`AzureDataProtectionBackupPolicyPostgresqlFlexibleServer`

The PostgreSQL flexible-server variant: scheduled full backups
kept on the vault tier.

### spec.postgresqlFlexibleServer.backupRepeatingTimeIntervals

`[]string` · required

When backups run: ISO-8601 repeating intervals, e.g.
"R/2024-01-01T00:00:00+00:00/P1W" (weekly). At least one interval
is required.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"dpbp_pg_interval_iso8601_repeating","message":"each interval must be an ISO-8601 repeating interval, e.g. R/2024-01-01T00:00:00+00:00/P1W","expression":"this.matches('^R/.+/P((\\\\d+Y)?(\\\\d+M)?(\\\\d+W)?(\\\\d+D)?)(T(\\\\d+H)?(\\\\d+M)?(\\\\d+S)?)?$')"}]}}}

### spec.postgresqlFlexibleServer.defaultRetentionRule

`AzureDataProtectionBackupPolicyFlexibleServerDefaultRetentionRule` · required

The default retention applied to every backup no rule tags.

- rule: {"required":true}

### spec.postgresqlFlexibleServer.defaultRetentionRule.lifeCycles

`[]AzureDataProtectionBackupPolicyFlexibleServerLifeCycle` · required

How long default-retained backups live, per store.

- rule: {"repeated":{"minItems":"1"}}

### spec.postgresqlFlexibleServer.defaultRetentionRule.lifeCycles[].dataStoreType

`string` · required

The store the retention applies to. "VaultStore" is the only
value the service accepts today (the provider mirrors the service
team's own statement; the vocabulary widens if ArchiveStore ever
arrives).

- rule: {"required":true,"string":{"in":["VaultStore"]}}

### spec.postgresqlFlexibleServer.defaultRetentionRule.lifeCycles[].duration

`string` · required

The retention duration, ISO-8601 (e.g. "P3M").

- rule: duration must be an ISO-8601 duration, e.g. P3M
- rule: {"required":true}

### spec.postgresqlFlexibleServer.retentionRules

`[]AzureDataProtectionBackupPolicyFlexibleServerRetentionRule`

Named retention rules that keep SPECIFIC backups longer than the
default.

### spec.postgresqlFlexibleServer.retentionRules[].name

`string` · required

The rule's name (also the tag ARM stamps on the backups it
keeps).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.postgresqlFlexibleServer.retentionRules[].criteria

`AzureDataProtectionBackupPolicyFlexibleServerCriteria` · required

WHICH backups this rule keeps.

- rule: {"required":true}

### spec.postgresqlFlexibleServer.retentionRules[].criteria.absoluteCriteria

`string`

An absolute marker (the wire values).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AllBackup","FirstOfDay","FirstOfWeek","FirstOfMonth","FirstOfYear"]}}

### spec.postgresqlFlexibleServer.retentionRules[].criteria.daysOfWeek

`[]string`

Days of the week, e.g. "Sunday".

- rule: {"repeated":{"items":{"string":{"in":["Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"]}}}}

### spec.postgresqlFlexibleServer.retentionRules[].criteria.monthsOfYear

`[]string`

Months of the year, e.g. "January".

- rule: {"repeated":{"items":{"string":{"in":["January","February","March","April","May","June","July","August","September","October","November","December"]}}}}

### spec.postgresqlFlexibleServer.retentionRules[].criteria.scheduledBackupTimes

`[]string`

Specific backup times as RFC-3339 timestamps (only the time of
day is honored), e.g. "2024-01-01T05:30:00Z".

### spec.postgresqlFlexibleServer.retentionRules[].criteria.weeksOfMonth

`[]string`

Weeks of the month (the wire values).

- rule: {"repeated":{"items":{"string":{"in":["First","Second","Third","Fourth","Last"]}}}}

### spec.postgresqlFlexibleServer.retentionRules[].lifeCycles

`[]AzureDataProtectionBackupPolicyFlexibleServerLifeCycle` · required

HOW LONG the matched backups are kept, per store.

- rule: {"repeated":{"minItems":"1"}}

### spec.postgresqlFlexibleServer.retentionRules[].lifeCycles[].dataStoreType

`string` · required

The store the retention applies to. "VaultStore" is the only
value the service accepts today (the provider mirrors the service
team's own statement; the vocabulary widens if ArchiveStore ever
arrives).

- rule: {"required":true,"string":{"in":["VaultStore"]}}

### spec.postgresqlFlexibleServer.retentionRules[].lifeCycles[].duration

`string` · required

The retention duration, ISO-8601 (e.g. "P3M").

- rule: duration must be an ISO-8601 duration, e.g. P3M
- rule: {"required":true}

### spec.postgresqlFlexibleServer.retentionRules[].priority

`int32` · required · optional (explicit presence)

The rule's priority -- when several rules tag the same backup,
the LOWEST priority number wins. Required (the provider's own
contract); modeled with explicit presence so 0 stays expressible.

- rule: {"required":true}

### spec.postgresqlFlexibleServer.timeZone

`string`

The Windows time zone the schedule is interpreted in, e.g. "UTC".
Unspecified leaves the service interpretation (UTC).

### spec.dataLakeStorage

`AzureDataProtectionBackupPolicyDataLakeStorage`

The Data Lake storage variant: scheduled backups of a
hierarchical-namespace storage account's ADLS blob services,
kept on the vault tier.

### spec.dataLakeStorage.backupSchedule

`[]string` · required

When backups run: 1-5 ISO-8601 repeating intervals, e.g.
"R/2024-01-01T00:00:00+00:00/P1D" (daily).

- rule: {"repeated":{"minItems":"1","maxItems":"5","items":{"cel":[{"id":"dpbp_adls_schedule_iso8601_repeating","message":"each backup_schedule entry must be an ISO-8601 repeating interval, e.g. R/2024-01-01T00:00:00+00:00/P1D","expression":"this.matches('^R/.+/P((\\\\d+Y)?(\\\\d+M)?(\\\\d+W)?(\\\\d+D)?)(T(\\\\d+H)?(\\\\d+M)?(\\\\d+S)?)?$')"}]}}}

### spec.dataLakeStorage.defaultRetentionDuration

`string` · required

How long backups are kept by default, as an ISO-8601 duration
(e.g. "P30D").

- rule: default_retention_duration must be an ISO-8601 duration, e.g. P30D
- rule: {"required":true}

### spec.dataLakeStorage.retentionRules

`[]AzureDataProtectionBackupPolicyDataLakeStorageRetentionRule`

Named retention rules that keep SPECIFIC backups longer than the
default, in priority order (first = highest priority).

- rule: each Data Lake retention rule requires absolute_criteria or days_of_week (the provider's own contract)

### spec.dataLakeStorage.retentionRules[].name

`string` · required

The rule's name (also the tag ARM stamps on the backups it
keeps).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dataLakeStorage.retentionRules[].duration

`string` · required

The retention duration for matched backups, ISO-8601 (e.g.
"P12W"). Data Lake retains on the vault store (the provider pins
it).

- rule: duration must be an ISO-8601 duration, e.g. P12W
- rule: {"required":true}

### spec.dataLakeStorage.retentionRules[].absoluteCriteria

`string`

An absolute marker (the wire values).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AllBackup","FirstOfDay","FirstOfWeek","FirstOfMonth","FirstOfYear"]}}

### spec.dataLakeStorage.retentionRules[].daysOfWeek

`[]string`

Days of the week, e.g. "Sunday".

- rule: {"repeated":{"items":{"string":{"in":["Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"]}}}}

### spec.dataLakeStorage.retentionRules[].monthsOfYear

`[]string`

Months of the year, e.g. "January".

- rule: {"repeated":{"items":{"string":{"in":["January","February","March","April","May","June","July","August","September","October","November","December"]}}}}

### spec.dataLakeStorage.retentionRules[].scheduledBackupTimes

`[]string`

Specific backup times as RFC-3339 timestamps (only the time of
day is honored), e.g. "2024-01-01T05:30:00Z".

### spec.dataLakeStorage.retentionRules[].weeksOfMonth

`[]string`

Weeks of the month (the wire values).

- rule: {"repeated":{"items":{"string":{"in":["First","Second","Third","Fourth","Last"]}}}}

### spec.dataLakeStorage.timeZone

`string`

The Windows time zone the schedule is interpreted in (the
provider validates this variant against the Windows time-zone
catalog, which here includes "Coordinated Universal Time").
Unspecified leaves the service interpretation (UTC).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Afghanistan Standard Time","Alaskan Standard Time","Aleutian Standard Time","Altai Standard Time","Arab Standard Time","Arabian Standard Time","Arabic Standard Time","Argentina Standard Time","Astrakhan Standard Time","Atlantic Standard Time","AUS Central Standard Time","Aus Central W. Standard Time","AUS Eastern Standard Time","Azerbaijan Standard Time","Azores Standard Time","Bahia Standard Time","Bangladesh Standard Time","Belarus Standard Time","Bougainville Standard Time","Canada Central Standard Time","Cape Verde Standard Time","Caucasus Standard Time","Cen. Australia Standard Time","Central America Standard Time","Central Asia Standard Time","Central Brazilian Standard Time","Central Europe Standard Time","Central European Standard Time","Central Pacific Standard Time","Central Standard Time","Central Standard Time (Mexico)","Chatham Islands Standard Time","China Standard Time","Coordinated Universal Time","Cuba Standard Time","Dateline Standard Time","E. Africa Standard Time","E. Australia Standard Time","E. Europe Standard Time","E. South America Standard Time","Easter Island Standard Time","Eastern Standard Time","Eastern Standard Time (Mexico)","Egypt Standard Time","Ekaterinburg Standard Time","Fiji Standard Time","FLE Standard Time","Georgian Standard Time","GMT Standard Time","Greenland Standard Time","Greenwich Standard Time","GTB Standard Time","Haiti Standard Time","Hawaiian Standard Time","India Standard Time","Iran Standard Time","Israel Standard Time","Jordan Standard Time","Kaliningrad Standard Time","Kamchatka Standard Time","Korea Standard Time","Libya Standard Time","Line Islands Standard Time","Lord Howe Standard Time","Magadan Standard Time","Magallanes Standard Time","Marquesas Standard Time","Mauritius Standard Time","Mid-Atlantic Standard Time","Middle East Standard Time","Montevideo Standard Time","Morocco Standard Time","Mountain Standard Time","Mountain Standard Time (Mexico)","Myanmar Standard Time","N. Central Asia Standard Time","Namibia Standard Time","Nepal Standard Time","New Zealand Standard Time","Newfoundland Standard Time","Norfolk Standard Time","North Asia East Standard Time","North Asia Standard Time","North Korea Standard Time","Omsk Standard Time","Pacific SA Standard Time","Pacific Standard Time","Pacific Standard Time (Mexico)","Pakistan Standard Time","Paraguay Standard Time","Qyzylorda Standard Time","Romance Standard Time","Russia Time Zone 10","Russia Time Zone 11","Russia Time Zone 3","Russian Standard Time","SA Eastern Standard Time","SA Pacific Standard Time","SA Western Standard Time","Saint Pierre Standard Time","Sakhalin Standard Time","Samoa Standard Time","Sao Tome Standard Time","Saratov Standard Time","SE Asia Standard Time","Singapore Standard Time","South Africa Standard Time","South Sudan Standard Time","Sri Lanka Standard Time","Sudan Standard Time","Syria Standard Time","Taipei Standard Time","Tasmania Standard Time","Tocantins Standard Time","Tokyo Standard Time","Tomsk Standard Time","Tonga Standard Time","Transbaikal Standard Time","Turkey Standard Time","Turks And Caicos Standard Time","Ulaanbaatar Standard Time","US Eastern Standard Time","US Mountain Standard Time","UTC","UTC-02","UTC-08","UTC-09","UTC-11","UTC+12","UTC+13","Venezuela Standard Time","Vladivostok Standard Time","Volgograd Standard Time","W. Australia Standard Time","W. Central Africa Standard Time","W. Europe Standard Time","W. Mongolia Standard Time","West Asia Standard Time","West Bank Standard Time","West Pacific Standard Time","Yakutsk Standard Time","Yukon Standard Time"]}}

## Validation Rules

- `exactly_one_variant`: exactly one of blob_storage, disk, kubernetes_cluster, mysql_flexible_server, postgresql_flexible_server or data_lake_storage must be set -- the block is the datasource type
- `data_lake_name_starts_with_letter`: the data_lake_storage variant requires the policy name to start with a letter (the provider's own rule for this datasource)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataProtectionBackupPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.backup_policy_id` | `string` | The Azure Resource Manager ID of the policy -- what backup instances bind their policy by. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.DataProtection/backupVaults/{vault}/backupPolicies/{name} |
| `status.outputs.backup_policy_name` | `string` | The policy's name, unique on its vault. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vaultId` | AzureDataProtectionBackupVault | `status.outputs.backup_vault_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataProtectionBackupInstance | `spec.backupPolicyId` | `status.outputs.backup_policy_id` |

## See Also

- [Overview](../README.md)
