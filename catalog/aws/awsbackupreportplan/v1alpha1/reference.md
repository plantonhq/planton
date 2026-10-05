# AwsBackupReportPlan

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBackupReportPlanSpec defines a Backup Audit Manager report plan:
a scheduled report (daily, AWS-managed cadence) of backup jobs,
copy jobs, restore jobs, or control compliance, delivered as
CSV/JSON files to an S3 bucket.

AWS report plan names forbid hyphens (letter first, then letters,
digits, and underscores), which is stricter than metadata.name
conventions - so the name is an explicit field rather than
metadata.name. Changing it forces replacement.

## Example

```yaml
# Canonical AwsBackupReportPlan example (hack/dev manifest and refgen
# Example source): a daily backup-job report into S3. Literal names
# stand in for composed references so the offline `tofu plan` renders
# every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupReportPlan
metadata:
  name: daily-job-report
  id: daily-job-report
  org: test-org
  env: dev
spec:
  region: us-west-2
  reportPlanName: daily_backup_jobs
  description: Daily backup job outcomes delivered to S3
  deliveryChannel:
    s3BucketName:
      value: backup-reports-bucket
    s3KeyPrefix: backup-reports
    formats: ["CSV", "JSON"]
  reportSetting:
    reportTemplate: BACKUP_JOB_REPORT
    regions: ["us-west-2"]
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.reportPlanName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.deliveryChannel` | `AwsBackupReportPlanDeliveryChannel` | yes |  |  |
| `spec.deliveryChannel.s3BucketName` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.deliveryChannel.s3KeyPrefix` | `string` |  |  |  |
| `spec.deliveryChannel.formats` | `[]string` |  |  |  |
| `spec.reportSetting` | `AwsBackupReportPlanReportSetting` | yes |  |  |
| `spec.reportSetting.reportTemplate` | `string` |  |  |  |
| `spec.reportSetting.frameworkArns` | `[]string \| valueFrom` |  |  | AwsBackupFramework (`status.outputs.framework_arn`) |
| `spec.reportSetting.numberOfFrameworks` | `int32` |  |  |  |
| `spec.reportSetting.accounts` | `[]string` |  |  |  |
| `spec.reportSetting.organizationUnits` | `[]string` |  |  |  |
| `spec.reportSetting.regions` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the report plan lives in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.reportPlanName

`string` · required

The report plan's name in AWS: a letter followed by up to 255
letters, digits, or underscores (NO hyphens). Changing it forces
replacement.

- rule: {"string":{"minLen":"1","maxLen":"256","pattern":"^[A-Za-z][0-9A-Za-z_]{0,255}$"}}

### spec.description

`string`

What the report covers, shown in the Backup console.

- rule: {"string":{"maxLen":"1024"}}

### spec.deliveryChannel

`AwsBackupReportPlanDeliveryChannel` · required

Where the report files land.

- rule: {"required":true}

### spec.deliveryChannel.s3BucketName

`string | valueFrom` · required

The destination S3 bucket (by name).

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.deliveryChannel.s3KeyPrefix

`string`

Key prefix inside the bucket. Unset = the bucket root.

### spec.deliveryChannel.formats

`[]string`

Report file formats. Unset = CSV (the AWS default).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["CSV","JSON"]}}}}

### spec.reportSetting

`AwsBackupReportPlanReportSetting` · required

What the report contains.

- rule: {"required":true}

### spec.reportSetting.reportTemplate

`string`

The report template. CHANGING THE TEMPLATE REPLACES THE WHOLE
REPORT PLAN (a ForceNew field nested inside the provider's
report_setting block - easy to miss). The two *_COMPLIANCE
templates report on Audit Manager frameworks and need
framework_arns; the three job templates report on backup, copy,
and restore jobs.

- rule: {"string":{"in":["BACKUP_JOB_REPORT","CONTROL_COMPLIANCE_REPORT","COPY_JOB_REPORT","RESOURCE_COMPLIANCE_REPORT","RESTORE_JOB_REPORT"]}}

### spec.reportSetting.frameworkArns

`[]string | valueFrom`

The Audit Manager frameworks the compliance templates report on.
A report plan may reference frameworks from many components -
this is a many-to-many edge, which is why report plans are their
own kind rather than a framework field.

- references: AwsBackupFramework (`status.outputs.framework_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBackupFramework, name: <that resource's name>, fieldPath: status.outputs.framework_arn}} -- a bare string does not parse

### spec.reportSetting.numberOfFrameworks

`int32`

The number of frameworks the report covers. The provider sends
this only when positive and AWS computes it otherwise - set it
only when pinning the framework count deliberately.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.reportSetting.accounts

`[]string`

Restrict the report to these account IDs ("*" for all accounts
in the organization). Unset = the current account.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.reportSetting.organizationUnits

`[]string`

Restrict the report to these organizational units.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.reportSetting.regions

`[]string`

Restrict the report to these regions ("*" for all regions).
Unset = the report plan's own region.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBackupReportPlan, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.report_plan_arn` | `string` | The report plan's ARN. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.deliveryChannel.s3BucketName` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.reportSetting.frameworkArns` | AwsBackupFramework | `status.outputs.framework_arn` |

## See Also

- [Overview](../README.md)
