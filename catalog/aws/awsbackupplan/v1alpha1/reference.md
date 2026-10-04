# AwsBackupPlan

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBackupPlanSpec defines the desired configuration for an AWS
Backup plan: the scheduled rules that create recovery points, plus
the resource selections that assign resources to the plan.

The plan's name is metadata.name and changing it forces
replacement. AWS identifies the plan by a generated UUID (the
import ID), not the name. Selections fold in as name-keyed entries
- AWS refuses to delete a plan while selections exist, and both
modules create/destroy them with the plan so ordering is never the
author's problem.

## Example

```yaml
# Canonical AwsBackupPlan example (hack/dev manifest and refgen
# Example source): a daily rule with lifecycle and a copy action, plus
# a tag-driven selection. Literal ARNs/names stand in for composed
# references so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupPlan
metadata:
  name: daily-backup-plan
  id: daily-backup-plan
  org: test-org
  env: dev
spec:
  region: us-west-2
  rules:
    - name: daily
      targetVaultName:
        value: app-backup-vault
      schedule: cron(0 5 ? * * *)
      scheduleExpressionTimezone: Etc/UTC
      startWindowMinutes: 60
      completionWindowMinutes: 180
      lifecycle:
        coldStorageAfterDays: 30
        deleteAfterDays: 365
      copyActions:
        - destinationVaultArn:
            value: arn:aws:backup:us-east-1:123456789012:backup-vault:dr-backup-vault
          lifecycle:
            deleteAfterDays: 365
      recoveryPointTags:
        source-plan: daily-backup-plan
  advancedBackupSettings:
    - resourceType: EC2
      backupOptions:
        WindowsVSS: enabled
  selections:
    - name: tagged
      iamRoleArn:
        value: arn:aws:iam::123456789012:role/backup-service-role
      selectionTags:
        - type: STRINGEQUALS
          key: backup
          value: "true"
      conditions:
        stringLike:
          - key: aws:ResourceTag/environment
            value: prod*
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.rules` | `[]AwsBackupPlanRule` | yes |  |  |
| `spec.rules[].name` | `string` | yes |  |  |
| `spec.rules[].targetVaultName` | `string \| valueFrom` | yes |  | AwsBackupVault (`status.outputs.vault_name`) |
| `spec.rules[].schedule` | `string` |  |  |  |
| `spec.rules[].scheduleExpressionTimezone` | `string` |  |  |  |
| `spec.rules[].startWindowMinutes` | `int32` |  |  |  |
| `spec.rules[].completionWindowMinutes` | `int32` |  |  |  |
| `spec.rules[].enableContinuousBackup` | `bool` |  |  |  |
| `spec.rules[].recoveryPointTags` | `map<string, string>` |  |  |  |
| `spec.rules[].targetLogicallyAirGappedBackupVaultArn` | `string \| valueFrom` |  |  | AwsBackupVault (`status.outputs.vault_arn`) |
| `spec.rules[].lifecycle` | `AwsBackupPlanLifecycle` |  |  |  |
| `spec.rules[].lifecycle.coldStorageAfterDays` | `int32` |  |  |  |
| `spec.rules[].lifecycle.deleteAfterDays` | `int32` |  |  |  |
| `spec.rules[].lifecycle.optInToArchiveForSupportedResources` | `bool` |  |  |  |
| `spec.rules[].copyActions` | `[]AwsBackupPlanCopyAction` |  |  |  |
| `spec.rules[].copyActions[].destinationVaultArn` | `string \| valueFrom` | yes |  | AwsBackupVault (`status.outputs.vault_arn`) |
| `spec.rules[].copyActions[].lifecycle` | `AwsBackupPlanLifecycle` |  |  |  |
| `spec.rules[].copyActions[].lifecycle.coldStorageAfterDays` | `int32` |  |  |  |
| `spec.rules[].copyActions[].lifecycle.deleteAfterDays` | `int32` |  |  |  |
| `spec.rules[].copyActions[].lifecycle.optInToArchiveForSupportedResources` | `bool` |  |  |  |
| `spec.rules[].scanActions` | `[]AwsBackupPlanScanAction` |  |  |  |
| `spec.rules[].scanActions[].malwareScanner` | `string` |  |  |  |
| `spec.rules[].scanActions[].scanMode` | `string` |  |  |  |
| `spec.advancedBackupSettings` | `[]AwsBackupPlanAdvancedBackupSetting` |  |  |  |
| `spec.advancedBackupSettings[].resourceType` | `string` |  |  |  |
| `spec.advancedBackupSettings[].backupOptions` | `map<string, string>` | yes |  |  |
| `spec.scanSetting` | `AwsBackupPlanScanSetting` |  |  |  |
| `spec.scanSetting.malwareScanner` | `string` |  |  |  |
| `spec.scanSetting.resourceTypes` | `[]string` | yes |  |  |
| `spec.scanSetting.scannerRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.selections` | `[]AwsBackupPlanSelection` |  |  |  |
| `spec.selections[].name` | `string` | yes |  |  |
| `spec.selections[].iamRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.selections[].resources` | `[]string` |  |  |  |
| `spec.selections[].notResources` | `[]string` |  |  |  |
| `spec.selections[].selectionTags` | `[]AwsBackupPlanSelectionTag` |  |  |  |
| `spec.selections[].selectionTags[].type` | `string` |  |  |  |
| `spec.selections[].selectionTags[].key` | `string` | yes |  |  |
| `spec.selections[].selectionTags[].value` | `string` | yes |  |  |
| `spec.selections[].conditions` | `AwsBackupPlanSelectionConditions` |  |  |  |
| `spec.selections[].conditions.stringEquals` | `[]AwsBackupPlanSelectionConditionPair` |  |  |  |
| `spec.selections[].conditions.stringEquals[].key` | `string` | yes |  |  |
| `spec.selections[].conditions.stringEquals[].value` | `string` | yes |  |  |
| `spec.selections[].conditions.stringNotEquals` | `[]AwsBackupPlanSelectionConditionPair` |  |  |  |
| `spec.selections[].conditions.stringNotEquals[].key` | `string` | yes |  |  |
| `spec.selections[].conditions.stringNotEquals[].value` | `string` | yes |  |  |
| `spec.selections[].conditions.stringLike` | `[]AwsBackupPlanSelectionConditionPair` |  |  |  |
| `spec.selections[].conditions.stringLike[].key` | `string` | yes |  |  |
| `spec.selections[].conditions.stringLike[].value` | `string` | yes |  |  |
| `spec.selections[].conditions.stringNotLike` | `[]AwsBackupPlanSelectionConditionPair` |  |  |  |
| `spec.selections[].conditions.stringNotLike[].key` | `string` | yes |  |  |
| `spec.selections[].conditions.stringNotLike[].value` | `string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the backup plan lives in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.rules

`[]AwsBackupPlanRule` · required

The plan's backup rules: when to back up, where recovery points
land, and how long they live. At least one rule is required (the
provider does not pre-check this; AWS rejects an empty plan).

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].name

`string` · required

Rule name, 1-50 characters of letters, digits, hyphens,
underscores, and periods. The for_each key on both engines.

- rule: {"string":{"minLen":"1","maxLen":"50","pattern":"^[0-9A-Za-z_.-]+$"}}

### spec.rules[].targetVaultName

`string | valueFrom` · required

The standard vault that receives this rule's recovery points -
required on EVERY rule, even when also targeting an air-gapped
vault below.

- references: AwsBackupVault (`status.outputs.vault_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBackupVault, name: <that resource's name>, fieldPath: status.outputs.vault_name}} -- a bare string does not parse

