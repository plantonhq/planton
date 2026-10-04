# AwsSsmAssociation

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSsmAssociationSpec defines the desired configuration for one AWS
Systems Manager State Manager association: the binding of an SSM
document to targets on a schedule ("run AWS-RunPatchBaseline on
every instance tagged env=prod every night").

The document reference accepts ANY document name - an AWS-managed
document (AWS-RunShellScript, AmazonCloudWatch-ManageAgent, ...) as
a literal value, or a customer-owned AwsSsmDocument by reference -
which is why the association is its own kind rather than a
document satellite. Changing the document forces replacement; every
other change creates a new association version in place.

AWS identifies the association by a generated UUID (the import ID),
not by name.

## Example

```yaml
# Canonical AwsSsmAssociation example (hack/dev manifest and refgen
# Example source): a nightly patch scan of tagged instances via the
# AWS-managed AWS-RunPatchBaseline document, with S3 output and the
# full compliance/rate surface. Literal values stand in for composed
# references so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSsmAssociation
metadata:
  name: nightly-patch-scan
  id: nightly-patch-scan
  org: test-org
  env: dev
spec:
  region: us-west-2
  documentName:
    value: AWS-RunPatchBaseline
  associationName: nightly-patch-scan
  documentVersion: $DEFAULT
  parameters:
    Operation: Scan
  targets:
    - key: tag:env
      values:
        - prod
  scheduleExpression: cron(0 2 ? * * *)
  applyOnlyAtCronInterval: true
  complianceSeverity: HIGH
  syncCompliance: AUTO
  maxConcurrency: 10%
  maxErrors: "0"
  calendarNames:
    - change-freeze-calendar
  outputLocation:
    s3BucketName:
      value: patch-scan-output-bucket
    s3KeyPrefix: ssm/patch-scans
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.documentName` | `string \| valueFrom` | yes |  | AwsSsmDocument (`status.outputs.document_name`) |
| `spec.associationName` | `string` | yes |  |  |
| `spec.documentVersion` | `string` |  |  |  |
| `spec.parameters` | `map<string, string>` |  |  |  |
| `spec.targets` | `[]AwsSsmAssociationTarget` |  |  |  |
| `spec.targets[].key` | `string` | yes |  |  |
| `spec.targets[].values` | `[]string` | yes |  |  |
| `spec.scheduleExpression` | `string` |  |  |  |
| `spec.applyOnlyAtCronInterval` | `bool` |  |  |  |
| `spec.complianceSeverity` | `string` |  |  |  |
| `spec.syncCompliance` | `string` |  |  |  |
| `spec.maxConcurrency` | `string` |  |  |  |
| `spec.maxErrors` | `string` |  |  |  |
| `spec.automationTargetParameterName` | `string` | yes |  |  |
| `spec.calendarNames` | `[]string` |  |  |  |
| `spec.outputLocation` | `AwsSsmAssociationOutputLocation` |  |  |  |
| `spec.outputLocation.s3BucketName` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.outputLocation.s3KeyPrefix` | `string` |  |  |  |
| `spec.outputLocation.s3Region` | `string` | yes |  |  |
| `spec.waitForSuccessTimeoutSeconds` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the association lives in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.documentName

`string | valueFrom` · required

The SSM document to associate: an AWS-managed document's name as a
literal value, or a customer-owned AwsSsmDocument by reference.
Changing it forces replacement.

- references: AwsSsmDocument (`status.outputs.document_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSsmDocument, name: <that resource's name>, fieldPath: status.outputs.document_name}} -- a bare string does not parse

### spec.associationName

`string` · required

A descriptive name for the association (shown in the State
Manager console). Unset = AWS shows the association unnamed.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"3","maxLen":"128","pattern":"^[0-9A-Za-z_.-]{3,128}$"}}

### spec.documentVersion

`string`

The document version to associate: "$LATEST", "$DEFAULT", or a
concrete version number. Unset = "$DEFAULT". AWS updates the
association when the pinned version's meaning changes.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([$]LATEST|[$]DEFAULT|[1-9][0-9]*)$"}}

### spec.parameters

`map<string, string>`

