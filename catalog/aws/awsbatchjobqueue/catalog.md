# AWS Batch Job Queue

Deploys an AWS Batch job queue: the place jobs are submitted to, and the routing layer that decides WHICH compute environment runs them. A queue maps onto up to three compute environments in preference order — which is what makes the canonical Batch cost pattern (Spot first, On-Demand overflow) one row away.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Batch Job Queue** -- with the configured priority, accept/drain state, and the ordered compute-environment mapping
- **Fair-share attachment** -- when `schedulingPolicy` references an [AWS Batch Scheduling Policy](/infra-catalog/aws-batch-scheduling-policy), jobs are ordered by share identifier instead of first-in-first-out
- **Stuck-job fuses** -- optional `jobStateTimeLimitActions` that automatically CANCEL jobs stuck at the head of the queue in RUNNABLE past a threshold
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

Jobs themselves are submitted against the queue at runtime (SubmitJob) using an [AWS Batch Job Definition](/infra-catalog/aws-batch-job-definition); the queue holds them until a mapped environment has capacity.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **At least one Batch compute environment** in the same region, in the VALID state. Reference an AwsBatchComputeEnvironment Infra Component or provide a literal ARN.
- **A scheduling policy** (optional) for fair-share ordering. Reference an AwsBatchSchedulingPolicy Infra Component or provide a literal ARN — and note that once attached, a policy can be replaced but never removed.

## Deploy

### Console

Open the deployment store, find **AWS Batch Job Queue**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the queue dials, and the environment mapping. Start from the **Single Environment Queue** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBatchJobQueue
metadata:
  name: batch-queue
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  priority: 10
  computeEnvironmentOrder:
    - order: 1
      computeEnvironment:
        valueFrom:
          kind: AwsBatchComputeEnvironment
          name: data-processing
          fieldPath: status.outputs.compute_environment_arn
  jobStateTimeLimitActions:
    - action: CANCEL
      maxTimeSeconds: 3600
      reason: MISCONFIGURATION:JOB_RESOURCE_REQUIREMENT
      state: RUNNABLE
```

```shell
planton apply -f batch-job-queue.yaml
```

This creates a queue mapped onto one compute environment with a stuck-job fuse: jobs whose resource requirements the environment can never satisfy are cancelled after an hour. An Infra Job tracks the provisioning in real time.

### InfraChart

The Spot-first overflow pattern wires two environments in preference order:

```yaml
spec:
  computeEnvironmentOrder:
    - order: 1
      computeEnvironment:
        valueFrom:
          kind: AwsBatchComputeEnvironment
          name: spot-env
          fieldPath: status.outputs.compute_environment_arn
    - order: 2
      computeEnvironment:
        valueFrom:
          kind: AwsBatchComputeEnvironment
          name: on-demand-env
          fieldPath: status.outputs.compute_environment_arn
  schedulingPolicy:
    valueFrom:
      kind: AwsBatchSchedulingPolicy
      name: team-fair-share
      fieldPath: status.outputs.scheduling_policy_arn
```

The InfraPipeline resolves the dependency graph, deploys both environments and the policy first, then provisions the queue with the resolved ARNs.

## Key Configuration

These are the most important decisions when configuring a job queue. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Environment mapping** -- up to three environments in preference order; the scheduler tries the LOWEST order first and overflows upward when capacity runs out. All mapped environments must be one compute family: EC2-based (EC2/SPOT) or Fargate-based (FARGATE/FARGATE_SPOT), never mixed.

**Priority** -- BETWEEN queues sharing a compute environment, not between jobs: when two queues compete for the same capacity, the higher value schedules first. Ordering within one queue comes from submission order or the fair-share policy.

**Fair-share policy** -- attach an AwsBatchSchedulingPolicy to order jobs by share identifier. The one asymmetric decision on this kind: a live queue can REPLACE its policy but never REMOVE it — returning to first-in-first-out requires recreating the queue.

**Stuck-job fuses** -- one entry per stuck-cause (`reason` is a matcher against AWS's own three causes, not free text): insufficient capacity, exceeds-environment-maximum, or unsatisfiable resource requirements. `action` and `state` are fixed by AWS at CANCEL/RUNNABLE.

**Drain switch** -- set `state: DISABLED` to reject new submissions while queued jobs finish — the safe first step of maintenance or decommissioning.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsBatchComputeEnvironment** (1-3) | `computeEnvironmentOrder[].computeEnvironment` | `status.outputs.compute_environment_arn` |
| **AwsBatchSchedulingPolicy** (optional) | `schedulingPolicy` | `status.outputs.scheduling_policy_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `job_queue_arn` | Job queue ARN | SubmitJob calls, EventBridge Batch targets, IAM policies |
| `job_queue_name` | Job queue name | CLI commands, monitoring dashboards |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single environment** -- the standard starting point: one queue on one environment with a stuck-job fuse. Start from the **Single Environment Queue** preset.

**Spot overflow** -- Spot environment at order 1, On-Demand at order 2: lower-cost Spot capacity when it exists, reliable capacity when it does not. Start from the **Spot-First With On-Demand Overflow** preset.

**Two-tier priority** -- an urgent queue (priority 100) and a backfill queue (priority 1) on the same environment: urgent jobs always claim capacity first.

## Works With

- [**AWS Batch Compute Environment**](/infra-catalog/aws-batch-compute-environment) -- the capacity this queue dispatches onto, in preference order
- [**AWS Batch Job Definition**](/infra-catalog/aws-batch-job-definition) -- the container blueprint jobs are submitted from at runtime
- [**AWS Batch Scheduling Policy**](/infra-catalog/aws-batch-scheduling-policy) -- fair-share ordering within this queue (replaceable, never removable)
