# AWS Data Lifecycle Manager Policy

Deploys one Data Lifecycle Manager policy — account-level automation that creates, retains, copies, archives, and deletes EBS snapshots (or EBS-backed AMIs) on a schedule, selecting its targets by tags. The policy references no specific volume or snapshot: it acts on whatever carries the target tags when a schedule fires, so coverage is dynamic — tag a new volume and it is covered the next time the schedule runs. A policy runs in exactly one of two modes: AWS's simplified default posture ("snapshot every volume or instance daily-ish, keep for N days", with exclusions) or the full custom engine (tag-targeted, up to four named schedules with create, retain, archive, copy, share, and deprecate rules — or an event-driven policy reacting to snapshots shared into the account).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **DLM Lifecycle Policy** — one policy in the configured mode: `defaultPolicy` renders AWS's simplified posture, `customPolicy` renders the full schedule or event-based engine. The provider's policy language is derived from the arm (SIMPLIFIED for default mode, STANDARD for custom) so a spec field can never contradict it. The policy's enabled/disabled state and the IAM role DLM acts through are part of it.
- **AWS Tags** — resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, including DLM and EC2 permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An execution role trusting `dlm.amazonaws.com`** — AWS's service default `AWSDataLifecycleManagerDefaultRole` (create it once per account with `aws dlm create-default-role`), or an AwsIamRole Infra Component carrying the documented DLM permissions. The policy is rejected without it, and it silently stops working if the role later loses permissions.
- **KMS keys in the destination regions** (only for encrypted cross-region copies) — each copy rule's key lives in ITS target region.

## Deploy

### Console

Open the deployment store, find **AWS Data Lifecycle Manager Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Daily Volume Backups** preset in the [Presets](#presets) tab for the workhorse shape: every volume tagged `backup: daily` gets a nightly snapshot with a rolling 14-count retention.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsDlmLifecyclePolicy
metadata:
  name: daily-volume-backups
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  description: daily snapshots of tagged volumes
  executionRoleArn:
    value: arn:aws:iam::123456789012:role/AWSDataLifecycleManagerDefaultRole
  customPolicy:
    resourceTypes:
      - VOLUME
    targetTags:
      backup: daily
    schedules:
      - name: daily
        copyTags: true
        createRule:
          intervalHours: 24
          times:
            - "03:00"
        retainRule:
          count: 14
```

```shell
planton apply -f dlm-policy.yaml
```

This creates a policy that snapshots every volume tagged `backup: daily` at 03:00 UTC and keeps the newest 14, with source tags riding along. An Infra Job tracks the provisioning in real time.

### InfraChart

When the policy deploys alongside its execution role in one chart, wire the role reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  executionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: dlm-execution
      fieldPath: status.outputs.role_arn
  customPolicy:
    resourceTypes:
      - VOLUME
    targetTags:
      backup: daily
    schedules:
      - name: daily
        createRule:
          intervalHours: 24
        retainRule:
          count: 14
```

The InfraPipeline resolves the dependency graph, provisions the role first, then creates the policy acting through it.

## Key Configuration

These are the most important decisions when configuring a DLM policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Default mode or custom mode** — `defaultPolicy` is AWS's account-wide blanket, and AWS allows exactly ONE default policy per resource type per region: it is an account-global singleton in disguise. Custom tag-targeted policies coexist freely — prefer them anywhere two teams share an account.

**The policy never fails loudly** — a policy whose role lost permissions, or whose target tags match nothing, does not error your deploys; it silently stops producing snapshots (the policy state flips to ERROR at best). Alarm on snapshot AGE, not on the policy resource — the absence of new snapshots is the real signal.

**Namespace your target tags** — two policies sharing the same `targetTags` is an AWS-side rejection that surfaces at apply, not at plan. Use per-policy tags (`backup: hourly`, `backup: daily`) instead of reusing one `backup: true` everywhere.

**Retain by count or by age, per schedule** — `retainRule` takes `count` (keep the newest N, deleted as new ones arrive) XOR `interval` + `intervalUnit` (keep for a duration). Count-based retention is self-limiting under any cadence; age-based retention grows with the cadence — decide which failure mode you prefer when the schedule changes.

**Retention math compounds across rules** — a schedule keeping 24 hourlies, cross-region-copying with its own retention, and archiving after that is three storage meters from one schedule. Read a policy as a cost graph: every rule that keeps something has its own meter, and `fastRestoreRule` adds a per-zone-hour meter on top.

**copyTags is a schedule-replacing decision** — changing a schedule's `copyTags` replaces the WHOLE schedule at the provider, and retention counting restarts from the new schedule's snapshots. Decide tag propagation before the first fire, not after a month of retained history.

**Pause without deleting** — `disabled: true` stops the schedules from firing while existing snapshots stay untouched; flipping it back resumes the cadence. Deleting the policy also leaves existing snapshots alone.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `executionRoleArn` | `status.outputs.role_arn` |
| **AwsKmsKey** (optional, per copy rule) | `customPolicy.schedules[].crossRegionCopyRules[].cmkArn` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional, event-based copies) | `customPolicy.action.crossRegionCopies[].cmkArn` | `status.outputs.key_arn` |

### What This Kind Provides

`status.outputs` carries `policy_id` (the `policy-...` identifier, also the provider's import ID) and `policy_arn`. They identify the policy for audit and import purposes; no downstream Infra Component composes on a DLM policy, because the policy targets volumes by tags rather than being referenced by them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Tag-and-forget daily backups** — one custom policy, one daily schedule, count-based retention, source tags copied onto snapshots. Tag a volume and it is covered — no per-volume wiring, ever. The judgment is in the tag namespace: give each policy its own tag value so policies never overlap. Start from the **Daily Volume Backups** preset.

**Tiered retention with DR** — two schedules in one policy: rolling hourlies for oops-recovery, plus a weekly kept for months whose copies replicate encrypted to another region. Every retention rule is its own storage meter, so the cost graph is explicit and deliberate. Start from the **Tiered Retention With DR** preset.

**DR intake from another account** — an event-based policy (`policyType: EVENT_BASED_POLICY`) that reacts to snapshots shared INTO this account and copies them to your own regions with your own retention. This is the receiving half of a cross-account backup contract; the sharing half lives in the source account's share rules.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role DLM assumes, wired via `executionRoleArn`
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — the destination-region keys encrypted cross-region copies use, wired via `cmkArn`
- [**AWS EBS Volume**](/infra-catalog/aws-ebs-volume) — the tagged targets; the policy discovers them by tag at fire time rather than by reference
- [**AWS EBS Snapshot**](/infra-catalog/aws-ebs-snapshot) — the deliberate one-off captures that complement the policy's recurring cadence