### spec.rules[].schedule

`string`

When the rule fires, as a CloudWatch cron expression (e.g.
"cron(0 5 ? * * *)" for daily 05:00 UTC). Unset = the rule never
fires on a schedule (on-demand only).

### spec.rules[].scheduleExpressionTimezone

`string`

IANA timezone the schedule is evaluated in. Unset = "Etc/UTC"
(the provider default; a state-upgrade backfills it on older
plans).

### spec.rules[].startWindowMinutes

`int32`

Minutes a started backup job waits for resources before
canceling, at least 60. Unset = 60 (the provider default).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":60}}

### spec.rules[].completionWindowMinutes

`int32`

Minutes a running backup job may take before AWS cancels it.
Unset = 180 (the provider default).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.rules[].enableContinuousBackup

`bool`

Continuous (point-in-time-restore) backups instead of snapshots,
for the resource types that support them. AWS caps continuous
retention at 35 days - keep lifecycle.delete_after_days within
it.

### spec.rules[].recoveryPointTags

`map<string, string>`

Tags stamped onto every recovery point this rule creates (NOT
merged with the module's identity tags - these are recovery-point
metadata).

### spec.rules[].targetLogicallyAirGappedBackupVaultArn

`string | valueFrom`

Additionally copy this rule's recovery points into a logically
air-gapped vault (the ransomware-recovery posture). The rule
still needs its standard target_vault_name.

