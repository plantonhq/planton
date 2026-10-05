# AzureBackupPolicyVm

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureBackupPolicyVmSpec** defines an Azure Backup policy for IaaS
virtual machines (ARM: Microsoft.RecoveryServices/vaults/{vault}/
backupPolicies/{name}) -- WHEN VMs under it are backed up (the
schedule) and HOW LONG each backup is kept (the retention rules,
layered daily/weekly/monthly/yearly like grandfather-father-son
tape rotation). The policy itself is a free configuration object;
cost follows the protected VMs and their backup storage.

**The schedule's frequency decides which retention layers are
legal** (the provider's own contract, front-loaded here as
validation): Hourly and Daily schedules require retention_daily;
a Weekly schedule requires retention_weekly and forbids
retention_daily. Hourly schedules exist only on V2 (enhanced)
policies.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: a V2 (enhanced)
# policy with an HOURLY schedule (interval/duration dials), crash-only
# consistency (V2-gated), instant restore at a V2-only depth, the
# named instant-restore resource group, a TierAfter archive rule with
# its age, and all four retention layers -- monthly in the month-days
# form and yearly in the week-of-month form, so both exclusive
# grammars render.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBackupPolicyVm
metadata:
  name: test-backup-policy-vm
  org: test-org
  env: dev
spec:
  resourceGroup:
    value: test-rg
  recoveryVaultName:
    value: test-backup-vault
  name: hourly-enhanced-policy
  policyType: V2
  backup:
    frequency: Hourly
    time: "08:00"
    hourInterval: 4
    hourDuration: 12
  instantRestoreRetentionDays: 10
  instantRestoreResourceGroup:
    prefix: backup-snapshots
    suffix: dev
  tieringPolicy:
    archivedRestorePoint:
      mode: TierAfter
      duration: 3
      durationType: Months
  consistencyType: OnlyCrashConsistent
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
    count: 7
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
| `spec.policyType` | `string` |  | `V1` |  |
| `spec.backup` | `AzureBackupPolicyVmSchedule` | yes |  |  |
| `spec.backup.frequency` | `string` | yes |  |  |
| `spec.backup.time` | `string` | yes |  |  |
| `spec.backup.weekdays` | `[]string` |  |  |  |
| `spec.backup.hourInterval` | `int32` |  |  |  |
| `spec.backup.hourDuration` | `int32` |  |  |  |
| `spec.instantRestoreRetentionDays` | `int32` |  |  |  |
| `spec.instantRestoreResourceGroup` | `AzureBackupPolicyVmInstantRestoreResourceGroup` |  |  |  |
| `spec.instantRestoreResourceGroup.prefix` | `string` | yes |  |  |
| `spec.instantRestoreResourceGroup.suffix` | `string` |  |  |  |
| `spec.tieringPolicy` | `AzureBackupPolicyVmTieringPolicy` |  |  |  |
| `spec.tieringPolicy.archivedRestorePoint` | `AzureBackupPolicyVmArchivedRestorePoint` | yes |  |  |
| `spec.tieringPolicy.archivedRestorePoint.mode` | `string` | yes |  |  |
| `spec.tieringPolicy.archivedRestorePoint.duration` | `int32` |  |  |  |
| `spec.tieringPolicy.archivedRestorePoint.durationType` | `string` |  |  |  |
| `spec.consistencyType` | `string` |  |  |  |
| `spec.timezone` | `string` |  | `UTC` |  |
| `spec.retentionDaily` | `AzureBackupPolicyVmRetentionDaily` |  |  |  |
| `spec.retentionDaily.count` | `int32` | yes |  |  |
| `spec.retentionWeekly` | `AzureBackupPolicyVmRetentionWeekly` |  |  |  |
| `spec.retentionWeekly.count` | `int32` | yes |  |  |
| `spec.retentionWeekly.weekdays` | `[]string` | yes |  |  |
| `spec.retentionMonthly` | `AzureBackupPolicyVmRetentionMonthly` |  |  |  |
| `spec.retentionMonthly.count` | `int32` | yes |  |  |
| `spec.retentionMonthly.weeks` | `[]string` |  |  |  |
| `spec.retentionMonthly.weekdays` | `[]string` |  |  |  |
| `spec.retentionMonthly.days` | `[]int32` |  |  |  |
| `spec.retentionMonthly.includeLastDays` | `bool` |  |  |  |
| `spec.retentionYearly` | `AzureBackupPolicyVmRetentionYearly` |  |  |  |
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

