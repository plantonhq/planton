# AzureBackupPolicyFileShare

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureBackupPolicyFileShareSpec** defines an Azure Backup policy
for Azure Files shares (ARM: Microsoft.RecoveryServices/vaults/
{vault}/backupPolicies/{name}) -- WHEN shares under it are backed
up (the schedule) and HOW LONG each backup is kept (the retention
rules, layered daily/weekly/monthly/yearly like
grandfather-father-son tape rotation). The policy itself is a free
configuration object; cost follows the protected shares and their
backup storage.

**File-share schedules run Daily or Hourly only** (unlike VM
policies there is no Weekly schedule -- the provider's own
surface). A Daily schedule names its time of day; an Hourly
schedule configures a window (start, interval, duration) instead.
retention_daily is ALWAYS required -- it is the base layer both
frequencies retain into.

**backup_tier picks where backups live**: "snapshot" (the default)
keeps share snapshots in the storage account itself -- fast
restores, but the data shares the account's fate; "vault-standard"
ADDITIONALLY copies backups into the vault (protection against
account deletion/compromise), with snapshot_retention_in_days
governing how long the local snapshots persist alongside.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: an HOURLY
# schedule (the window dials -- interval, start, duration), the
# vault-standard tier with its local snapshot retention (strictly
# below the daily count), and all four retention layers -- monthly in
# the month-days form and yearly in the week-of-month form, so both
# exclusive grammars render.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBackupPolicyFileShare
metadata:
  name: test-backup-policy-file-share
  org: test-org
  env: dev
spec:
  resourceGroup:
    value: test-rg
  recoveryVaultName:
    value: test-backup-vault
  name: hourly-vault-standard-policy
  backup:
    frequency: Hourly
    hourly:
      interval: 4
      startTime: "06:00"
      windowDuration: 12
  backupTier: vault-standard
  snapshotRetentionInDays: 5
  timezone: UTC
  retentionDaily:
    count: 30
  retentionWeekly:
    count: 12
    weekdays: [Sunday]
  retentionMonthly:
    count: 12
    days: [1, 15]
    includeLastDays: true
  retentionYearly:
    count: 10
    months: [January]
    weeks: [First]
    weekdays: [Sunday]
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.recoveryVaultName` | `string \| valueFrom` | yes |  | AzureRecoveryServicesVault (`status.outputs.recovery_services_vault_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.backup` | `AzureBackupPolicyFileShareSchedule` | yes |  |  |
| `spec.backup.frequency` | `string` | yes |  |  |
| `spec.backup.time` | `string` |  |  |  |
| `spec.backup.hourly` | `AzureBackupPolicyFileShareHourlySchedule` |  |  |  |
| `spec.backup.hourly.interval` | `int32` | yes |  |  |
| `spec.backup.hourly.startTime` | `string` | yes |  |  |
| `spec.backup.hourly.windowDuration` | `int32` | yes |  |  |
| `spec.backupTier` | `string` |  | `snapshot` |  |
| `spec.snapshotRetentionInDays` | `int32` |  |  |  |
| `spec.timezone` | `string` |  | `UTC` |  |
| `spec.retentionDaily` | `AzureBackupPolicyFileShareRetentionDaily` | yes |  |  |
| `spec.retentionDaily.count` | `int32` | yes |  |  |
| `spec.retentionWeekly` | `AzureBackupPolicyFileShareRetentionWeekly` |  |  |  |
| `spec.retentionWeekly.count` | `int32` | yes |  |  |
| `spec.retentionWeekly.weekdays` | `[]string` | yes |  |  |
| `spec.retentionMonthly` | `AzureBackupPolicyFileShareRetentionMonthly` |  |  |  |
| `spec.retentionMonthly.count` | `int32` | yes |  |  |
| `spec.retentionMonthly.weeks` | `[]string` |  |  |  |
| `spec.retentionMonthly.weekdays` | `[]string` |  |  |  |
| `spec.retentionMonthly.days` | `[]int32` |  |  |  |
| `spec.retentionMonthly.includeLastDays` | `bool` |  |  |  |
| `spec.retentionYearly` | `AzureBackupPolicyFileShareRetentionYearly` |  |  |  |
| `spec.retentionYearly.count` | `int32` | yes |  |  |
| `spec.retentionYearly.months` | `[]string` | yes |  |  |
| `spec.retentionYearly.weeks` | `[]string` |  |  |  |
| `spec.retentionYearly.weekdays` | `[]string` |  |  |  |
| `spec.retentionYearly.days` | `[]int32` |  |  |  |
| `spec.retentionYearly.includeLastDays` | `bool` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the policy's VAULT lives in (a policy is
an ARM child of its vault). Can be a literal resource-group name
or a reference to an AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.recoveryVaultName

`string | valueFrom` · required

The Recovery Services vault the policy lives in, by NAME (ARM
addresses backup policies as children of a vault). Fixed at
creation.

- references: AzureRecoveryServicesVault (`status.outputs.recovery_services_vault_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureRecoveryServicesVault, name: <that resource's name>, fieldPath: status.outputs.recovery_services_vault_name}} -- a bare string does not parse

