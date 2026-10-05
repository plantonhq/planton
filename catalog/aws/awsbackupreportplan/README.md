<p align="center">
  <img src="logo.svg" alt="AWS Backup Report Plan" width="80"/>
</p>

# AWS Backup Report Plan

Manage a [Backup Audit Manager report plan](https://docs.aws.amazon.com/aws-backup/latest/devguide/create-report-plan.html)
— scheduled reports of backup jobs, copy jobs, restore jobs, or
control compliance, delivered as CSV/JSON files to an S3 bucket.

## What Gets Managed

- **The report plan** (`spec.report_plan_name` is the AWS name —
  report plan names forbid hyphens, so `metadata.name` stays
  Planton-side).
- **The delivery channel**: destination bucket, key prefix, and
  formats (CSV/JSON).
- **The report setting**: the template (job reports or compliance
  reports — changing the template REPLACES the whole report plan),
  the frameworks compliance templates report on (a many-to-many
  reference), and account/OU/region coverage.

The frameworks the compliance templates evaluate are deliberately NOT
part of this kind — see
[AwsBackupFramework](../awsbackupframework).

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
