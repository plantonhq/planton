# AwsDlmLifecyclePolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsDlmLifecyclePolicySpec defines one Data Lifecycle Manager policy
- account-level automation that creates, retains, copies, archives,
and deletes EBS snapshots (or EBS-backed AMIs) on a schedule,
selecting its targets by TAGS. The policy references no specific
volume or snapshot; it acts on whatever carries the target tags
when a schedule fires.

A policy runs in exactly one of two modes:

DEFAULT mode (default_policy set) - AWS's simplified account-wide
posture: "snapshot every volume (or instance) daily-ish, keep for
N days", with tag/type exclusions. One default policy per resource
type per region.

CUSTOM mode (custom_policy set) - the full engine: up to four named
schedules with cron or interval triggers, per-schedule retention,
archive tiering, cross-region copies, deprecation, fast snapshot
restore, and cross-account sharing; or an event-driven policy that
reacts to snapshots shared INTO this account.

The provider's policy_language argument is derived by the modules
(SIMPLIFIED for default mode, STANDARD for custom mode) - a
separate field could contradict the configured arm.

## Example

```yaml
# Canonical AwsDlmLifecyclePolicy example (hack/dev manifest and
# refgen Example source): a custom policy snapshotting tagged volumes
# daily, keeping a week.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsDlmLifecyclePolicy
metadata:
  name: daily-volume-backups
  id: daily-volume-backups
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: daily snapshots of tagged volumes
  executionRoleArn:
    value: arn:aws:iam::123456789012:role/AWSDataLifecycleManagerDefaultRole # replace with your DLM role
  customPolicy:
    resourceTypes:
      - VOLUME
    targetTags:
      backup: "true"
    schedules:
      - name: daily
        copyTags: true
        createRule:
          intervalHours: 24
          times:
            - "03:00"
        retainRule:
          count: 7
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.disabled` | `bool` |  |  |  |
| `spec.defaultPolicy` | `AwsDlmDefaultPolicy` |  |  |  |
| `spec.defaultPolicy.resourceType` | `string` |  |  |  |
| `spec.defaultPolicy.createIntervalDays` | `int64` |  |  |  |
| `spec.defaultPolicy.retainIntervalDays` | `int64` |  |  |  |
| `spec.defaultPolicy.copyTags` | `bool` |  |  |  |
| `spec.defaultPolicy.extendDeletion` | `bool` |  |  |  |
| `spec.defaultPolicy.exclusions` | `AwsDlmDefaultPolicyExclusions` |  |  |  |
| `spec.defaultPolicy.exclusions.excludeBootVolumes` | `bool` |  |  |  |
| `spec.defaultPolicy.exclusions.excludeTags` | `map<string, string>` |  |  |  |
| `spec.defaultPolicy.exclusions.excludeVolumeTypes` | `[]string` |  |  |  |
| `spec.customPolicy` | `AwsDlmCustomPolicy` |  |  |  |
| `spec.customPolicy.policyType` | `string` |  |  |  |
| `spec.customPolicy.resourceTypes` | `[]string` |  |  |  |
| `spec.customPolicy.resourceLocations` | `[]string` |  |  |  |
| `spec.customPolicy.targetTags` | `map<string, string>` |  |  |  |
| `spec.customPolicy.parameters` | `AwsDlmParameters` |  |  |  |
| `spec.customPolicy.parameters.excludeBootVolume` | `bool` |  |  |  |
| `spec.customPolicy.parameters.noReboot` | `bool` |  |  |  |
| `spec.customPolicy.schedules` | `[]AwsDlmSchedule` |  |  |  |
| `spec.customPolicy.schedules[].name` | `string` | yes |  |  |
| `spec.customPolicy.schedules[].copyTags` | `bool` |  |  |  |
| `spec.customPolicy.schedules[].tagsToAdd` | `map<string, string>` |  |  |  |
| `spec.customPolicy.schedules[].variableTags` | `map<string, string>` |  |  |  |
| `spec.customPolicy.schedules[].createRule` | `AwsDlmCreateRule` | yes |  |  |
| `spec.customPolicy.schedules[].createRule.intervalHours` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].createRule.times` | `[]string` |  |  |  |
| `spec.customPolicy.schedules[].createRule.cronExpression` | `string` |  |  |  |
| `spec.customPolicy.schedules[].createRule.location` | `string` |  |  |  |
| `spec.customPolicy.schedules[].createRule.scripts` | `AwsDlmScripts` |  |  |  |
| `spec.customPolicy.schedules[].createRule.scripts.executionHandler` | `string` | yes |  |  |
| `spec.customPolicy.schedules[].createRule.scripts.stages` | `[]string` |  |  |  |
| `spec.customPolicy.schedules[].createRule.scripts.executionHandlerService` | `string` |  |  |  |
| `spec.customPolicy.schedules[].createRule.scripts.executeOperationOnScriptFailure` | `bool` |  |  |  |
| `spec.customPolicy.schedules[].createRule.scripts.executionTimeoutSeconds` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].createRule.scripts.maximumRetryCount` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].retainRule` | `AwsDlmRetainRule` | yes |  |  |
| `spec.customPolicy.schedules[].retainRule.count` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].retainRule.interval` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].retainRule.intervalUnit` | `string` |  |  |  |
| `spec.customPolicy.schedules[].archiveRule` | `AwsDlmArchiveRule` |  |  |  |
| `spec.customPolicy.schedules[].archiveRule.count` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].archiveRule.interval` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].archiveRule.intervalUnit` | `string` |  |  |  |
| `spec.customPolicy.schedules[].crossRegionCopyRules` | `[]AwsDlmCrossRegionCopyRule` |  |  |  |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].targetRegion` | `string` | yes |  |  |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].encrypted` | `bool` |  |  |  |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].cmkArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].copyTags` | `bool` |  |  |  |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].retainRule` | `AwsDlmCopyRetainRule` |  |  |  |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].retainRule.interval` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].retainRule.intervalUnit` | `string` |  |  |  |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].deprecateRule` | `AwsDlmCopyRetainRule` |  |  |  |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].deprecateRule.interval` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].deprecateRule.intervalUnit` | `string` |  |  |  |
| `spec.customPolicy.schedules[].deprecateRule` | `AwsDlmDeprecateRule` |  |  |  |
| `spec.customPolicy.schedules[].deprecateRule.count` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].deprecateRule.interval` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].deprecateRule.intervalUnit` | `string` |  |  |  |
| `spec.customPolicy.schedules[].fastRestoreRule` | `AwsDlmFastRestoreRule` |  |  |  |
| `spec.customPolicy.schedules[].fastRestoreRule.availabilityZones` | `[]string` | yes |  |  |
| `spec.customPolicy.schedules[].fastRestoreRule.count` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].fastRestoreRule.interval` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].fastRestoreRule.intervalUnit` | `string` |  |  |  |
| `spec.customPolicy.schedules[].shareRule` | `AwsDlmShareRule` |  |  |  |
| `spec.customPolicy.schedules[].shareRule.targetAccounts` | `[]string` | yes |  |  |
| `spec.customPolicy.schedules[].shareRule.unshareInterval` | `int64` |  |  |  |
| `spec.customPolicy.schedules[].shareRule.unshareIntervalUnit` | `string` |  |  |  |
| `spec.customPolicy.eventSource` | `AwsDlmEventSource` |  |  |  |
| `spec.customPolicy.eventSource.eventType` | `string` |  |  |  |
| `spec.customPolicy.eventSource.descriptionRegex` | `string` | yes |  |  |
| `spec.customPolicy.eventSource.snapshotOwners` | `[]string` | yes |  |  |
| `spec.customPolicy.action` | `AwsDlmAction` |  |  |  |
| `spec.customPolicy.action.name` | `string` | yes |  |  |
| `spec.customPolicy.action.crossRegionCopies` | `[]AwsDlmActionCrossRegionCopy` | yes |  |  |
| `spec.customPolicy.action.crossRegionCopies[].target` | `string` | yes |  |  |
| `spec.customPolicy.action.crossRegionCopies[].encrypted` | `bool` |  |  |  |
| `spec.customPolicy.action.crossRegionCopies[].cmkArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.customPolicy.action.crossRegionCopies[].retainRule` | `AwsDlmCopyRetainRule` |  |  |  |
| `spec.customPolicy.action.crossRegionCopies[].retainRule.interval` | `int64` |  |  |  |
| `spec.customPolicy.action.crossRegionCopies[].retainRule.intervalUnit` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the policy runs in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

