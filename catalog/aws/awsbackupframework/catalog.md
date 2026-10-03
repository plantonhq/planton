# AWS Backup Framework

Deploys a Backup Audit Manager framework: a set of controls that continuously evaluate the account's backup posture — whether resources are covered by a backup plan, retention minimums are met, recovery points are encrypted, and restore times stay within bounds. Evaluations run on AWS Config, so the region needs an active Config recorder recording the backup resource types; without one the framework's deployment lands FAILED. Control evaluations materialize as Config rules named after the framework — this is the evidence layer compliance reviews ask for.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Backup Framework** — the Audit Manager framework holding every control with its parameters and scopes. Controls come from AWS's Backup Audit Manager vocabulary (e.g. `BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN`, `BACKUP_RECOVERY_POINT_MINIMUM_RETENTION_CHECK`); AWS derives one Config rule per control from it

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AWS Backup and AWS Config permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An ACTIVE AWS Config recorder in the region, recording the backup resource types. This dependency is real and silent: without it, control evaluations cannot run and the framework's deployment lands FAILED — and the provider treats FAILED as a completed apply, so the failure shows in the framework's `deployment_status`, not as an apply error.

## Deploy

### Console

Open the deployment store, find **AWS Backup Framework**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the framework name and its controls. Start from the **Backup Coverage Audit** preset in the [Presets](#presets) tab to audit coverage and retention, or the **Immutable Backup Posture** preset for encryption and vault-lock checks.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupFramework
metadata:
  name: backup-posture
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  frameworkName: backup_posture
  description: Are resources protected by a plan and retained long enough
  controls:
    - name: BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN
      scope:
        complianceResourceTypes:
          - EBS
          - RDS
    - name: BACKUP_RECOVERY_POINT_MINIMUM_RETENTION_CHECK
      inputParameters:
        - name: requiredRetentionDays
          value: "35"
```

```shell
planton apply -f backup-framework.yaml
```

This creates a framework named `backup_posture` that continuously checks whether EBS volumes and RDS databases are covered by a backup plan and whether every recovery point is retained at least 35 days. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a backup framework. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**frameworkName is not metadata.name** — AWS forbids hyphens in framework names (a letter first, then letters, digits, and underscores), which is stricter than Planton naming conventions, so the AWS name is an explicit spec field validated at apply time. Changing it forces replacement.

**Declare scope explicitly on resource-scoped controls** — A scope-less control like `BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN` is valid at AWS, but AWS materializes its default all-supported-types scope server-side and the pinned provider ships no diff suppression for that echo — every later plan proposes stripping a scope AWS will re-materialize. Declaring the resource types you care about keeps plans clean and pins the control's meaning: AWS's default list grows as new services gain Backup support, silently widening an unscoped control's coverage and its evaluation bill. Parameter-only checks (like the minimum-retention check) take no scope and are unaffected.

**The cost lever is controls × resources** — Every control becomes a Config rule, and Config bills per rule evaluation. A framework's cost scales with how many controls it carries and how many resources each one evaluates — scoping controls to the types or tags you actually audit is the control you have.

**Scope values are sticky** — AWS keeps scope values once set; clearing a `complianceResourceIds` or `complianceResourceTypes` list requires recreating the control entry. The scope's tag map accepts at most one key/value pair (the provider's documented limit — the spec rejects more at validate time).

**One framework per compliance regime** — Controls evaluate account-wide by default, so narrow with per-control scopes rather than multiplying frameworks. Control names are for_each keys on both engines: the spec forbids duplicates, and renaming a control entry destroys and recreates it.

**Edits cycle deployment status** — Every framework edit runs `UPDATE_IN_PROGRESS` for up to a few minutes; concurrent edits are rejected with conflicts the provider retries through. Check `deployment_status` after the first deploy in particular — that is where the missing-Config-recorder failure surfaces.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. A framework audits the account's backup posture as a whole — it has no schema edges to backup plans or vaults, and deploys with zero of either.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `framework_arn` | The framework's ARN | A backup report plan's framework list, scoping compliance reports to this framework's controls |
| `region` | The region the framework lives in | Frameworks are addressed by region + name; consumers reaching it off the ambient region need this alongside the ARN |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Coverage and retention audit** — a coverage control scoped to the resource types you back up, plus the parameterized minimum-retention check. This is the framework that answers "is everything we said we back up actually backed up, for long enough". Start from the **Backup Coverage Audit** preset.

**Immutable backup posture** — encryption, manual-deletion-disabled, and vault-lock controls that verify recovery points cannot be quietly deleted or stored unencrypted. Pairs with an air-gapped or locked vault strategy; the framework proves the posture the vaults claim. Start from the **Immutable Backup Posture** preset.

**Evidence pipeline** — wire the framework's `framework_arn` into a backup report plan for scheduled compliance evidence delivered to S3. The framework evaluates continuously; the report plan turns evaluations into auditor-ready documents.

## Works With

- [**AWS Config Recorder**](/infra-catalog/aws-config-recorder) — the hard prerequisite: framework evaluations run on Config, and the region's recorder must be active
- [**AWS Backup Report Plan**](/infra-catalog/aws-backup-report-plan) — consumes `framework_arn` to produce scheduled compliance reports
- [**AWS Backup Plan**](/infra-catalog/aws-backup-plan) — the coverage controls evaluate whether resources are protected by a plan
- [**AWS Backup Vault**](/infra-catalog/aws-backup-vault) — the vault-lock control verifies recovery points sit behind Vault Lock
