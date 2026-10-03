# AWS Backup Restore Testing Plan

Deploys an AWS Backup restore testing plan: scheduled, automated restore drills that prove recovery points actually restore, with per-resource-type selections deciding what gets tested under which IAM role. Each test restores into a temporary copy, holds it for a validation window, then deletes it — drills bill as real restores plus the copy's brief runtime. Results feed the restore-time metrics Backup Audit Manager can report on. A backup that has never been restored is a hope, not a backup; this is the component that turns hope into evidence.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Restore Testing Plan** — the drill schedule plus the recovery-point selection rule: which vaults to draw from, snapshot or continuous points, latest-or-random within a lookback window
- **Restore Testing Selections** — one per `selections` entry, each testing one protected resource type (EBS, EC2, RDS, S3, ...) under its IAM role, covering resources by explicit ARNs or by tag conditions

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AWS Backup permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Recovery points to test — a vault being filled by a backup plan. A testing plan over an empty vault schedules drills that never select anything.
- An IAM role trusting `backup.amazonaws.com` with restore permissions for each tested type (e.g. `ec2:CreateVolume` for EBS tests) — AWS's managed `AWSBackupServiceRolePolicyForRestores` policy is the usual grant.

## Deploy

### Console

Open the deployment store, find **AWS Backup Restore Testing Plan**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the drill schedule, recovery-point selection, and per-type selections. Start from the **Weekly Random Drills** preset in the [Presets](#presets) tab for the strongest default posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupRestoreTestingPlan
metadata:
  name: weekly-restore-drills
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  planName: weekly_restore_drills
  scheduleExpression: cron(0 5 ? * MON *)
  recoveryPointSelection:
    algorithm: RANDOM_WITHIN_WINDOW
    includeVaults: ["*"]
    recoveryPointTypes: ["SNAPSHOT"]
    selectionWindowDays: 30
  selections:
    - name: ebs_volumes
      protectedResourceType: EBS
      iamRoleArn:
        valueFrom:
          kind: AwsIamRole
          name: restore-testing-role
          fieldPath: status.outputs.role_arn
      protectedResourceArns: ["*"]
      validationWindowHours: 4
```

```shell
planton apply -f restore-testing-plan.yaml
```

This creates a plan that every Monday at 05:00 UTC restores a random snapshot from the last 30 days across all vaults, testing every EBS volume under the referenced role, holding each restored copy four hours for validation. An Infra Job tracks the provisioning in real time.

### InfraChart

When the testing plan deploys alongside its restore role in one chart, wire the role reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  planName: weekly_restore_drills
  scheduleExpression: cron(0 5 ? * MON *)
  recoveryPointSelection:
    algorithm: RANDOM_WITHIN_WINDOW
    includeVaults: ["*"]
    recoveryPointTypes: ["SNAPSHOT"]
  selections:
    - name: ebs_volumes
      protectedResourceType: EBS
      iamRoleArn:
        valueFrom:
          kind: AwsIamRole
          name: restore-testing-role
          fieldPath: status.outputs.role_arn
      protectedResourceArns: ["*"]
```

The InfraPipeline resolves the dependency graph, creates the role first, then provisions the testing plan with it.

## Key Configuration

These are the most important decisions when configuring a restore testing plan. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Random beats latest** — `RANDOM_WITHIN_WINDOW` exercises older recovery points too, proving the whole retention window is restorable rather than just yesterday's snapshot. `LATEST_WITHIN_WINDOW` is the right call only when the drill's purpose is verifying the newest backup of critical resources, on a tighter `selectionWindowDays`.

**ARNs or conditions, never both** — Each selection covers resources by explicit `protectedResourceArns` (`"*"` for everything of the type) or by `protectedResourceConditions` tag matching — exactly one, enforced at validate time because the provider enforces it resource-wide. Tag conditions are the shape that scales: teams opt resources into drills by tagging.

**Drills cost what restores cost** — Every test creates a real temporary resource. The cost drivers are drill frequency, how many selections run per drill, and `validationWindowHours` — size the window to what validation actually needs, since the restored copy lives (and bills) until it expires.

**Several knobs are one-way at AWS** — `scheduleExpressionTimezone`, `startWindowHours`, `selectionWindowDays`, `validationWindowHours`, and `excludeVaults` all keep a value once set; they can be changed but never cleared back to unset. Plan to flip them, not to remove them.

**Names are stricter than usual** — `planName` and selection names allow letters, digits, and underscores only — no hyphens, no periods (stricter than the backup plan's own rules). Changing `planName` forces replacement, and selection names are for_each keys: renaming one destroys and recreates it.

**Metadata overrides are lowercase** — AWS lowercases `restoreMetadataOverrides` keys on read, so author them lowercase (`availabilityzone`, not `AvailabilityZone`) or every later plan shows a phantom diff.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `selections[].iamRoleArn` | `status.outputs.role_arn` |

The vaults drilled against are named as ARN strings (or `"*"`) in `recoveryPointSelection.includeVaults`, not as typed references.

### What This Kind Provides

The single output, `restore_testing_plan_arn`, is an identity echo rather than a composition input — no catalog kind consumes it via ValueFromRef. The plan and its selections import by name (AWS assigns no separate ID); test results and restore-time metrics surface in the Backup console's restore testing view, which is where the drill evidence actually lives.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Weekly random drills across everything** — random selection over all vaults, a 30-day lookback, one selection per resource type with `"*"` coverage. The broad safety net that catches restore rot anywhere in the estate. Start from the **Weekly Random Drills** preset.

**Tag-scoped drills for the critical tier** — latest-within-a-short-window selection with tag conditions (`aws:ResourceTag/tier = critical`) and a longer validation window for deeper checks. Tighter, cheaper, and focused on the resources whose restore time is contractual. Start from the **Tag-Scoped Drills** preset.

**Drills as audit evidence** — pair the testing plan with a backup framework's restore-time controls and a report plan: drills generate restore-time metrics, the framework evaluates them, the report plan delivers the evidence to S3.

## Works With

- [**AWS Backup Plan**](/infra-catalog/aws-backup-plan) — creates the recovery points the drills restore
- [**AWS Backup Vault**](/infra-catalog/aws-backup-vault) — the vaults `includeVaults` draws recovery points from
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the restore-permissions role each selection runs under
- [**AWS Backup Framework**](/infra-catalog/aws-backup-framework) — evaluates the restore-time metrics the drills produce