### spec.name

`string` · required

The policy's name, unique on the vault: 3-150 characters,
starting with a letter; letters, digits, hyphens, underscores and
'!' (the provider's own rule). Changing the name replaces the
policy.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z][-_!a-zA-Z0-9]{2,149}$"}}

### spec.backup

`AzureBackupPolicyFileShareSchedule` · required

When backups run.

- rule: {"required":true}
- rule: a Daily schedule requires time (and no hourly block); an Hourly schedule requires the hourly block (and no time)

### spec.backup.frequency

`string` · required

How often backups run (the wire values). File-share policies have
no Weekly schedule -- the provider's own surface.

- rule: {"required":true,"string":{"in":["Daily","Hourly"]}}

### spec.backup.time

`string`

For Daily schedules ONLY: the time of day the backup runs,
"HH:mm" on the hour or half past (e.g. "23:00", "02:30") -- the
provider's own rule.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([01][0-9]|[2][0-3]):([03][0])$"}}

### spec.backup.hourly

`AzureBackupPolicyFileShareHourlySchedule`

For Hourly schedules ONLY: the backup window (start, interval,
duration).

### spec.backup.hourly.interval

`int32` · required

Hours between backups inside the window -- 4, 6, 8 or 12 (the
provider's own set).

- rule: {"required":true,"int32":{"in":[4,6,8,12]}}

### spec.backup.hourly.startTime

`string` · required

The window's start time, "HH:mm" on the hour or half past (e.g.
"06:00", "20:30") -- the provider's own rule.

- rule: {"required":true,"string":{"pattern":"^([01][0-9]|[2][0-3]):([03][0])$"}}

### spec.backup.hourly.windowDuration

`int32` · required

The window's length in hours: 4-24.

- rule: {"required":true,"int32":{"lte":24,"gte":4}}

### spec.backupTier

`string` · optional (explicit presence)

Where backups live (the wire values). "snapshot" keeps share
snapshots in the storage account only; "vault-standard" also
copies backups into the vault (survives storage-account deletion
and compromise -- and unlocks snapshot_retention_in_days).
Unspecified applies snapshot (the provider's default).

- default: `snapshot`
- rule: {"string":{"in":["snapshot","vault-standard"]}}

### spec.snapshotRetentionInDays

`int32` · optional (explicit presence)

For vault-standard ONLY: how many days the LOCAL share snapshots
are kept alongside the vaulted copies (fast restores without a
vault round-trip). Must be LESS THAN retention_daily.count (the
provider's own contract). Unspecified lets the service manage
snapshot retention.

- rule: {"int32":{"gte":1}}

### spec.timezone

`string` · optional (explicit presence)

The IANA/Windows timezone the schedule's time is interpreted in,
e.g. "UTC" or "Pacific Standard Time". Unspecified applies UTC.

- default: `UTC`

### spec.retentionDaily

`AzureBackupPolicyFileShareRetentionDaily` · required

How long each daily/hourly backup is kept: ALWAYS required (the
base retention layer for both frequencies).

- rule: {"required":true}

### spec.retentionDaily.count

`int32` · required

Days each backup is kept: 1-200 (file-share retention runs
shorter than VM retention -- the provider's own bounds).

- rule: {"required":true,"int32":{"lte":200,"gte":1}}

### spec.retentionWeekly

`AzureBackupPolicyFileShareRetentionWeekly`

Keeps one backup per configured weekday for N weeks.

### spec.retentionWeekly.count

`int32` · required

Weeks each kept backup is retained: 1-200.

- rule: {"required":true,"int32":{"lte":200,"gte":1}}

### spec.retentionWeekly.weekdays

`[]string` · required

The weekdays whose backups are kept (wire values).

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["Sunday","Monday","Tuesday","Wednesday","Thursday","Friday","Saturday"]}}}}