What this policy does, shown in the DLM console listing. Unset
means the modules use metadata.name. AWS restricts the character
set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"500","pattern":"^[0-9A-Za-z _-]+$"}}

### spec.executionRoleArn

`string | valueFrom` · required

The IAM role DLM assumes to create and manage snapshots/AMIs.
AWS ships a service default (AWSDataLifecycleManagerDefaultRole)
- create it once per account (aws dlm create-default-role) or
reference an AwsIamRole with the documented DLM permissions.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.disabled

`bool`

Pause the policy without deleting it - schedules stop firing,
existing snapshots are untouched.

### spec.defaultPolicy

`AwsDlmDefaultPolicy`

DEFAULT mode: the simplified account-wide snapshot posture.

- rule: exclusions.exclude_boot_volumes applies only when resource_type is VOLUME
- rule: exclusions.exclude_volume_types applies only when resource_type is VOLUME

### spec.defaultPolicy.resourceType

`string`

Snapshot VOLUMEs individually, or INSTANCEs as multi-volume
snapshot sets. One default policy per type per region.

- rule: {"string":{"in":["VOLUME","INSTANCE"]}}

### spec.defaultPolicy.createIntervalDays

`int64`

How often snapshots are created, in days (1-7). 0 means AWS's
default of 1 (daily).

