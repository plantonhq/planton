# AWS SSM Association

Deploys an AWS Systems Manager State Manager association that binds an SSM document — AWS-managed or your own — to instances by tag, ID, or resource group on a schedule. The document reference accepts any document name, so "run AWS-RunPatchBaseline on every instance tagged env=prod nightly" needs no custom document, while customer-owned documents wire in by reference. Changing the document replaces the association; every other change creates a new association version in place. When a run fails, the association reports compliance findings at the severity you choose.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **State Manager Association** — the document-to-target binding with its schedule, rate controls, and compliance posture. AWS identifies it by a generated UUID, not by name; `associationName` is display metadata for the State Manager console.
- **S3 Output Delivery** — configured only when `outputLocation` is set; without it, command output from association runs is not stored anywhere.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with permissions to manage SSM associations. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Managed nodes** — instances running the SSM agent with an instance profile that grants SSM access. The association itself deploys fine before any exist; tag-based targets pick up matching instances on the next interval.
- **A customer document** (only when `documentName` references your own) — an AwsSsmDocument the association runs instead of an AWS-managed one.
- **An output bucket** (only for `outputLocation`) — an S3 bucket to receive command output.

## Deploy

### Console

Open the deployment store, find **AWS SSM Association**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields covering the document, targets, schedule, and rate controls. Start from the **Tagged Patch Scan** preset in the [Presets](#presets) tab for the most common shape: an AWS-managed patch scan bound to tagged instances.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSsmAssociation
metadata:
  name: nightly-patch-scan
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  documentName:
    value: AWS-RunPatchBaseline
  associationName: nightly-patch-scan
  parameters:
    Operation: Scan
  targets:
    - key: tag:env
      values:
        - prod
  scheduleExpression: cron(0 2 ? * * *)
  applyOnlyAtCronInterval: true
  complianceSeverity: HIGH
  maxConcurrency: 10%
  maxErrors: 5%
```

```shell
planton apply -f ssm-association.yaml
```

This binds AWS's own `AWS-RunPatchBaseline` document to every instance tagged `env=prod`, scanning for missing patches nightly at 02:00 — installing nothing — with failures reported as HIGH-severity compliance findings. An Infra Job tracks the provisioning in real time.

### InfraChart

When the association runs a customer document deployed in the same chart, wire the document reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  documentName:
    valueFrom:
      kind: AwsSsmDocument
      name: agent-install
      fieldPath: status.outputs.document_name
  targets:
    - key: tag:env
      values:
        - prod
  scheduleExpression: rate(7 days)
```

The InfraPipeline resolves the dependency graph, deploys the document first, then binds the association to it.

## Key Configuration

These are the most important decisions when configuring an association. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The document is the one replacement trigger** — Changing `documentName` destroys and recreates the association (a new UUID, fresh run history); every other edit creates a new association version in place. Decide the document up front; iterate freely on everything else.

**Version pinning** — `documentVersion` unset means `$DEFAULT`: releasing a new default version of the document changes what runs with no association edit — the right posture when the document component owns its release process. Pin a concrete number for strict rollouts, or `$LATEST` only when you genuinely want unreleased edits running on your fleet.

**Immediate first run or not** — Without `applyOnlyAtCronInterval: true`, State Manager applies the document immediately on create and on every association change, then on schedule. For patch operations or anything disruptive, set it (it requires a cron `scheduleExpression`) so the first run waits for the window you chose.

**Rate controls before large fleets** — `maxConcurrency` and `maxErrors` take absolute counts or percentages. `maxErrors: "0"` is the strictest honest setting: one failure stops further scheduling for that interval. On tag-targeted fleets that grow over time, prefer percentages for both so the controls scale with the fleet.

**`waitForSuccessTimeoutSeconds` is a create-time gate only** — It fails the deploy unless the association's first run succeeds within the window, and it is never read back from AWS. Never set it when no matching targets exist yet: the wait times out by construction.

**Compliance posture** — `complianceSeverity` is what a failed run reports as; pick the severity your compliance dashboards should alarm on. `syncCompliance: MANUAL` hands compliance recording to an external process entirely (via PutComplianceItems) — the association stops writing its own compliance records, so don't set it casually.

**Targets can be legitimately absent** — Unset `targets` is for documents that manage their own scope, such as Automation runbooks driven by `automationTargetParameterName`, which fans the target resource IDs into a runbook parameter for rate-controlled automation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSsmDocument** | `documentName` | `status.outputs.document_name` |
| **AwsS3Bucket** | `outputLocation.s3BucketName` | `status.outputs.bucket_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `association_id` | The AWS-generated UUID that is the association's real identity | Addressing the association in CLI/API operations (start-associations-once, describe calls) |
| `association_arn` | The association's ARN | IAM policy statements scoping who may update or delete it |

`document_name` is also echoed — the resolved document the association bound to. It is an input echo, useful for confirming what a reference resolved to rather than as a composition input; downstream components compose with the document itself, not with the association.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Tagged patch visibility first** — Bind `AWS-RunPatchBaseline` with `Operation: Scan` to tagged instances before automating any installs: know what is missing, at what severity, with zero change risk. New instances join coverage by carrying the tag — no association edit. Start from the **Tagged Patch Scan** preset.

**Your runbook on a schedule** — Reference your own AwsSsmDocument and let `$DEFAULT` version tracking do the releases: the document component publishes a new default version, and the association runs it on the next interval with no edit. The trade is that a bad document release propagates fleet-wide on schedule — pin `documentVersion` to a number when rollouts must be deliberate. Start from the **Custom Document Schedule** preset.

**Change Calendar gating** — `calendarNames` makes the association run only when every named calendar is open. This is the clean way to encode freeze windows (quarter-close, launch weeks) without editing schedules: the calendar owns the freeze, the association keeps its cadence.

## Works With

- [**AWS SSM Document**](/infra-catalog/aws-ssm-document) — the customer-owned document the association runs, wired via the `documentName` reference
- [**AWS SSM Patch Baseline**](/infra-catalog/aws-ssm-patch-baseline) — governs what `AWS-RunPatchBaseline` approves when the association runs patch operations
- [**AWS SSM Maintenance Window**](/infra-catalog/aws-ssm-maintenance-window) — the alternative scheduling vehicle for disruptive operations like patch installs, with tighter cutoff control
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — receives command output when `outputLocation` is set
