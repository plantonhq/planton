# AWS Batch Scheduling Policy

Deploys an AWS Batch fair-share scheduling policy: the rules that divide a job queue's compute capacity across share identifiers instead of processing jobs strictly first-in-first-out. Without one, a single team's burst of ten thousand jobs starves every other submitter on the queue.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Batch Scheduling Policy** -- a fair-share policy with the configured compute reservation, share decay window, and per-share weight distributions
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

One policy is a standalone, shareable object: many job queues can reference the same policy, so an organization's fairness rules are defined once and reused. Every dial updates in place on a live policy.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

No pre-existing AWS resources are required — the policy is self-contained. It takes effect only when an AWS Batch Job Queue in the same region references it AND jobs are submitted with a share identifier.

## Deploy

### Console

Open the deployment store, find **AWS Batch Scheduling Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the fairness dials, and the share-weight builder. Start from the **Team Fair Share** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBatchSchedulingPolicy
metadata:
  name: team-fair-share
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  computeReservation: 50
  shareDecaySeconds: 3600
  shareDistributions:
    - shareIdentifier: mltraining
      weightFactor: 0.5
    - shareIdentifier: analytics*
      weightFactor: 1
```

```shell
planton apply -f batch-scheduling-policy.yaml
```

This creates a policy that reserves headroom for quiet teams, remembers one hour of usage history, and gives ML training twice the capacity of the analytics share family. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a scheduling policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Share weights** -- the counter-intuitive core rule: LOWER weight means MORE capacity. A share with `weightFactor: 0.5` receives twice the capacity of a weight-1.0 share. Shares absent from the list weigh 1.0. Identifiers are alphanumeric (no hyphens or underscores); end one with `*` to treat a prefix family (e.g. `analytics*`) as one share. AWS allows up to 500 entries.

**Compute reservation** -- the percentage (0-99) of queue capacity held back for shares with NO currently-running jobs, so a quiet team's first job does not wait behind a busy team's backlog. The effective slice is `(reservation/100)^N` for N active shares — the headroom self-adjusts as more shares wake up.

**Share decay** -- the sliding window (up to 7 days) over which past usage counts against a share's fair allocation. Longer windows make fairness account for history ("you had the cluster all morning"); 0 considers only currently-running jobs.

**Submission is the other half** -- weights only apply to jobs submitted WITH a share identifier (SubmitJob's `shareIdentifier`). Bake the identifier into your submission tooling — a job without one fails on a fair-share queue.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — the policy is self-contained, and the relationship runs the other way: job queues reference it through their `schedulingPolicy` field.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `scheduling_policy_arn` | Scheduling policy ARN | AwsBatchJobQueue `schedulingPolicy` field |
| `scheduling_policy_name` | Scheduling policy name | CLI commands, monitoring dashboards |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team fair share** -- one policy with a 50% compute reservation and per-team weights, referenced by every shared queue in the organization. Start from the **Team Fair Share** preset.

**Workload-class priority** -- weight interactive/urgent shares low (more capacity) and backfill shares high, on a single queue — cheaper than running separate queues per class when the compute profile is identical.

## Works With

- [**AWS Batch Job Queue**](/infra-catalog/aws-batch-job-queue) -- references this policy to order jobs by share instead of first-in-first-out; note AWS's quirk that a queue's policy can be replaced but never removed
- [**AWS Batch Compute Environment**](/infra-catalog/aws-batch-compute-environment) -- provides the capacity the policy divides
- [**AWS Batch Job Definition**](/infra-catalog/aws-batch-job-definition) -- its `schedulingPriority` orders jobs WITHIN a share on fair-share queues