- rule: create_interval_days must be between 1 and 7

### spec.defaultPolicy.retainIntervalDays

`int64`

How long snapshots are kept, in days (2-14). 0 means AWS's
default of 7.

- rule: retain_interval_days must be between 2 and 14

### spec.defaultPolicy.copyTags

`bool`

Copy the source resource's tags onto its snapshots.

### spec.defaultPolicy.extendDeletion

`bool`

Keep snapshots even after their source resource is deleted or
the policy stops targeting it (default: cleanup extends to them).

### spec.defaultPolicy.exclusions

`AwsDlmDefaultPolicyExclusions`

Carve-outs: resources the default policy must skip.

### spec.defaultPolicy.exclusions.excludeBootVolumes

`bool`

Skip boot volumes (VOLUME policies only) - data-volume-only
backup postures.

### spec.defaultPolicy.exclusions.excludeTags

`map<string, string>`

Skip resources carrying any of these tags.

### spec.defaultPolicy.exclusions.excludeVolumeTypes

`[]string`

Skip these volume types (VOLUME policies only). Example:
["standard", "sc1"].

- rule: {"repeated":{"maxItems":"6"}}

### spec.customPolicy

`AwsDlmCustomPolicy`

CUSTOM mode: named schedules or an event-driven policy.

- rule: EBS_SNAPSHOT_MANAGEMENT and IMAGE_MANAGEMENT policies select resources by tags - target_tags is required
- rule: EBS_SNAPSHOT_MANAGEMENT and IMAGE_MANAGEMENT policies carry 1 to 4 schedules
- rule: an EVENT_BASED_POLICY requires event_source and action (and takes no target_tags or schedules)
- rule: event_source and action belong to EVENT_BASED_POLICY policies
- rule: parameters (exclude_boot_volume / no_reboot) apply to policies whose resource_types is ["INSTANCE"]
- rule: schedule names must be unique

### spec.customPolicy.policyType

`string`

What the policy manages. EBS_SNAPSHOT_MANAGEMENT (the default
when unset) creates snapshots; IMAGE_MANAGEMENT creates
EBS-backed AMIs; EVENT_BASED_POLICY copies snapshots shared into
this account.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["EBS_SNAPSHOT_MANAGEMENT","IMAGE_MANAGEMENT","EVENT_BASED_POLICY"]}}

### spec.customPolicy.resourceTypes

`[]string`

What the schedules snapshot: VOLUME (individual volumes) or
INSTANCE (multi-volume snapshot sets / AMIs). IMAGE_MANAGEMENT
policies are implicitly INSTANCE.

- rule: {"repeated":{"maxItems":"1","items":{"string":{"in":["VOLUME","INSTANCE"]}}}}

### spec.customPolicy.resourceLocations

`[]string`

Where targeted resources live: CLOUD (the region - the default
when unset), OUTPOST, or LOCAL_ZONE.

- rule: {"repeated":{"maxItems":"1","items":{"string":{"in":["CLOUD","OUTPOST","LOCAL_ZONE"]}}}}

### spec.customPolicy.targetTags

`map<string, string>`

The tags that select target resources - a resource matching ANY
entry is in scope for every schedule.

### spec.customPolicy.parameters

`AwsDlmParameters`

Instance-policy dials (resource_types = ["INSTANCE"]).

### spec.customPolicy.parameters.excludeBootVolume

`bool`

Snapshot only the data volumes - skip the boot volume.

### spec.customPolicy.parameters.noReboot

`bool`

Create the AMI without rebooting the instance (IMAGE_MANAGEMENT).
No-reboot images can be filesystem-inconsistent; the default
(reboot) is the safe choice.

