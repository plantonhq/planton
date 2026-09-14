# AwsBackupVault

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsBackupVaultSpec defines the desired configuration for an AWS
Backup vault - the encrypted container that recovery points live
in.

A vault is EXACTLY ONE of two AWS vault types (AWS's own VaultType
discriminator): a standard backup vault, which carries the optional
lock/policy/notifications satellites, or a logically air-gapped
vault - an AWS-owned-key, immutably retained vault for ransomware
recovery that none of those satellites can attach to (the
provider's readers reject non-standard vaults by type). The vault
name is metadata.name on both arms (2-50 characters of letters,
digits, hyphens, and underscores).

An empty vault costs nothing - AWS Backup bills recovery-point
storage, not vaults. Deleting a vault requires it to be EMPTY of
recovery points (the standard arm's force_destroy can drain them;
air-gapped recovery points cannot be manually deleted - they age
out by retention, so destroy an air-gapped vault only after its
points expire).

## Example

```yaml
# Canonical AwsBackupVault example (hack/dev manifest and refgen
# Example source): a KMS-encrypted standard vault with a
# governance-mode lock, an access policy, and failure notifications.
# Literal ARNs stand in for composed references so the offline
# `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupVault
metadata:
  name: app-backup-vault
  id: app-backup-vault
  org: test-org
  env: dev
spec:
  region: us-west-2
  standard:
    kmsKeyArn:
      value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
    lock:
      minRetentionDays: 7
      maxRetentionDays: 365
    policy:
      Version: "2012-10-17"
      Statement:
        - Sid: DenyRecoveryPointDelete
          Effect: Deny
          Principal: "*"
          Action: backup:DeleteRecoveryPoint
          Resource: "*"
    notifications:
      snsTopicArn:
        value: arn:aws:sns:us-west-2:123456789012:backup-events
      events:
        - BACKUP_JOB_FAILED
        - RESTORE_JOB_FAILED
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.standard` | `AwsBackupVaultStandard` |  |  |  |
| `spec.standard.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.standard.forceDestroy` | `bool` |  |  |  |
| `spec.standard.lock` | `AwsBackupVaultLock` |  |  |  |
| `spec.standard.lock.changeableForDays` | `int32` |  |  |  |
| `spec.standard.lock.minRetentionDays` | `int32` |  |  |  |
| `spec.standard.lock.maxRetentionDays` | `int32` |  |  |  |
| `spec.standard.policy` | `object` |  |  |  |
| `spec.standard.notifications` | `AwsBackupVaultNotifications` |  |  |  |
| `spec.standard.notifications.snsTopicArn` | `string \| valueFrom` | yes |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.standard.notifications.events` | `[]string` | yes |  |  |
| `spec.airGapped` | `AwsBackupVaultAirGapped` |  |  |  |
| `spec.airGapped.minRetentionDays` | `int32` |  |  |  |
| `spec.airGapped.maxRetentionDays` | `int32` |  |  |  |
| `spec.airGapped.encryptionKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region the vault lives in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.standard

`AwsBackupVaultStandard`

The standard backup vault arm: the everyday vault type that
backup plans target by name, with optional KMS encryption, vault
lock, access policy, and event notifications.

### spec.standard.kmsKeyArn

`string | valueFrom`

KMS key that encrypts the vault's recovery points. Changing this
after creation forces a replacement, and once AWS fills in the
default service key the field cannot be cleared back to empty.
Unset = the AWS Backup service key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.standard.forceDestroy

`bool`

Delete ALL recovery points in the vault when this component is
destroyed. This is deploy-side delete behavior, not an AWS
attribute - AWS never reports it back, so it is invisible to
imports and asserted only at destroy time. Leave false unless the
vault is disposable: with it false, destroying a non-empty vault
fails (AWS refuses to delete vaults holding recovery points).

### spec.standard.lock

`AwsBackupVaultLock`

Vault Lock: write-once-read-many protection for every recovery
point in the vault. Omit to leave the vault unlocked. A lock is
all-or-replace: every lock field forces replacement of the lock
configuration.

- rule: a vault lock must set at least one of changeable_for_days, min_retention_days, or max_retention_days - a lock with none of them enforces nothing
- rule: max_retention_days must be greater than or equal to min_retention_days

### spec.standard.lock.changeableForDays

`int32` · optional (explicit presence)

Days the compliance-mode lock stays removable after creation
(minimum 3). SETTING THIS OPTS INTO COMPLIANCE MODE - once the
window passes, the lock is permanent until retention expires.
Leave unset for governance mode. AWS starts the countdown at
creation; the value cannot be changed afterward.

- rule: {"int32":{"gte":3}}

### spec.standard.lock.minRetentionDays

`int32` · optional (explicit presence)

Minimum retention (days) the vault enforces on recovery points -
backup plan rules that retain for less are rejected. Unset = no
minimum.

- rule: {"int32":{"gte":1}}

### spec.standard.lock.maxRetentionDays

`int32` · optional (explicit presence)

Maximum retention (days) the vault enforces on recovery points.
Unset = no maximum.

- rule: {"int32":{"gte":1}}

### spec.standard.policy

`object`

Resource-based access policy on the vault (who may act on the
vault and its recovery points), as an IAM policy document.

### spec.standard.notifications

`AwsBackupVaultNotifications`

Vault event notifications fanned out to an SNS topic. Omit to
leave notifications unconfigured.

### spec.standard.notifications.snsTopicArn

`string | valueFrom` · required

The SNS topic that receives the events. The topic's access policy
must allow backup.amazonaws.com to publish.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.standard.notifications.events

`[]string` · required

Which vault events to publish (the AWS BackupVaultEvent
vocabulary at the pinned provider).

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"in":["BACKUP_JOB_STARTED","BACKUP_JOB_COMPLETED","BACKUP_JOB_SUCCESSFUL","BACKUP_JOB_FAILED","BACKUP_JOB_EXPIRED","RESTORE_JOB_STARTED","RESTORE_JOB_COMPLETED","RESTORE_JOB_SUCCESSFUL","RESTORE_JOB_FAILED","COPY_JOB_STARTED","COPY_JOB_SUCCESSFUL","COPY_JOB_FAILED","RECOVERY_POINT_MODIFIED","BACKUP_PLAN_CREATED","BACKUP_PLAN_MODIFIED","CONTINUOUS_BACKUP_INTERRUPTED","RECOVERY_POINT_INDEX_COMPLETED","RECOVERY_POINT_INDEX_DELETED","RECOVERY_POINT_INDEXING_FAILED","EKS_RESTORE_OBJECT_FAILED","EKS_RESTORE_OBJECT_SKIPPED","EKS_BACKUP_OBJECT_FAILED"]}}}}

### spec.airGapped

`AwsBackupVaultAirGapped`

The logically air-gapped vault arm: a hardened vault whose
recovery points cannot be deleted or re-encrypted - retention is
locked in at creation and EVERY field forces replacement.
Recovery points reach it via a backup plan rule's
target_logically_air_gapped_backup_vault_arn (rules always need a
standard target vault too).

- rule: max_retention_days must be greater than or equal to min_retention_days

### spec.airGapped.minRetentionDays

`int32`

Minimum retention (days) enforced on every recovery point,
at least 7 (the AWS floor for air-gapped vaults). Required.

- rule: {"int32":{"gte":7}}

### spec.airGapped.maxRetentionDays

`int32`

Maximum retention (days) enforced on every recovery point.
Required; must be at least the minimum.

- rule: {"int32":{"gte":7}}

### spec.airGapped.encryptionKeyArn

`string | valueFrom`

KMS key encrypting the vault. Fixed at creation; once AWS fills
in its default the field cannot be cleared. Unset = the AWS-owned
key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBackupVault, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vault_arn` | `string` | The vault's ARN (either arm) - what backup plan copy actions and air-gapped targeting reference. |
| `status.outputs.vault_name` | `string` | The vault's name (also the provider's import ID on either arm) - what backup plan rules target. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.standard.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.standard.notifications.snsTopicArn` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.airGapped.encryptionKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBackupPlan | `spec.rules[].targetVaultName` | `status.outputs.vault_name` |
| AwsBackupPlan | `spec.rules[].targetLogicallyAirGappedBackupVaultArn` | `status.outputs.vault_arn` |
| AwsBackupPlan | `spec.rules[].copyActions[].destinationVaultArn` | `status.outputs.vault_arn` |

## See Also

- [Overview](../README.md)