### spec.policyType

`string` · optional (explicit presence)

The policy generation (the wire values). V1 is the classic
policy; V2 (an "enhanced" policy) adds Hourly schedules, zonally
resilient instant restore, and longer instant-restore retention.
Unspecified applies V1 (the provider's default). Fixed at
creation.

- default: `V1`
- rule: {"string":{"in":["V1","V2"]}}

### spec.backup

`AzureBackupPolicyVmSchedule` · required

When backups run.

- rule: {"required":true}
- rule: weekdays is required for a Weekly schedule and must be empty for Hourly and Daily schedules
- rule: hour_interval and hour_duration are required for an Hourly schedule and must be unset otherwise
- rule: hour_duration must be a multiple of hour_interval (e.g. interval 4 allows durations 4, 8, 12, 16, 20, 24)

### spec.backup.frequency

`string` · required

How often backups run (the wire values). Hourly needs a V2
policy; Weekly requires weekdays.

- rule: {"required":true,"string":{"in":["Hourly","Daily","Weekly"]}}

### spec.backup.time

`string` · required

The time of day backups start, "HH:mm" on the hour or half past
(e.g. "23:00", "02:30") -- the provider's own rule. For Hourly
schedules this is the backup WINDOW's start time.

- rule: {"required":true,"string":{"pattern":"^([01][0-9]|[2][0-3]):([03][0])$"}}

### spec.backup.weekdays

`[]string`

For Weekly schedules ONLY: the days backups run (wire values,
e.g. "Sunday", "Wednesday").

- rule: {"repeated":{"items":{"string":{"in":["Sunday","Monday","Tuesday","Wednesday","Thursday","Friday","Saturday"]}}}}

### spec.backup.hourInterval

`int32` · optional (explicit presence)

For Hourly schedules ONLY: hours between backups inside the
window -- 4, 6, 8 or 12 (the provider's own set).

- rule: {"int32":{"in":[4,6,8,12]}}

### spec.backup.hourDuration

`int32` · optional (explicit presence)

For Hourly schedules ONLY: the backup window's length in hours,
4-24; must be a multiple of hour_interval (the provider's own
contract).

- rule: {"int32":{"lte":24,"gte":4}}

### spec.instantRestoreRetentionDays

`int32` · optional (explicit presence)

How many days instant-restore snapshots are kept alongside the
VM (fast restores without vault round-trips): 1-30. On V1
policies the cap is 5 (the provider's own contract). Unspecified
applies the service default (2 for V1, 7 for V2).

Azure requires vaulted daily retention to be STRICTLY greater
than this value (BMSUserErrorInstantRPRetentionExceedsVaultedRetention
-- live-proven). A V2 policy that leaves this unset therefore
cannot use retention_daily.count of 7: set this field to 1-6, or
raise daily count above 7.

- rule: {"int32":{"lte":30,"gte":1}}

### spec.instantRestoreResourceGroup

`AzureBackupPolicyVmInstantRestoreResourceGroup`

Names the resource group Azure creates instant-restore snapshot
collections in (default: an AzureBackupRG_* group per region).

### spec.instantRestoreResourceGroup.prefix

`string` · required

The name prefix of the per-region snapshot resource group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.instantRestoreResourceGroup.suffix

`string`

An optional suffix appended after the region segment.

### spec.tieringPolicy

`AzureBackupPolicyVmTieringPolicy`

Moves aged recovery points to the cheaper archive tier.

### spec.tieringPolicy.archivedRestorePoint

`AzureBackupPolicyVmArchivedRestorePoint` · required

The archive rule for restore points.

- rule: {"required":true}
- rule: TierAfter requires duration and duration_type (the age after which restore points archive)
- rule: duration and duration_type only apply to TierAfter -- TierRecommended lets Azure pick the points

### spec.tieringPolicy.archivedRestorePoint.mode

`string` · required

The archival mode (the wire values). TierRecommended lets Azure
pick archivable points (cost-optimal); TierAfter archives every
point after the configured age -- duration and duration_type then
become required.

- rule: {"required":true,"string":{"in":["TierRecommended","TierAfter"]}}

### spec.tieringPolicy.archivedRestorePoint.duration

`int32` · optional (explicit presence)

For TierAfter: the age after which points archive. At least 3
(the provider's own floor).

- rule: {"int32":{"gte":3}}

### spec.tieringPolicy.archivedRestorePoint.durationType

`string`

For TierAfter: the age's unit (the wire values).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Days","Weeks","Months","Years"]}}

### spec.consistencyType

`string`

The snapshot consistency class (the wire value).
OnlyCrashConsistent trades application consistency for faster,
lighter snapshots -- V2 policies only (the provider's own
contract). Unspecified leaves the service default
(application/file-system consistent when possible).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["OnlyCrashConsistent"]}}

### spec.timezone

`string` · optional (explicit presence)

The IANA/Windows timezone the schedule's time is interpreted in,
e.g. "UTC" or "Pacific Standard Time". Unspecified applies UTC.

- default: `UTC`

### spec.retentionDaily

`AzureBackupPolicyVmRetentionDaily`

How long each daily/hourly backup is kept. REQUIRED for Hourly
and Daily schedules; must be absent for Weekly schedules.

### spec.retentionDaily.count

`int32` · required

Days each backup is kept: 1, or 7-9999. Azure REJECTS 2-6 daily
backups at create time (a service rule the provider surfaces only
at apply -- front-loaded here).

- rule: Azure no longer accepts 2-6 days of daily retention -- use 1, or 7 and above
- rule: {"required":true,"int32":{"lte":9999,"gte":1}}

### spec.retentionWeekly

`AzureBackupPolicyVmRetentionWeekly`

Keeps one backup per configured weekday for N weeks. REQUIRED for
Weekly schedules; optional layering on Hourly/Daily schedules.

### spec.retentionWeekly.count

`int32` · required

Weeks each kept backup is retained: 1-9999.

- rule: {"required":true,"int32":{"lte":9999,"gte":1}}

### spec.retentionWeekly.weekdays

`[]string` · required

The weekdays whose backups are kept (wire values).

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["Sunday","Monday","Tuesday","Wednesday","Thursday","Friday","Saturday"]}}}}

