# AwsBackupRestoreTestingPlan

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBackupRestoreTestingPlanSpec defines an AWS Backup restore
testing plan with its folded selections: scheduled, automated
restore tests that prove recovery points actually restore (and feed
the restore-time metrics Backup Audit Manager can report on). AWS
runs each test into a temporary copy and deletes it after the
validation window - tests bill as regular restores.

AWS restore testing names forbid hyphens and periods (letters,
digits, and underscores only), which is stricter than metadata.name
conventions - so the name is an explicit field rather than
metadata.name. Changing it forces replacement.

## Example

```yaml
# Canonical AwsBackupRestoreTestingPlan example (hack/dev manifest and
# refgen Example source): weekly random drills over every vault with
# one EBS selection. Literal ARNs stand in for composed references so
# the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupRestoreTestingPlan
metadata:
  name: weekly-restore-drills
  id: weekly-restore-drills
  org: test-org
  env: dev
spec:
  region: us-west-2
  planName: weekly_restore_drills
  scheduleExpression: cron(0 5 ? * MON *)
  scheduleExpressionTimezone: Etc/UTC
  startWindowHours: 8
  recoveryPointSelection:
    algorithm: RANDOM_WITHIN_WINDOW
    includeVaults: ["*"]
    recoveryPointTypes: ["SNAPSHOT"]
    selectionWindowDays: 30
  selections:
    - name: ebs_volumes
      protectedResourceType: EBS
      iamRoleArn:
        value: arn:aws:iam::123456789012:role/restore-testing-role
      protectedResourceArns: ["*"]
      validationWindowHours: 4
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.planName` | `string` | yes |  |  |
| `spec.scheduleExpression` | `string` | yes |  |  |
| `spec.scheduleExpressionTimezone` | `string` |  |  |  |
| `spec.startWindowHours` | `int32` |  |  |  |
| `spec.recoveryPointSelection` | `AwsBackupRestoreTestingPlanRecoveryPointSelection` | yes |  |  |
| `spec.recoveryPointSelection.algorithm` | `string` |  |  |  |
| `spec.recoveryPointSelection.includeVaults` | `[]string` | yes |  |  |
| `spec.recoveryPointSelection.recoveryPointTypes` | `[]string` | yes |  |  |
| `spec.recoveryPointSelection.excludeVaults` | `[]string` |  |  |  |
| `spec.recoveryPointSelection.selectionWindowDays` | `int32` |  |  |  |
| `spec.selections` | `[]AwsBackupRestoreTestingPlanSelection` |  |  |  |
| `spec.selections[].name` | `string` | yes |  |  |
| `spec.selections[].protectedResourceType` | `string` | yes |  |  |
| `spec.selections[].iamRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.selections[].protectedResourceArns` | `[]string` |  |  |  |
| `spec.selections[].protectedResourceConditions` | `AwsBackupRestoreTestingPlanSelectionConditions` |  |  |  |
| `spec.selections[].protectedResourceConditions.stringEquals` | `[]AwsBackupRestoreTestingPlanSelectionConditionPair` |  |  |  |
| `spec.selections[].protectedResourceConditions.stringEquals[].key` | `string` | yes |  |  |
| `spec.selections[].protectedResourceConditions.stringEquals[].value` | `string` | yes |  |  |
| `spec.selections[].protectedResourceConditions.stringNotEquals` | `[]AwsBackupRestoreTestingPlanSelectionConditionPair` |  |  |  |
| `spec.selections[].protectedResourceConditions.stringNotEquals[].key` | `string` | yes |  |  |
| `spec.selections[].protectedResourceConditions.stringNotEquals[].value` | `string` | yes |  |  |
| `spec.selections[].restoreMetadataOverrides` | `map<string, string>` |  |  |  |
| `spec.selections[].validationWindowHours` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the restore testing plan lives in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.planName

`string` · required

The plan's name in AWS: 1-50 letters, digits, or underscores (NO
hyphens or periods). Changing it forces replacement.

- rule: {"string":{"minLen":"1","maxLen":"50","pattern":"^[0-9A-Za-z_]+$"}}

### spec.scheduleExpression

`string` · required

When the tests run, as a CloudWatch cron expression (e.g.
"cron(0 5 ? * MON *)" for Mondays 05:00).

- rule: {"string":{"minLen":"1"}}

### spec.scheduleExpressionTimezone

`string`

IANA timezone the schedule is evaluated in. Unset = UTC. Once
set, AWS keeps a value - it cannot be cleared back to unset.

### spec.startWindowHours

`int32`

Hours the test waits for a scheduled run to start before skipping
it, 1-168. Unset = AWS's default. Once set, AWS keeps a value.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":168,"gte":1}}