- references: AwsBackupVault (`status.outputs.vault_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBackupVault, name: <that resource's name>, fieldPath: status.outputs.vault_arn}} -- a bare string does not parse

### spec.rules[].lifecycle

`AwsBackupPlanLifecycle`

Recovery-point lifecycle: cold-storage transition and expiry.
Unset = keep forever in warm storage.

- rule: delete_after_days must exceed cold_storage_after_days by at least 90 (AWS's cold-storage minimum)

### spec.rules[].lifecycle.coldStorageAfterDays

`int32` · optional (explicit presence)

Days after creation before the recovery point moves to cold
storage. Unset = never moves. Once in cold storage, it must stay
at least 90 days.

- rule: {"int32":{"gte":1}}

### spec.rules[].lifecycle.deleteAfterDays

`int32` · optional (explicit presence)

Days after creation before the recovery point is deleted. Unset =
kept until deleted manually. When cold storage is configured, AWS
requires this to exceed cold_storage_after_days by at least 90.

- rule: {"int32":{"gte":1}}

### spec.rules[].lifecycle.optInToArchiveForSupportedResources

`bool` · optional (explicit presence)

Move archive-eligible resource types (currently EBS) to archive
tier instead of deleting. Sent to AWS only when true (the
provider never transmits an explicit false).

### spec.rules[].copyActions

`[]AwsBackupPlanCopyAction`

Cross-vault/cross-region/cross-account copies of this rule's
recovery points, each with its own lifecycle.

### spec.rules[].copyActions[].destinationVaultArn

`string | valueFrom` · required

The destination vault's ARN.

- references: AwsBackupVault (`status.outputs.vault_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBackupVault, name: <that resource's name>, fieldPath: status.outputs.vault_arn}} -- a bare string does not parse

### spec.rules[].copyActions[].lifecycle

`AwsBackupPlanLifecycle`

Lifecycle of the COPY (independent of the source recovery
point's). Unset = the copy is kept until deleted manually.

- rule: delete_after_days must exceed cold_storage_after_days by at least 90 (AWS's cold-storage minimum)

### spec.rules[].copyActions[].lifecycle.coldStorageAfterDays

`int32` · optional (explicit presence)

Days after creation before the recovery point moves to cold
storage. Unset = never moves. Once in cold storage, it must stay
at least 90 days.

- rule: {"int32":{"gte":1}}

### spec.rules[].copyActions[].lifecycle.deleteAfterDays

`int32` · optional (explicit presence)

Days after creation before the recovery point is deleted. Unset =
kept until deleted manually. When cold storage is configured, AWS
requires this to exceed cold_storage_after_days by at least 90.

- rule: {"int32":{"gte":1}}

### spec.rules[].copyActions[].lifecycle.optInToArchiveForSupportedResources

`bool` · optional (explicit presence)

Move archive-eligible resource types (currently EBS) to archive
tier instead of deleting. Sent to AWS only when true (the
provider never transmits an explicit false).

### spec.rules[].scanActions

`[]AwsBackupPlanScanAction`

Per-rule malware scans of the recovery points this rule creates.

### spec.rules[].scanActions[].malwareScanner

`string`

The scanner. Only "GUARDDUTY" exists at the pinned provider.

- rule: {"string":{"in":["GUARDDUTY"]}}

### spec.rules[].scanActions[].scanMode

`string`

Full scan of every recovery point, or incremental scan of changed
data only.

- rule: {"string":{"in":["FULL_SCAN","INCREMENTAL_SCAN"]}}

### spec.advancedBackupSettings

`[]AwsBackupPlanAdvancedBackupSetting`

Windows VSS advanced backup settings, one entry per resource type
(the provider accepts only "EC2" at the pinned version).

### spec.advancedBackupSettings[].resourceType

`string`

The resource type. The provider accepts exactly "EC2" at the
pinned version.

- rule: {"string":{"in":["EC2"]}}

### spec.advancedBackupSettings[].backupOptions

`map<string, string>` · required

Backup options for the type - in practice {"WindowsVSS":
"enabled"} (or "disabled").

- rule: {"map":{"minPairs":"1"}}

### spec.scanSetting

`AwsBackupPlanScanSetting`

Plan-wide malware scanning of recovery points (GuardDuty).
Distinct from the per-rule scan_actions: this setting scans
recovery points the plan creates for the listed resource types.

### spec.scanSetting.malwareScanner

`string`

The scanner. Only "GUARDDUTY" exists at the pinned provider.

- rule: {"string":{"in":["GUARDDUTY"]}}

### spec.scanSetting.resourceTypes

`[]string` · required

Resource types whose recovery points are scanned (AWS documents
"EBS", "EC2", "S3", and "ALL").

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"minLen":"1","maxLen":"50","pattern":"^[a-zA-Z0-9\\-\\_\\.]{1,50}$"}}}}

### spec.scanSetting.scannerRoleArn

`string | valueFrom` · required

IAM role GuardDuty assumes to scan the recovery points.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.selections

`[]AwsBackupPlanSelection`