### spec.customPolicy.schedules

`[]AwsDlmSchedule`

The named schedules (1-4). The first schedule is the mandatory
one; extras add cadences (e.g. daily + weekly + monthly tiers).

### spec.customPolicy.schedules[].name

`string` · required

The schedule's key and console label. Fixed choices below are
create-only at AWS: copy_tags replaces the whole schedule on
change.

- rule: {"string":{"minLen":"1","maxLen":"120"}}

### spec.customPolicy.schedules[].copyTags

`bool`

Copy the source's tags onto created snapshots/AMIs. Changing this
replaces the schedule.

### spec.customPolicy.schedules[].tagsToAdd

`map<string, string>`

Extra tags stamped onto every snapshot/AMI this schedule creates.

### spec.customPolicy.schedules[].variableTags

`map<string, string>`

Tags with runtime variables: $(instance-id) and $(timestamp)
expand per snapshot (INSTANCE policies).

### spec.customPolicy.schedules[].createRule

`AwsDlmCreateRule` · required

When snapshots are created.

- rule: {"required":true}
- rule: configure exactly one of interval_hours and cron_expression
- rule: times (the HH:MM start time) accompanies interval_hours - a cron expression carries its own timing

### spec.customPolicy.schedules[].createRule.intervalHours

`int64`

Snapshot every N hours: 1, 2, 3, 4, 6, 8, 12, or 24.

- rule: interval_hours must be one of 1, 2, 3, 4, 6, 8, 12, 24

### spec.customPolicy.schedules[].createRule.times

`[]string`

The time of day (UTC, "HH:MM") interval-based snapshots start.
At most one entry; unset lets AWS pick.

- rule: {"repeated":{"maxItems":"1","items":{"string":{"pattern":"^\\d{2}:\\d{2}$"}}}}

### spec.customPolicy.schedules[].createRule.cronExpression

`string`

A cron expression ("cron(0 3 * * ? *)") for cadences the interval
form cannot express - weekly, monthly, yearly.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"120","pattern":"^cron\\(.+\\)$"}}

### spec.customPolicy.schedules[].createRule.location

`string`

Where the schedule creates snapshots: CLOUD (the default when
unset), OUTPOST_LOCAL, or LOCAL_ZONE. Must match the policy's
resource_locations.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CLOUD","OUTPOST_LOCAL","LOCAL_ZONE"]}}

### spec.customPolicy.schedules[].createRule.scripts

`AwsDlmScripts`

Run pre/post scripts around snapshot creation (application-
consistent snapshots via SSM).

### spec.customPolicy.schedules[].createRule.scripts.executionHandler

`string` · required

The SSM document that quiesces/resumes the application:
AWS_VSS_BACKUP (Windows VSS) or an SSM document name/ARN.

- rule: {"string":{"minLen":"1"}}

### spec.customPolicy.schedules[].createRule.scripts.stages

`[]string`

Which side(s) run: PRE (before the snapshot), POST (after), or
both. Unset means AWS's default for the handler.

- rule: {"repeated":{"maxItems":"2","items":{"string":{"in":["PRE","POST"]}}}}

### spec.customPolicy.schedules[].createRule.scripts.executionHandlerService

`string`

The handler's service. Unset means AWS_SYSTEMS_MANAGER (the only
published value).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AWS_SYSTEMS_MANAGER"]}}

### spec.customPolicy.schedules[].createRule.scripts.executeOperationOnScriptFailure

`bool`

Snapshot anyway when the script fails (default: the failure
blocks the snapshot).

### spec.customPolicy.schedules[].createRule.scripts.executionTimeoutSeconds

`int64`

Seconds the script may run (10-120). 0 means AWS's default of 10.

- rule: execution_timeout_seconds must be between 10 and 120

### spec.customPolicy.schedules[].createRule.scripts.maximumRetryCount

`int64`

Script retries (0-3).

- rule: {"int64":{"lte":"3","gte":"0"}}

### spec.customPolicy.schedules[].retainRule

`AwsDlmRetainRule` · required

How long they are kept.

- rule: {"required":true}
- rule: configure exactly one of count (keep the newest N) and interval + interval_unit (keep for a duration)
- rule: interval and interval_unit are set together

### spec.customPolicy.schedules[].retainRule.count

`int64`

Keep the newest N snapshots (1-1000); older ones are deleted as
new ones arrive.

- rule: {"int64":{"lte":"1000","gte":"0"}}

