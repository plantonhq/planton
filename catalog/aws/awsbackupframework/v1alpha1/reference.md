# AwsBackupFramework

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBackupFrameworkSpec defines a Backup Audit Manager framework: a
set of controls that continuously evaluate the account's backup
posture (are resources covered by a plan, are retention minimums
met, are recovery points encrypted, ...).

Framework evaluations run on AWS Config: the region needs an
ACTIVE Config recorder recording the backup resource types, or the
framework's deployment lands FAILED (the provider treats a FAILED
deployment as a completed apply - the failure shows in
deployment_status, not as an error).

AWS framework names forbid hyphens (letter first, then letters,
digits, and underscores), which is stricter than metadata.name
conventions - so the name is an explicit field rather than
metadata.name. Changing it forces replacement.

## Example

```yaml
# Canonical AwsBackupFramework example (hack/dev manifest and refgen
# Example source): the coverage-and-retention audit pair.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupFramework
metadata:
  name: backup-coverage-audit
  id: backup-coverage-audit
  org: test-org
  env: dev
spec:
  region: us-west-2
  frameworkName: backup_coverage_audit
  description: Are resources protected by a plan and retained long enough
  controls:
    - name: BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN
    - name: BACKUP_RECOVERY_POINT_MINIMUM_RETENTION_CHECK
      inputParameters:
        - name: requiredRetentionDays
          value: "35"
      scope:
        complianceResourceTypes:
          - EBS
        tags:
          environment: prod
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.frameworkName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.controls` | `[]AwsBackupFrameworkControl` | yes |  |  |
| `spec.controls[].name` | `string` | yes |  |  |
| `spec.controls[].inputParameters` | `[]AwsBackupFrameworkControlInputParameter` |  |  |  |
| `spec.controls[].inputParameters[].name` | `string` | yes |  |  |
| `spec.controls[].inputParameters[].value` | `string` | yes |  |  |
| `spec.controls[].scope` | `AwsBackupFrameworkControlScope` |  |  |  |
| `spec.controls[].scope.complianceResourceIds` | `[]string` |  |  |  |
| `spec.controls[].scope.complianceResourceTypes` | `[]string` |  |  |  |
| `spec.controls[].scope.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the framework lives in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.frameworkName

`string` · required

The framework's name in AWS: a letter followed by up to 255
letters, digits, or underscores (NO hyphens). Changing it forces
replacement.

- rule: {"string":{"minLen":"1","maxLen":"256","pattern":"^[A-Za-z][0-9A-Za-z_]{0,255}$"}}

### spec.description

`string`

What the framework audits, shown in the Backup console.

- rule: {"string":{"maxLen":"1024"}}

### spec.controls

`[]AwsBackupFrameworkControl` · required

The controls the framework evaluates. Control names come from
AWS's Backup Audit Manager control vocabulary (e.g.
BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN,
BACKUP_RECOVERY_POINT_MINIMUM_RETENTION_CHECK).

- rule: {"repeated":{"minItems":"1"}}

### spec.controls[].name

`string` · required

The AWS control name (Backup Audit Manager's vocabulary).

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.controls[].inputParameters

`[]AwsBackupFrameworkControlInputParameter`

The control's parameters (e.g. requiredRetentionDays for the
minimum-retention check). Parameterless controls omit this.

### spec.controls[].inputParameters[].name

`string` · required

The parameter's name (defined by the control).

- rule: {"string":{"minLen":"1"}}

### spec.controls[].inputParameters[].value

`string` · required

The parameter's value.

- rule: {"string":{"minLen":"1"}}

### spec.controls[].scope

`AwsBackupFrameworkControlScope`

Which resources the control evaluates. PREFER DECLARING THIS
EXPLICITLY on resource-scoped controls (e.g.
BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN): omitting it is valid
at AWS - the control then evaluates everything it applies to -
but AWS materializes its default all-supported-types scope
server-side, and the pinned Terraform provider ships no diff
suppression for that echo, so a scope-less control makes every
later plan propose stripping a scope AWS will re-materialize
(live-verified 2026-08-25; the provider's own acceptance tests
only exercise the scoped form). Parameter-only checks (e.g. the
minimum-retention check) take no scope and are unaffected.

### spec.controls[].scope.complianceResourceIds

`[]string`

Specific resource IDs to evaluate (at most 100).

- rule: {"repeated":{"maxItems":"100","items":{"string":{"minLen":"1"}}}}

### spec.controls[].scope.complianceResourceTypes

`[]string`

Resource types to evaluate (e.g. "EBS", "RDS").

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.controls[].scope.tags

`map<string, string>`

Scope by tag - AWS accepts AT MOST ONE key/value pair here (the
provider documents the single-pair limit).

- rule: {"map":{"maxPairs":"1"}}

## Validation Rules

- `spec.control_names_unique`: controls entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBackupFramework, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.framework_arn` | `string` | The framework's ARN - what report plans reference. |
| `status.outputs.region` | `string` | The AWS region the framework lives in. Frameworks are region-scoped and addressed by region + name, so any consumer (or verifier) reaching the framework off the ambient region needs this alongside the ARN. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBackupReportPlan | `spec.reportSetting.frameworkArns` | `status.outputs.framework_arn` |

## See Also

- [Overview](../README.md)
