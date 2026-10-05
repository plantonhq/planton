# AWS Backup Report Plan

Deploys a Backup Audit Manager report plan: a scheduled report of backup jobs, copy jobs, restore jobs, or control compliance, delivered daily as CSV/JSON files to an S3 bucket. The two compliance templates report over Audit Manager frameworks (a many-to-many edge — one report plan can cover frameworks from any number of components); the three job templates need no framework at all. Reports run on AWS's own daily cadence — there is no cron to tune.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Backup Report Plan** — the report plan carrying its template, framework references, account/OU/region coverage, and the S3 delivery channel (bucket, key prefix, formats)

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AWS Backup permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- A destination S3 bucket whose policy grants the AWS Backup report service `s3:PutObject` (AWS documents the exact statement in the Backup Audit Manager guide). A missing grant fails report JOBS, not the deploy — the report plan creates fine and then every daily cycle fails, so check `aws backup list-report-jobs` after the first cycle.
- (Only for the compliance templates) deployed Backup Audit Manager frameworks in the same region, referenced via `frameworkArns`.

## Deploy

### Console

Open the deployment store, find **AWS Backup Report Plan**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the template choice and delivery channel. Start from the **Daily Job Report** preset in the [Presets](#presets) tab for job outcomes, or the **Compliance Evidence** preset for framework compliance reports.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupReportPlan
metadata:
  name: daily-job-report
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  reportPlanName: daily_backup_jobs
  description: Daily backup job outcomes delivered to S3
  deliveryChannel:
    s3BucketName:
      valueFrom:
        kind: AwsS3Bucket
        name: backup-reports
        fieldPath: status.outputs.bucket_id
    s3KeyPrefix: backup-reports
    formats: ["CSV", "JSON"]
  reportSetting:
    reportTemplate: BACKUP_JOB_REPORT
```

```shell
planton apply -f backup-report-plan.yaml
```

This creates a report plan named `daily_backup_jobs` that delivers daily backup-job outcome reports in both formats to the referenced bucket under the `backup-reports/` prefix. An Infra Job tracks the provisioning in real time.

### InfraChart

When the report plan deploys alongside its framework and bucket in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  reportPlanName: control_compliance_evidence
  description: Control compliance over the audit frameworks
  deliveryChannel:
    s3BucketName:
      valueFrom:
        kind: AwsS3Bucket
        name: backup-reports
        fieldPath: status.outputs.bucket_id
  reportSetting:
    reportTemplate: CONTROL_COMPLIANCE_REPORT
    frameworkArns:
      - valueFrom:
          kind: AwsBackupFramework
          name: backup-posture
          fieldPath: status.outputs.framework_arn
```

The InfraPipeline resolves the dependency graph, creates the bucket and framework first, then provisions the report plan against them.

## Key Configuration

These are the most important decisions when configuring a report plan. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The template is a hidden one-way door** — Changing `reportTemplate` replaces the whole report plan: the provider marks it ForceNew from inside the nested block, which is easy to miss. Choose deliberately between the three job templates (`BACKUP_JOB_REPORT`, `COPY_JOB_REPORT`, `RESTORE_JOB_REPORT`) and the two compliance templates (`CONTROL_COMPLIANCE_REPORT`, `RESOURCE_COMPLIANCE_REPORT`) up front; a template swap later means a new plan and a break in the report series.

**reportPlanName is not metadata.name** — AWS forbids hyphens in report plan names (a letter first, then letters, digits, and underscores), so the AWS name is an explicit spec field. Changing it forces replacement. It is also what `aws backup start-report-job` takes to run a report on demand.

**The bucket policy is the failure point** — The deploy succeeds regardless of bucket permissions; only report jobs fail. When the first daily report never lands, the bucket policy's missing `s3:PutObject` grant to the report service is the first place to look.

**Compliance templates need same-region frameworks** — `frameworkArns` accepts any number of framework references, but they must live in the report plan's region. Cross-region evidence takes one report plan per region plus the `regions` coverage list ("*" covers all regions). Leave `numberOfFrameworks` unset unless deliberately pinning the count — AWS computes it otherwise.

**Coverage defaults to here and now** — Unset `accounts` means the current account and unset `regions` means the report plan's own region. Organization-wide evidence takes `accounts: ["*"]` or explicit `organizationUnits` — decide whether this plan is a local record or the organization's audit trail.

**Formats serve different readers** — CSV for humans and spreadsheets, JSON for pipelines; both can be delivered in the same run. Unset defaults to CSV only, which quietly starves any automation expecting JSON.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsS3Bucket** | `deliveryChannel.s3BucketName` | `status.outputs.bucket_id` |
| **AwsBackupFramework** | `reportSetting.frameworkArns[]` | `status.outputs.framework_arn` |

### What This Kind Provides

The single output, `report_plan_arn`, is an identity echo rather than a composition input — no catalog kind consumes it via ValueFromRef. It serves IAM policies that scope report administration and AWS CLI/API addressing; the report files themselves land in the delivery bucket, which is where downstream consumers (auditors, pipelines) actually read.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Daily job evidence** — a `BACKUP_JOB_REPORT` plan delivering both formats to a dedicated reports bucket under a clear prefix. The unglamorous record that answers "did last night's backups run" without console archaeology; pair it with `COPY_JOB_REPORT` and `RESTORE_JOB_REPORT` plans when copies and restores need the same paper trail. Start from the **Daily Job Report** preset.

**Framework compliance evidence** — a `CONTROL_COMPLIANCE_REPORT` plan wired to one or more audit frameworks, delivering the scheduled evidence compliance reviews ask for. The framework evaluates continuously; this turns evaluations into dated documents in S3. Start from the **Compliance Evidence** preset.

**One evidence bucket, many prefixes** — point every report plan at the same hardened bucket and separate them by `s3KeyPrefix`. One bucket policy to maintain, one retention story for auditors, and lifecycle rules on the bucket manage evidence aging.

## Works With

- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the delivery destination; its policy must let the report service write
- [**AWS Backup Framework**](/infra-catalog/aws-backup-framework) — the frameworks the compliance templates report over
- [**AWS Backup Plan**](/infra-catalog/aws-backup-plan) — the source of the backup, copy, and restore jobs the job templates report on