### spec.customPolicy.schedules[].retainRule.interval

`int64`

Keep each snapshot for this long, paired with interval_unit.

- rule: {"int64":{"gte":"0"}}

### spec.customPolicy.schedules[].retainRule.intervalUnit

`string`

The retention unit.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DAYS","WEEKS","MONTHS","YEARS"]}}

### spec.customPolicy.schedules[].archiveRule

`AwsDlmArchiveRule`

Move aging snapshots to the archive tier instead of deleting
(VOLUME snapshot schedules only).

- rule: configure exactly one of count and interval + interval_unit for the archive tier
- rule: interval and interval_unit are set together

### spec.customPolicy.schedules[].archiveRule.count

`int64`

Keep the newest N archived snapshots.

- rule: {"int64":{"gte":"0"}}

### spec.customPolicy.schedules[].archiveRule.interval

`int64`

Keep archived snapshots for this long, paired with interval_unit.

- rule: {"int64":{"gte":"0"}}

### spec.customPolicy.schedules[].archiveRule.intervalUnit

`string`

The archive retention unit.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DAYS","WEEKS","MONTHS","YEARS"]}}

### spec.customPolicy.schedules[].crossRegionCopyRules

`[]AwsDlmCrossRegionCopyRule`

Replicate created snapshots/AMIs to other regions (up to 3).

- rule: {"repeated":{"maxItems":"3"}}
- rule: cmk_arn requires encrypted=true

### spec.customPolicy.schedules[].crossRegionCopyRules[].targetRegion

`string` · required

The destination region. Example: "us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.customPolicy.schedules[].crossRegionCopyRules[].encrypted

`bool`

Encrypt the copies. Copies of encrypted snapshots are always
encrypted; this forces encryption for unencrypted sources.

### spec.customPolicy.schedules[].crossRegionCopyRules[].cmkArn

`string | valueFrom`

The KMS key in the DESTINATION region. Unset with encrypted=true
means the destination's AWS-managed aws/ebs key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.customPolicy.schedules[].crossRegionCopyRules[].copyTags

`bool`

Copy the snapshot's tags to the copy.

### spec.customPolicy.schedules[].crossRegionCopyRules[].retainRule

`AwsDlmCopyRetainRule`

How long copies are kept in the destination region. Unset keeps
them forever.

### spec.customPolicy.schedules[].crossRegionCopyRules[].retainRule.interval

`int64`

The duration, paired with interval_unit.

- rule: {"int64":{"gte":"1"}}

### spec.customPolicy.schedules[].crossRegionCopyRules[].retainRule.intervalUnit

`string`

The duration unit.

- rule: {"string":{"in":["DAYS","WEEKS","MONTHS","YEARS"]}}

### spec.customPolicy.schedules[].crossRegionCopyRules[].deprecateRule

`AwsDlmCopyRetainRule`

Deprecate AMI copies after a duration (IMAGE_MANAGEMENT).

### spec.customPolicy.schedules[].crossRegionCopyRules[].deprecateRule.interval

`int64`

The duration, paired with interval_unit.

- rule: {"int64":{"gte":"1"}}

### spec.customPolicy.schedules[].crossRegionCopyRules[].deprecateRule.intervalUnit

`string`

The duration unit.

- rule: {"string":{"in":["DAYS","WEEKS","MONTHS","YEARS"]}}

### spec.customPolicy.schedules[].deprecateRule

`AwsDlmDeprecateRule`

Deprecate aging AMIs (IMAGE_MANAGEMENT) so launches stop picking
them while retention still holds them.

- rule: configure exactly one of count and interval + interval_unit
- rule: interval and interval_unit are set together

### spec.customPolicy.schedules[].deprecateRule.count

`int64`

Deprecate all but the newest N AMIs (at most the schedule's
retain count).

- rule: {"int64":{"lte":"1000","gte":"0"}}

### spec.customPolicy.schedules[].deprecateRule.interval

`int64`