### spec.recoveryPointSelection

`AwsBackupRestoreTestingPlanRecoveryPointSelection` · required

Which recovery points each test restores.

- rule: {"required":true}
- rule: include_vaults entries must be vault ARNs or the literal "*"
- rule: exclude_vaults entries must be vault ARNs or the literal "*"

### spec.recoveryPointSelection.algorithm

`string`

How to pick within the selection window: the newest recovery
point, or a random one (random exercises older points too - the
stronger proof).

- rule: {"string":{"in":["LATEST_WITHIN_WINDOW","RANDOM_WITHIN_WINDOW"]}}

### spec.recoveryPointSelection.includeVaults

`[]string` · required

Vaults to draw recovery points from: vault ARNs, or the literal
"*" for every vault.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.recoveryPointSelection.recoveryPointTypes

`[]string` · required

Which recovery point kinds are eligible.

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"in":["CONTINUOUS","SNAPSHOT"]}}}}

### spec.recoveryPointSelection.excludeVaults

`[]string`

Vaults to exclude (ARNs or "*"). Once set, AWS keeps a value - it
cannot be cleared back to unset.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.recoveryPointSelection.selectionWindowDays

`int32`

How many days back the test looks for eligible recovery points,
1-365. Unset = 30 (the AWS default). Once set, AWS keeps a value.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":365,"gte":1}}

### spec.selections

`[]AwsBackupRestoreTestingPlanSelection`

The per-resource-type selections tested under this plan, keyed by
name. A plan without selections schedules tests that never select
anything - add at least one to test for real.

- rule: set exactly one of protected_resource_arns / protected_resource_conditions

### spec.selections[].name

`string` · required

Selection name, 1-50 letters, digits, or underscores (NO hyphens
or periods). The for_each key on both engines and half of the
composed import ID.

- rule: {"string":{"minLen":"1","maxLen":"50","pattern":"^[0-9A-Za-z_]+$"}}

### spec.selections[].protectedResourceType

`string` · required

The resource type this selection tests (e.g. "EBS", "EC2",
"RDS", "S3" - AWS's restore-testing type vocabulary).

- rule: {"string":{"minLen":"1"}}

### spec.selections[].iamRoleArn

`string | valueFrom` · required

IAM role AWS Backup assumes to run the restore tests. The role
must trust backup.amazonaws.com with restore permissions for the
tested type.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.selections[].protectedResourceArns

`[]string`

Explicit resource ARNs to test ("*" for everything of the type).
Exactly one of this and protected_resource_conditions.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.selections[].protectedResourceConditions

`AwsBackupRestoreTestingPlanSelectionConditions`

Tag conditions selecting the resources to test. Exactly one of
this and protected_resource_arns. AWS returns empty condition
lists as present-but-empty; both modules collapse that to absent
so an omitted block stays omitted.

- rule: set at least one of string_equals / string_not_equals

### spec.selections[].protectedResourceConditions.stringEquals

`[]AwsBackupRestoreTestingPlanSelectionConditionPair`

Exact-match conditions.

### spec.selections[].protectedResourceConditions.stringEquals[].key

`string` · required

The condition key (e.g. "aws:ResourceTag/backup").

- rule: {"string":{"minLen":"1"}}

### spec.selections[].protectedResourceConditions.stringEquals[].value

`string` · required

The condition value.

- rule: {"string":{"minLen":"1"}}

### spec.selections[].protectedResourceConditions.stringNotEquals

`[]AwsBackupRestoreTestingPlanSelectionConditionPair`

Exact-mismatch conditions.

### spec.selections[].protectedResourceConditions.stringNotEquals[].key

`string` · required

The condition key (e.g. "aws:ResourceTag/backup").

- rule: {"string":{"minLen":"1"}}

### spec.selections[].protectedResourceConditions.stringNotEquals[].value

`string` · required

The condition value.

- rule: {"string":{"minLen":"1"}}

### spec.selections[].restoreMetadataOverrides

`map<string, string>`

Restore metadata overrides for the test restore (e.g. an
alternate subnet or instance type), keyed by metadata key. AWS
lowercases the keys on read.

### spec.selections[].validationWindowHours

`int32`

Hours the restored copy stays up for validation before AWS
deletes it, 1-168. Unset = AWS's default. Once set, AWS keeps a
value.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":168,"gte":1}}

## Validation Rules

- `spec.selection_names_unique`: selections entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBackupRestoreTestingPlan, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.restore_testing_plan_arn` | `string` | The restore testing plan's ARN. (The plan and its selections import by name - the plan's name is its identity; AWS assigns no separate ID.) |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.selections[].iamRoleArn` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)
