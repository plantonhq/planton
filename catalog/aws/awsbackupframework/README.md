<p align="center">
  <img src="logo.svg" alt="AWS Backup Framework" width="80"/>
</p>

# AWS Backup Framework

Manage a [Backup Audit Manager framework](https://docs.aws.amazon.com/aws-backup/latest/devguide/audit-frameworks.html)
— a set of controls that continuously evaluate the account's backup
posture: are resources covered by a plan, are retention minimums met,
are recovery points encrypted.

## What Gets Managed

- **The framework** (`spec.framework_name` is the AWS name — framework
  names forbid hyphens, so `metadata.name` stays Planton-side).
- **Controls** from AWS's Backup Audit Manager vocabulary (e.g.
  `BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN`,
  `BACKUP_RECOVERY_POINT_MINIMUM_RETENTION_CHECK`), each with its
  input parameters and an optional scope (resource IDs, types, or one
  tag).

Framework evaluations run on AWS Config: the region needs an ACTIVE
Config recorder recording the backup resource types, or the
framework's deployment lands `FAILED` (visible in the console's
deployment status, not as an apply error).

The compliance REPORT the framework feeds is deliberately NOT part of
this kind — see [AwsBackupReportPlan](../awsbackupreportplan).

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