Deprecate AMIs older than this, paired with interval_unit (at
most the schedule's retention).

- rule: {"int64":{"gte":"0"}}

### spec.customPolicy.schedules[].deprecateRule.intervalUnit

`string`

The deprecation unit.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DAYS","WEEKS","MONTHS","YEARS"]}}

### spec.customPolicy.schedules[].fastRestoreRule

`AwsDlmFastRestoreRule`

Enable fast snapshot restore on created snapshots in these zones.
Billed per zone-hour per snapshot while enabled.

- rule: configure at most one of count and interval + interval_unit (how long FSR stays enabled); leaving both unset keeps FSR until the snapshot is deleted
- rule: interval and interval_unit are set together

### spec.customPolicy.schedules[].fastRestoreRule.availabilityZones

`[]string` · required

The availability zones (1-10) FSR is enabled in.

- rule: {"repeated":{"minItems":"1","maxItems":"10","unique":true}}

### spec.customPolicy.schedules[].fastRestoreRule.count

`int64`

Keep FSR on the newest N snapshots only.

- rule: {"int64":{"lte":"1000","gte":"0"}}

### spec.customPolicy.schedules[].fastRestoreRule.interval

`int64`

Keep FSR enabled for this long per snapshot, paired with
interval_unit.

- rule: {"int64":{"gte":"0"}}

### spec.customPolicy.schedules[].fastRestoreRule.intervalUnit

`string`

The FSR duration unit.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DAYS","WEEKS","MONTHS","YEARS"]}}

### spec.customPolicy.schedules[].shareRule

`AwsDlmShareRule`

Share created snapshots with other accounts.

- rule: unshare_interval and unshare_interval_unit are set together

### spec.customPolicy.schedules[].shareRule.targetAccounts

`[]string` · required

The accounts to share with.

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"pattern":"^[0-9]{12}$"}}}}

### spec.customPolicy.schedules[].shareRule.unshareInterval

`int64`

Unshare after this long, paired with unshare_interval_unit.
Unset shares until the snapshot is deleted.

- rule: {"int64":{"gte":"0"}}

### spec.customPolicy.schedules[].shareRule.unshareIntervalUnit

`string`

The unshare duration unit.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DAYS","WEEKS","MONTHS","YEARS"]}}

### spec.customPolicy.eventSource

`AwsDlmEventSource`

EVENT_BASED_POLICY: the event to react to.

### spec.customPolicy.eventSource.eventType

`string`

The only published event: another account shares a snapshot with
this one.

- rule: {"string":{"in":["shareSnapshot"]}}

### spec.customPolicy.eventSource.descriptionRegex

`string` · required

React only to snapshots whose description matches this regular
expression. AWS requires it - "^.*$" reacts to everything.

- rule: {"string":{"minLen":"1","maxLen":"1000"}}

### spec.customPolicy.eventSource.snapshotOwners

`[]string` · required

React only to snapshots shared BY these accounts.

- rule: {"repeated":{"minItems":"1","maxItems":"50","unique":true,"items":{"string":{"pattern":"^[0-9]{12}$"}}}}

### spec.customPolicy.action

`AwsDlmAction`

EVENT_BASED_POLICY: the action to take.

### spec.customPolicy.action.name

`string` · required

The action's console label.

- rule: {"string":{"minLen":"1","maxLen":"120"}}

### spec.customPolicy.action.crossRegionCopies

`[]AwsDlmActionCrossRegionCopy` · required

Where to copy the shared snapshot (1-3 destinations).

- rule: {"repeated":{"minItems":"1","maxItems":"3"}}
- rule: cmk_arn requires encrypted=true

### spec.customPolicy.action.crossRegionCopies[].target

`string` · required

The destination region or Outpost ARN. Example: "us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.customPolicy.action.crossRegionCopies[].encrypted

`bool`

Encrypt the copies.

### spec.customPolicy.action.crossRegionCopies[].cmkArn

`string | valueFrom`

The KMS key in the destination. Unset with encrypted=true means
the destination's AWS-managed aws/ebs key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.customPolicy.action.crossRegionCopies[].retainRule

`AwsDlmCopyRetainRule`

How long copies are kept. Unset keeps them forever.

### spec.customPolicy.action.crossRegionCopies[].retainRule.interval

`int64`

The duration, paired with interval_unit.

- rule: {"int64":{"gte":"1"}}

### spec.customPolicy.action.crossRegionCopies[].retainRule.intervalUnit

`string`

The duration unit.

- rule: {"string":{"in":["DAYS","WEEKS","MONTHS","YEARS"]}}

## Validation Rules

- `spec.exactly_one_mode`: configure exactly one of default_policy (AWS's simplified account-wide posture) and custom_policy (named schedules or an event-driven policy)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsDlmLifecyclePolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_id` | `string` | The policy's id (policy-...) - the provider's import ID. |
| `status.outputs.policy_arn` | `string` | The policy's ARN. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.customPolicy.schedules[].crossRegionCopyRules[].cmkArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.customPolicy.action.crossRegionCopies[].cmkArn` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
