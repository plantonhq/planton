# AWS Backup Vault

Deploys an AWS Backup vault — the encrypted container recovery points live in — as exactly one of two AWS vault types: a standard vault for everyday recovery points, or a logically air-gapped vault whose contents cannot be deleted before retention expires. The standard arm carries the attachable satellites — Vault Lock (governance or compliance mode), a resource-based access policy, and event notifications to SNS — none of which can attach to an air-gapped vault. The vault name is `metadata.name` on both arms, and it is what backup plan rules target.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Backup Vault** — a standard vault encrypted with the AWS Backup service key, or the KMS key named in `kmsKeyArn` (created when the `standard` arm is set)
- **Logically Air-Gapped Vault** — an immutably retained vault whose recovery points cannot be manually deleted; they age out by retention (created when the `airGapped` arm is set)
- **Vault Lock Configuration** — write-once-read-many protection on every recovery point, created only when `standard.lock` is set
- **Vault Policy** — the vault's resource-based IAM policy, created only when `standard.policy` is set
- **Vault Notifications** — vault events fanned out to an SNS topic, created only when `standard.notifications` is set

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AWS Backup permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Nothing for the vault itself. For customer-managed encryption, a KMS key (wired via `kmsKeyArn` or `encryptionKeyArn`); for notifications, an SNS topic whose access policy allows `backup.amazonaws.com` to publish — an ARN alone is not enough, and events silently never arrive without the policy.

## Deploy

### Console

Open the deployment store, find **AWS Backup Vault**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the vault-type choice with its satellites. Start from the **Standard Vault** preset in the [Presets](#presets) tab for the everyday shape, or the **Air-Gapped Vault** preset for the ransomware-recovery tier.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupVault
metadata:
  name: app-backups
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  standard:
    lock:
      minRetentionDays: 30
      maxRetentionDays: 365
```

```shell
planton apply -f backup-vault.yaml
```

This creates a standard vault named `app-backups`, encrypted with the AWS Backup service key, with a governance-mode Vault Lock holding every recovery point between 30 and 365 days. An Infra Job tracks the provisioning in real time.

### InfraChart

When the vault deploys alongside its encryption key in one chart, wire the key reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  standard:
    kmsKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: backup-key
        fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, creates the key first, then provisions the vault encrypted with it.

## Key Configuration

These are the most important decisions when configuring a backup vault. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Standard or air-gapped — exactly one, chosen forever** — The spec mirrors AWS's own `VaultType` discriminator: configure exactly one of `standard` / `airGapped`. An air-gapped vault is immutable apart from tags — name, retention bounds, and encryption key all force replacement — and it fills only through a backup plan rule's air-gapped copy target, never by direct writes (the rule still needs a standard target vault). Its recovery points cannot be manually deleted; they age out by retention, which also means the vault itself destroys only after every point expires.

**Vault Lock mode is decided by one field** — `changeableForDays` unset means governance mode: privileged principals can remove the lock later. Set (minimum 3) means compliance mode: once the cooling-off window passes, the lock is immutable — nobody, not the account root, not AWS support, can remove it or delete the vault until every recovery point exceeds max retention. AWS never reports `changeableForDays` back, so the mode choice is invisible to imports and must be re-stated in config. Rehearse in governance mode; flip to compliance only with retention bounds you can live with for years.

**The KMS key is fixed at creation** — On both arms, changing the key after creation forces vault replacement, and once AWS fills in its default the field cannot be cleared back to empty. Pick the key before the first recovery point lands.

**forceDestroy is deploy-side, not an AWS attribute** — With it false (the default), destroying a vault that holds recovery points fails: AWS refuses to delete non-empty vaults. Setting it true drains all recovery points at destroy. AWS never reports the flag back — an imported vault always shows it false — so treat it as a per-deploy assertion and leave it false for anything holding real backups.

**Lock retention bounds police your plans** — `minRetentionDays` and `maxRetentionDays` on the lock reject backup plan rules that retain outside the window. Air-gapped vaults require both bounds (minimum 7 days, the AWS floor). Vault count is never a cost lever; recovery-point storage and the retention windows in your backup plans are.

**A permissions failure can masquerade as a missing vault** — The provider maps `AccessDeniedException` to not-found on the vault's read path. If a vault "disappears" from plans, check IAM before assuming deletion.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `standard.kmsKeyArn` | `status.outputs.key_arn` |
| **AwsKmsKey** | `airGapped.encryptionKeyArn` | `status.outputs.key_arn` |
| **AwsSnsTopic** | `standard.notifications.snsTopicArn` | `status.outputs.topic_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vault_name` | The vault's name (also the provider's import ID) | Backup plan rules target vaults by name |
| `vault_arn` | The vault's ARN | Backup plan copy actions and air-gapped copy targeting |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One standard vault per environment** — the plain container backup plans target, encrypted with the service key, no lock. Add Vault Lock later once retention policy settles; the lock attaches by vault name without replacing the vault. Start from the **Standard Vault** preset.

**Air-gapped ransomware-recovery tier** — a logically air-gapped vault beside the standard one, filled by a backup plan rule's air-gapped copy target. Compromised credentials can delete the standard vault's points but not these; the trade is rigidity — retention is locked at creation and every change is a replacement. Start from the **Air-Gapped Vault** preset.

**Compliance lock rollout** — run governance mode first and verify no backup plan rule violates the retention window, then set `changeableForDays` to opt into compliance mode with a cooling-off window long enough to catch mistakes. The countdown starts at creation and the value cannot be changed afterward.

## Works With

- [**AWS Backup Plan**](/infra-catalog/aws-backup-plan) — rules target this vault by name; copy actions and air-gapped targeting use its ARN
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for either vault arm, wired via `kmsKeyArn` or `encryptionKeyArn`
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) — the destination for vault event notifications on standard vaults
