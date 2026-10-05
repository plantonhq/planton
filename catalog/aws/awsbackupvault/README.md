<p align="center">
  <img src="logo.svg" alt="AWS Backup Vault" width="80"/>
</p>

# AWS Backup Vault

Manage an [AWS Backup vault](https://docs.aws.amazon.com/aws-backup/latest/devguide/vaults.html)
— the encrypted container recovery points live in — as either a
standard vault with its attachable satellites or a logically
air-gapped vault for ransomware recovery.

## What Gets Managed

- **Exactly one vault type** (`metadata.name` is the vault name on
  either): the **standard** vault (optional KMS key, `force_destroy`
  drain-at-destroy) or the **logically air-gapped** vault (required
  7-day-minimum retention bounds; immutable apart from tags).
- **Standard-vault satellites** (they attach by vault name and only to
  standard vaults): **Vault Lock** (governance mode, or compliance
  mode via `changeable_for_days` — immutable once the cooling-off
  window passes), the **access policy**, and **event notifications**
  to an SNS topic.

An empty vault costs nothing — AWS bills recovery-point storage, not
vaults. Deleting a vault requires it to be EMPTY: the standard arm's
`force_destroy` drains recovery points at destroy; air-gapped recovery
points cannot be manually deleted (they age out by retention).

The backup plan that fills the vault is deliberately NOT part of this
kind — see [AwsBackupPlan](../awsbackupplan).

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