### spec.retentionMonthly

`AzureBackupPolicyFileShareRetentionMonthly`

Keeps one backup per month (chosen by week-of-month days or by
month days) for N months.

- rule: pick ONE form: weeks+weekdays (week-of-month) or days/include_last_days (month days) -- they are mutually exclusive
- rule: configure which backup each month keeps: weeks+weekdays, days, or include_last_days
- rule: weeks and weekdays go together -- 'First Sunday' needs both the week and the weekday

### spec.retentionMonthly.count

`int32` · required

Months each kept backup is retained: 1-120 (the provider's own
bound for file shares).

- rule: {"required":true,"int32":{"lte":120,"gte":1}}

### spec.retentionMonthly.weeks

`[]string`

Week-of-month form: which weeks (wire values First, Second,
Third, Fourth, Last). Requires weekdays.

- rule: {"repeated":{"items":{"string":{"in":["First","Second","Third","Fourth","Last"]}}}}

### spec.retentionMonthly.weekdays

`[]string`

Week-of-month form: which weekdays inside those weeks (wire
values). Requires weeks.

- rule: {"repeated":{"items":{"string":{"in":["Sunday","Monday","Tuesday","Wednesday","Thursday","Friday","Saturday"]}}}}

### spec.retentionMonthly.days

`[]int32`

Month-days form: which days of the month (1-31).

- rule: {"repeated":{"items":{"int32":{"lte":31,"gte":1}}}}

### spec.retentionMonthly.includeLastDays

`bool`

Month-days form: whether the month's LAST day's backup is kept
(months differ in length; this covers "the last day" exactly).

### spec.retentionYearly

`AzureBackupPolicyFileShareRetentionYearly`

Keeps one backup per year (chosen months, then week-of-month days
or month days) for N years.

- rule: pick ONE form: weeks+weekdays (week-of-month) or days/include_last_days (month days) -- they are mutually exclusive
- rule: configure which backup each year keeps: weeks+weekdays, days, or include_last_days
- rule: weeks and weekdays go together -- 'First Sunday' needs both the week and the weekday

### spec.retentionYearly.count

`int32` · required

Years each kept backup is retained: 1-10 (the provider's own
bound for file shares).

- rule: {"required":true,"int32":{"lte":10,"gte":1}}

### spec.retentionYearly.months

`[]string` · required

The months whose backup is kept (wire values, e.g. "January").

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["January","February","March","April","May","June","July","August","September","October","November","December"]}}}}

### spec.retentionYearly.weeks

`[]string`

Week-of-month form: which weeks (wire values). Requires weekdays.

- rule: {"repeated":{"items":{"string":{"in":["First","Second","Third","Fourth","Last"]}}}}

### spec.retentionYearly.weekdays

`[]string`

Week-of-month form: which weekdays inside those weeks (wire
values). Requires weeks.

- rule: {"repeated":{"items":{"string":{"in":["Sunday","Monday","Tuesday","Wednesday","Thursday","Friday","Saturday"]}}}}

### spec.retentionYearly.days

`[]int32`

Month-days form: which days of the month (1-31).

- rule: {"repeated":{"items":{"int32":{"lte":31,"gte":1}}}}

### spec.retentionYearly.includeLastDays

`bool`

Month-days form: whether the month's LAST day's backup is kept.

## Validation Rules

- `bpfs_snapshot_retention_needs_vault_standard`: snapshot_retention_in_days only applies when backup_tier is vault-standard -- on the snapshot tier, retention_daily IS the snapshot retention
- `bpfs_snapshot_retention_below_daily`: snapshot_retention_in_days must be less than retention_daily.count (the provider's own contract)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureBackupPolicyFileShare, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.backup_policy_id` | `string` | The Azure Resource Manager ID of the backup policy. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.RecoveryServices/vaults/{vault}/backupPolicies/{name} |
| `status.outputs.backup_policy_name` | `string` | The policy's name -- unique on its vault. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.recoveryVaultName` | AzureRecoveryServicesVault | `status.outputs.recovery_services_vault_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureBackupProtectedFileShare | `spec.backupPolicyId` | `status.outputs.backup_policy_id` |

## See Also

- [Overview](../README.md)
