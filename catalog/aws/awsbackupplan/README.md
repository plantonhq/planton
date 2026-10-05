<p align="center">
  <img src="logo.svg" alt="AWS Backup Plan" width="80"/>
</p>

# AWS Backup Plan

Manage an [AWS Backup plan](https://docs.aws.amazon.com/aws-backup/latest/devguide/about-backup-plans.html)
— the scheduled rules that create recovery points, plus the resource
selections that assign resources to the plan.

## What Gets Managed

- **The plan** (`metadata.name` is the plan name; AWS identifies it by
  a generated UUID): one or more **rules** with schedules (cron +
  timezone), start/completion windows, continuous (point-in-time)
  backups, recovery-point tags, lifecycle (cold storage + expiry,
  honoring AWS's 90-day cold-storage minimum), cross-vault/region
  **copy actions**, per-rule GuardDuty **scan actions**, and
  air-gapped-vault targeting.
- **Windows VSS** advanced settings (EC2) and the plan-wide **malware
  scan setting**.
- **Selections** folded as name-keyed entries: the IAM role AWS Backup
  assumes plus resource coverage by ARN, tag matchers, and
  fine-grained conditions. AWS-generated selection IDs land in the
  `selection_ids` output map.

The vault the plan targets is deliberately NOT part of this kind
— see [AwsBackupVault](../awsbackupvault).

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