The resource selections assigned to this plan, keyed by name.
Each selection pairs an IAM role (which AWS Backup assumes to
take the backups) with the resources it covers.

### spec.selections[].name

`string` · required

Selection name, 1-50 characters of letters, digits, hyphens,
underscores, and periods. The for_each key on both engines and
half of the composed import ID.

- rule: {"string":{"minLen":"1","maxLen":"50","pattern":"^[0-9A-Za-z_.-]+$"}}

### spec.selections[].iamRoleArn

`string | valueFrom` · required

IAM role AWS Backup assumes to take and manage backups of the
selected resources. The role must trust backup.amazonaws.com
(AWS's managed AWSBackupServiceRolePolicyForBackup policy is the
usual grant).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.selections[].resources

`[]string`

Resource ARNs to include ("*" selects everything the role can
reach).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.selections[].notResources

`[]string`

Resource ARNs to exclude from whatever the other matchers
selected. Once set, AWS keeps the value - it cannot be cleared
back to empty without replacing the selection.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.selections[].selectionTags

`[]AwsBackupPlanSelectionTag`

Tag-match entries: a resource is selected when it carries the
tag. Entries OR together.

### spec.selections[].selectionTags[].type

`string`

The match operator. Only "STRINGEQUALS" exists at the pinned
provider.

- rule: {"string":{"in":["STRINGEQUALS"]}}

### spec.selections[].selectionTags[].key

`string` · required

The tag key to match.

- rule: {"string":{"minLen":"1"}}

### spec.selections[].selectionTags[].value

`string` · required

The tag value to match.

- rule: {"string":{"minLen":"1"}}

### spec.selections[].conditions

`AwsBackupPlanSelectionConditions`

Fine-grained tag conditions (AND semantics across the four
operator lists). Once set, AWS keeps the value - clearing it
requires replacing the selection.

### spec.selections[].conditions.stringEquals

`[]AwsBackupPlanSelectionConditionPair`

Exact-match conditions.

### spec.selections[].conditions.stringEquals[].key

`string` · required

The condition key (e.g. "aws:ResourceTag/environment").

- rule: {"string":{"minLen":"1"}}

### spec.selections[].conditions.stringEquals[].value

`string` · required

The condition value.

- rule: {"string":{"minLen":"1"}}

### spec.selections[].conditions.stringNotEquals

`[]AwsBackupPlanSelectionConditionPair`

Exact-mismatch conditions.

### spec.selections[].conditions.stringNotEquals[].key

`string` · required

The condition key (e.g. "aws:ResourceTag/environment").

- rule: {"string":{"minLen":"1"}}

### spec.selections[].conditions.stringNotEquals[].value

`string` · required

The condition value.

- rule: {"string":{"minLen":"1"}}

### spec.selections[].conditions.stringLike

`[]AwsBackupPlanSelectionConditionPair`

Wildcard-match conditions.

### spec.selections[].conditions.stringLike[].key

`string` · required

The condition key (e.g. "aws:ResourceTag/environment").

- rule: {"string":{"minLen":"1"}}

### spec.selections[].conditions.stringLike[].value

`string` · required

The condition value.

- rule: {"string":{"minLen":"1"}}

### spec.selections[].conditions.stringNotLike

`[]AwsBackupPlanSelectionConditionPair`

Wildcard-mismatch conditions.

### spec.selections[].conditions.stringNotLike[].key

`string` · required

The condition key (e.g. "aws:ResourceTag/environment").

- rule: {"string":{"minLen":"1"}}

### spec.selections[].conditions.stringNotLike[].value

`string` · required

The condition value.

- rule: {"string":{"minLen":"1"}}

## Validation Rules

- `spec.rule_names_unique`: rules entries must have unique names
- `spec.selection_names_unique`: selections entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBackupPlan, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.plan_id` | `string` | The plan's AWS-generated ID (a UUID - also the provider's import ID; the plan's name is NOT its identity at AWS). |
| `status.outputs.plan_arn` | `string` | The plan's ARN. |
| `status.outputs.plan_version` | `string` | The plan's version ID (changes on every plan update). |
| `status.outputs.selection_ids` | `map<string, string>` | AWS-generated selection IDs keyed by selection name (each selection imports as "{plan_id}\|{selection_id}"). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.rules[].targetVaultName` | AwsBackupVault | `status.outputs.vault_name` |
| `spec.rules[].targetLogicallyAirGappedBackupVaultArn` | AwsBackupVault | `status.outputs.vault_arn` |
| `spec.rules[].copyActions[].destinationVaultArn` | AwsBackupVault | `status.outputs.vault_arn` |
| `spec.scanSetting.scannerRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.selections[].iamRoleArn` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)