### spec.retentionMonthly

`AzureBackupPolicyVmRetentionMonthly`

Keeps one backup per month (chosen by week-of-month days or by
month days) for N months.

- rule: pick ONE form: weeks+weekdays (week-of-month) or days/include_last_days (month days) -- they are mutually exclusive
- rule: configure which backup each month keeps: weeks+weekdays, days, or include_last_days
- rule: weeks and weekdays go together -- 'First Sunday' needs both the week and the weekday

### spec.retentionMonthly.count

`int32` · required

Months each kept backup is retained: 1-9999.

- rule: {"required":true,"int32":{"lte":9999,"gte":1}}

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

`AzureBackupPolicyVmRetentionYearly`

Keeps one backup per year (chosen months, then week-of-month days
or month days) for N years.

- rule: pick ONE form: weeks+weekdays (week-of-month) or days/include_last_days (month days) -- they are mutually exclusive
- rule: configure which backup each year keeps: weeks+weekdays, days, or include_last_days
- rule: weeks and weekdays go together -- 'First Sunday' needs both the week and the weekday

### spec.retentionYearly.count

`int32` · required

Years each kept backup is retained: 1-9999.

- rule: {"required":true,"int32":{"lte":9999,"gte":1}}

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

- `bpv_daily_retention_for_hourly_daily`: retention_daily is required when backup.frequency is Hourly or Daily
- `bpv_weekly_retention_for_weekly`: a Weekly schedule requires retention_weekly and must not set retention_daily (weekly backups have no daily layer)
- `bpv_hourly_requires_v2`: an Hourly schedule requires policy_type V2 (enhanced policy) -- V1 policies back up at most once a day
- `bpv_crash_consistent_requires_v2`: consistency_type OnlyCrashConsistent requires policy_type V2 (the provider's own contract)
- `bpv_instant_restore_v1_cap`: instant_restore_retention_days is capped at 5 on V1 policies -- use policy_type V2 for up to 30 days
- `bpv_instant_lt_daily`: instant_restore_retention_days must be less than retention_daily.count -- Azure rejects Instant RP that meets or exceeds vaulted daily retention (BMSUserErrorInstantRPRetentionExceedsVaultedRetention). V2 defaults Instant RP to 7 days when unset, so a daily count of 7 needs an explicit smaller instant value

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureBackupPolicyVm, name: <resource-name>, fieldPath: status.outputs.<output>}`.

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
| AzureBackupProtectedVm | `spec.backupPolicyId` | `status.outputs.backup_policy_id` |

## See Also

- [Overview](../README.md)