Values for the document's input parameters (parameter name to
value; list-typed document parameters take comma-separated
values). AWS materializes the document's declared defaults into
this map server-side, so a freshly imported association shows the
merged result - the first plan reconciles it.

### spec.targets

`[]AwsSsmAssociationTarget`

What the association runs on: up to 5 target entries (instance
IDs, tag matches, or resource groups). Unset targets a document
that manages its own scope (e.g. automation runbooks driven by
automation_target_parameter_name).

- rule: {"repeated":{"maxItems":"5"}}

### spec.targets[].key

`string` · required

The target key: "InstanceIds", "tag:<key>", "resource-groups:Name",
or "aws:PrincipalTag/<key>" forms per the SSM targeting grammar.

- rule: {"string":{"minLen":"1","maxLen":"163"}}

### spec.targets[].values

`[]string` · required

The values for the key (instance IDs, the tag value, the resource
group name; "*" with key "InstanceIds" targets everything).

- rule: {"repeated":{"minItems":"1","maxItems":"50","items":{"string":{"minLen":"1"}}}}

### spec.scheduleExpression

`string`

When the association (re)applies, as a cron or rate expression
(e.g. "cron(0 2 ? * SUN *)", "rate(7 days)"). Unset = the
association applies once on create and on every association
change.

- rule: {"string":{"maxLen":"256"}}

### spec.applyOnlyAtCronInterval

`bool`

Run ONLY at the next cron interval instead of immediately on
create/update. Meaningful only with a cron schedule_expression.

### spec.complianceSeverity

`string`

Severity the association's compliance records report when a run
fails. Unset = AWS's UNSPECIFIED.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CRITICAL","HIGH","MEDIUM","LOW","UNSPECIFIED"]}}

### spec.syncCompliance

`string`

How compliance is reported: AUTO (Systems Manager marks compliance
from the association run itself) or MANUAL (an external process
writes compliance via PutComplianceItems). Unset = AUTO.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AUTO","MANUAL"]}}

### spec.maxConcurrency

`string`

Maximum targets the association runs on concurrently: an absolute
count ("10") or a percentage ("10%"). Unset = AWS's default (all
targets at once).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([1-9][0-9]*|[1-9][0-9]%|[1-9]%|100%)$"}}

### spec.maxErrors

`string`

Failures after which AWS stops scheduling new runs for this
association interval: an absolute count ("1", "0") or a
percentage ("10%").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([1-9][0-9]*|[0]|[1-9][0-9]%|[0-9]%|100%)$"}}

### spec.automationTargetParameterName

`string` · required

For Automation documents: the document parameter the target
resource IDs feed into (rate-controlled automation on targets).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"50"}}

### spec.calendarNames

`[]string`

Change Calendar document names (or ARNs) gating the association:
it runs only when every named calendar is open.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.outputLocation

`AwsSsmAssociationOutputLocation`

Where command output lands (S3). Unset = output is not stored.

### spec.outputLocation.s3BucketName

`string | valueFrom` · required

The destination bucket.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.outputLocation.s3KeyPrefix

`string`

Key prefix inside the bucket.

- rule: {"string":{"maxLen":"500"}}

### spec.outputLocation.s3Region

`string` · required

The bucket's region, when it differs from the association's.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"3","maxLen":"20"}}

### spec.waitForSuccessTimeoutSeconds

`int32`

Seconds the CREATE waits for the association's first run to
succeed before failing the deploy. Deploy-side behavior, never
read back from AWS - a freshly imported association shows it
unset. Leave unset when no targets exist yet (the wait would
always time out).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSsmAssociation, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.association_id` | `string` | The association's AWS-generated ID (a UUID - also the provider's import ID; the association's name is NOT its identity at AWS). |
| `status.outputs.association_arn` | `string` | The association's ARN. |
| `status.outputs.document_name` | `string` | The document name the association resolved to (the chart-useful echo of the document reference). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.documentName` | AwsSsmDocument | `status.outputs.document_name` |
| `spec.outputLocation.s3BucketName` | AwsS3Bucket | `status.outputs.bucket_id` |

## See Also

- [Overview](../README.md)
