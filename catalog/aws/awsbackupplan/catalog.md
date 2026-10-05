# AWS Backup Plan

Deploys an AWS Backup plan: the scheduled rules that create recovery points, and the resource selections that decide which resources those rules cover. Rules control when backups run, which vault recovery points land in, lifecycle to cold storage and expiry, cross-vault and cross-region copies, air-gapped vault targeting, and GuardDuty malware scans; selections pair an IAM role with ARN-, tag-, or condition-based coverage. Selections fold into the plan on both engines, so AWS's refusal to delete a plan with live selections is never the author's ordering problem.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Backup Plan** — the named plan holding every rule: schedules, backup windows, continuous (point-in-time) backups, lifecycle, copy actions, air-gapped targeting, and scan actions. Windows VSS application-consistent backups (EC2 only at the pinned provider) ride along via `advancedBackupSettings`
- **Backup Selections** — one per `selections` entry, each assigning resources to the plan under the IAM role AWS Backup assumes. Created and destroyed with the plan

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AWS Backup permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- A standard backup vault for every rule's `targetVaultName` — rules cannot exist without one.
- For selections, an IAM role trusting `backup.amazonaws.com` (AWS's managed `AWSBackupServiceRolePolicyForBackup` policy is the usual grant). The trust is checked server-side: a role without it fails selection creation after a couple of minutes of provider retries.
- (Only for malware scanning) a GuardDuty detector in the region and a scanner role for `scanSetting.scannerRoleArn`.

## Deploy

### Console

Open the deployment store, find **AWS Backup Plan**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the plan's rules and selections. Start from the **Daily Tagged Backups** preset in the [Presets](#presets) tab for the workhorse shape: one daily rule plus a tag-driven selection.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupPlan
metadata:
  name: daily-backups
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  rules:
    - name: daily
      targetVaultName:
        valueFrom:
          kind: AwsBackupVault
          name: app-backups
          fieldPath: status.outputs.vault_name
      schedule: cron(0 5 ? * * *)
      lifecycle:
        deleteAfterDays: 35
  selections:
    - name: tagged
      iamRoleArn:
        valueFrom:
          kind: AwsIamRole
          name: backup-service-role
          fieldPath: status.outputs.role_arn
      selectionTags:
        - type: STRINGEQUALS
          key: backup
          value: "true"
```

```shell
planton apply -f backup-plan.yaml
```

This creates a plan with one rule firing daily at 05:00 UTC into the referenced vault, expiring recovery points after 35 days, covering every resource tagged `backup=true` under the referenced role. An Infra Job tracks the provisioning in real time.

### InfraChart

When the plan deploys alongside its vault and role in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  rules:
    - name: daily
      targetVaultName:
        valueFrom:
          kind: AwsBackupVault
          name: app-backups
          fieldPath: status.outputs.vault_name
      schedule: cron(0 5 ? * * *)
      lifecycle:
        deleteAfterDays: 35
  selections:
    - name: tagged
      iamRoleArn:
        valueFrom:
          kind: AwsIamRole
          name: backup-service-role
          fieldPath: status.outputs.role_arn
      selectionTags:
        - type: STRINGEQUALS
          key: backup
          value: "true"
```

The InfraPipeline resolves the dependency graph, creates the vault and role first, then provisions the plan against them.

## Key Configuration

These are the most important decisions when configuring a backup plan. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Retention is the bill** — Recovery-point storage is what AWS Backup charges for, so `lifecycle.deleteAfterDays` on every rule is the cost lever that matters. Unset lifecycle means recovery points are kept in warm storage until someone deletes them manually. For long retention, `coldStorageAfterDays` moves points to the cheaper cold tier — but AWS requires a 90-day minimum cold stay, so `deleteAfterDays` must exceed `coldStorageAfterDays` by at least 90 (the spec rejects violations at validate time).

**Continuous backups cap at 35 days** — `enableContinuousBackup: true` buys point-in-time restore for the resource types that support it, but AWS caps continuous retention at 35 days. Keep `deleteAfterDays` within that on continuous rules or the plan is rejected.

**Selections replace, never update** — Every selection field forces replacement at the provider: a coverage change swaps the selection object (same name, new AWS-generated ID) while backup history is unaffected. Two fields are stickier: once `notResources` or `conditions` is set, AWS keeps the value — clearing it back to empty requires replacing the selection deliberately.

**Names are identity** — Rule and selection names are for_each keys on both engines (the spec enforces uniqueness), so renaming an entry destroys and recreates it. The plan's own name is `metadata.name` and changing it forces plan replacement — AWS identifies the plan by a generated UUID, not the name.

**Air-gapped targeting still needs a standard vault** — A rule copying into a logically air-gapped vault via `targetLogicallyAirGappedBackupVaultArn` must also name its standard `targetVaultName`; the air-gapped copy is additional, not a substitute.

**Malware scanning scales with scanned data** — Both the plan-wide `scanSetting` and per-rule `scanActions` run GuardDuty scans billed by scanned volume. Scope `resourceTypes` deliberately and prefer `INCREMENTAL_SCAN` where a full scan of every recovery point is not required.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsBackupVault** | `rules[].targetVaultName` | `status.outputs.vault_name` |
| **AwsBackupVault** | `rules[].targetLogicallyAirGappedBackupVaultArn` | `status.outputs.vault_arn` |
| **AwsBackupVault** | `rules[].copyActions[].destinationVaultArn` | `status.outputs.vault_arn` |
| **AwsIamRole** | `selections[].iamRoleArn` | `status.outputs.role_arn` |
| **AwsIamRole** | `scanSetting.scannerRoleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `plan_id` | The plan's AWS-generated UUID — its real identity at AWS and the provider's import ID | Addressing the plan in AWS CLI/API operations and imports |
| `plan_arn` | The plan's ARN | IAM policies scoping backup administration to specific plans |

`plan_version` (a new ID on every plan update) and `selection_ids` (AWS-generated selection IDs keyed by selection name, importable as `{plan_id}|{selection_id}`) are also exported; they are operational echoes for auditing and import, not composition inputs — no catalog kind consumes them via ValueFromRef.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Daily tagged backups** — one daily rule with a bounded lifecycle, coverage driven by a single tag so teams opt resources in by tagging rather than editing the plan. The selection is the union of its matchers; keep it tag-only and the plan never needs a change when infrastructure grows. Start from the **Daily Tagged Backups** preset.

**DR cold storage** — a weekly rule that transitions recovery points to cold storage after 30 days and expires them at a year, plus a copy action into a second-region vault with its own lifecycle. The copy's lifecycle is independent of the source's — budget both. Start from the **DR Cold Storage** preset.

**Ransomware-recovery copies** — a rule that targets its standard vault and additionally copies into a logically air-gapped vault. Compromised credentials can delete standard recovery points but not air-gapped ones; the trade is the air-gapped vault's locked retention window, which the rule's lifecycle must respect.

## Works With

- [**AWS Backup Vault**](/infra-catalog/aws-backup-vault) — every rule's target, plus copy-action and air-gapped destinations
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the role AWS Backup assumes for selections, and the GuardDuty scanner role
- [**AWS Backup Restore Testing Plan**](/infra-catalog/aws-backup-restore-testing-plan) — proves the recovery points this plan creates actually restore
- [**AWS Backup Framework**](/infra-catalog/aws-backup-framework) — audits that resources are covered by a plan and that retention meets policy
