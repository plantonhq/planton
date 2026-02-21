---
title: "Batch Compute Environment"
description: "Batch Compute Environment deployment documentation"
icon: "package"
order: 100
componentName: "awsbatchcomputeenvironment"
---

# AWS Batch Compute Environment

Deploys a MANAGED AWS Batch compute environment with bundled job queues and an optional fair-share scheduling policy. Supports EC2, SPOT, FARGATE, and FARGATE_SPOT resource types with automatic vCPU scaling, VPC networking, and multi-queue priority routing. The component provisions the compute infrastructure, one or more job queues, and an optional scheduling policy in a single resource definition.

## What Gets Created

When you deploy an AwsBatchComputeEnvironment resource, OpenMCF provisions:

- **Compute Environment** — a `batch.ComputeEnvironment` of type `MANAGED` with the specified resource type (EC2/SPOT/FARGATE/FARGATE_SPOT), vCPU limits, VPC subnets, security groups, and optional update policy
- **Job Queues** — one `batch.JobQueue` per entry in `jobQueues`, each referencing the compute environment with configurable priority, state, and optional job-state time-limit actions for automatic cancellation of stuck jobs
- **Scheduling Policy** (optional) — a `batch.SchedulingPolicy` with fair-share configuration when `schedulingPolicy` is provided, attached to all bundled job queues for capacity distribution across share identifiers

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **At least one VPC subnet** (private subnets recommended) — use `AwsVpc` to provision
- **A security group** allowing outbound access for containers — use `AwsSecurityGroup` to provision
- **For EC2/SPOT types**: an ECS instance profile IAM role with `AmazonEC2ContainerServiceforEC2Role` policy
- **For SPOT type**: a Spot Fleet IAM role with `AmazonEC2SpotFleetTaggingRole` policy

## Quick Start

Create a file `batch.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsBatchComputeEnvironment
metadata:
  name: my-batch
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsBatchComputeEnvironment.my-batch
spec:
  region: us-west-2
  computeResources:
    type: FARGATE
    maxVcpus: 256
    subnetIds:
      - value: subnet-0a1b2c3d4e5f00001
      - value: subnet-0a1b2c3d4e5f00002
    securityGroupIds:
      - value: sg-0a1b2c3d4e5f00001
  jobQueues:
    - name: default
      priority: 1
```

Deploy:

```shell
openmcf apply -f batch.yaml
```

This creates a serverless Fargate compute environment with up to 256 vCPUs of capacity and a single `default` job queue. AWS manages all compute infrastructure — no EC2 instances, patching, or AMI management required.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the compute environment will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `computeResources` | object | Infrastructure configuration for the compute environment | Required |
| `computeResources.type` | string | Compute resource type | Must be `EC2`, `SPOT`, `FARGATE`, or `FARGATE_SPOT` |
| `computeResources.maxVcpus` | int32 | Maximum vCPU capacity | >= 1 |
| `computeResources.subnetIds` | list(StringValueOrRef) | VPC subnets for compute resources | At least 1 required |
| `jobQueues` | list(object) | Job queues routing to this compute environment | At least 1 required |
| `jobQueues[].name` | string | Queue name | 1-128 chars, alphanumeric/hyphen/underscore, starts with alphanumeric |
| `jobQueues[].priority` | int32 | Dispatch priority (higher = higher priority) | Required |

### Optional Fields — Top Level

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `state` | string | `ENABLED` | Compute environment state: `ENABLED` or `DISABLED` |
| `serviceRole` | StringValueOrRef | Service-linked role | IAM role for AWS Batch to make API calls |
| `updatePolicy.terminateJobsOnUpdate` | bool | false | Whether to terminate running jobs during infrastructure updates |
| `updatePolicy.jobExecutionTimeoutMinutes` | int32 | — | Max wait time for jobs during updates (1-360 minutes) |
| `schedulingPolicy` | object | — | Fair-share scheduling policy (see below) |

### Optional Fields — Compute Resources

| Field | Type | Applies To | Description |
|-------|------|------------|-------------|
| `computeResources.minVcpus` | int32 | EC2/SPOT | Minimum vCPUs to maintain (default: 0) |
| `computeResources.desiredVcpus` | int32 | EC2/SPOT | Initial desired vCPUs |
| `computeResources.securityGroupIds` | list(StringValueOrRef) | All | VPC security groups |
| `computeResources.instanceTypes` | list(string) | EC2/SPOT | Instance types (e.g., `["optimal"]`, `["m5.xlarge", "c5.xlarge"]`) |
| `computeResources.allocationStrategy` | string | EC2/SPOT | `BEST_FIT_PROGRESSIVE`, `SPOT_CAPACITY_OPTIMIZED`, or `SPOT_PRICE_CAPACITY_OPTIMIZED` |
| `computeResources.instanceRole` | StringValueOrRef | EC2/SPOT | ECS instance profile ARN (**required** for EC2/SPOT) |
| `computeResources.ec2KeyPair` | string | EC2/SPOT | SSH key pair name |
| `computeResources.bidPercentage` | int32 | SPOT | Max % of On-Demand price (0-100) |
| `computeResources.spotIamFleetRole` | StringValueOrRef | SPOT | Spot Fleet IAM role (**required** for SPOT) |
| `computeResources.launchTemplate` | object | EC2/SPOT | Custom launch template (id or name + optional version) |
| `computeResources.ec2Configurations` | list(object) | EC2/SPOT | AMI customization (max 2 entries) |
| `computeResources.resourceTags` | map(string) | EC2/SPOT | Tags for launched compute resources |

### Optional Fields — Job Queue

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `jobQueues[].state` | string | `ENABLED` | Queue state: `ENABLED` or `DISABLED` |
| `jobQueues[].jobStateTimeLimitActions` | list(object) | — | Auto-cancel jobs stuck in a state |
| `jobQueues[].jobStateTimeLimitActions[].action` | string | — | Action to take: `CANCEL` |
| `jobQueues[].jobStateTimeLimitActions[].maxTimeSeconds` | int32 | — | Time threshold (600-86400 seconds) |
| `jobQueues[].jobStateTimeLimitActions[].reason` | string | — | Human-readable reason |
| `jobQueues[].jobStateTimeLimitActions[].state` | string | — | Job state to monitor (e.g., `RUNNABLE`) |

### Optional Fields — Scheduling Policy

| Field | Type | Description |
|-------|------|-------------|
| `schedulingPolicy.computeReservation` | int32 | % of vCPUs reserved for new share identifiers (0-99) |
| `schedulingPolicy.shareDecaySeconds` | int32 | Usage history decay period (0-604800 seconds) |
| `schedulingPolicy.shareDistributions` | list(object) | Weight per share identifier |
| `schedulingPolicy.shareDistributions[].shareIdentifier` | string | Unique share identifier (supports `*` wildcard suffix) |
| `schedulingPolicy.shareDistributions[].weightFactor` | double | Relative share weight (0.0001-999.9999) |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `compute_environment_arn` | string | ARN of the compute environment |
| `compute_environment_name` | string | Name of the compute environment |
| `ecs_cluster_arn` | string | ARN of the underlying ECS cluster |
| `status` | string | Compute environment status |
| `job_queue_arns.<name>` | string | Per-queue ARN (one entry per queue name) |
| `scheduling_policy_arn` | string | Scheduling policy ARN (if created) |

## Presets

| Name | Description |
|------|-------------|
| [01-fargate-batch](presets/01-fargate-batch.yaml) | Serverless Fargate, single queue, zero-management |
| [02-ec2-managed-batch](presets/02-ec2-managed-batch.yaml) | EC2 with optimal instances, two priority queues, update policy |
| [03-spot-cost-optimized-batch](presets/03-spot-cost-optimized-batch.yaml) | Spot instances, fair-share scheduling, multi-team capacity |

## Design Decisions

**Bundling scope.** The compute environment and job queues are bundled because a compute environment without a queue is incomplete infrastructure — you cannot submit jobs without a queue. Job definitions are excluded because they represent application-level workloads with independent lifecycles (versioned, frequently updated, reusable across queues).

**MANAGED only.** UNMANAGED compute environments (where the user manages compute) and EKS-based compute environments are deferred to v2 as they are niche use cases that add significant complexity.

**Scheduling policy as top-level.** The scheduling policy is defined at the spec level and attached to all bundled queues. Per-queue scheduling policies with external references are deferred to v2.

**State defaults.** Both compute environment and job queue states default to `ENABLED` via the OpenMCF middleware default mechanism. State validation is delegated to the AWS API to keep the proto schema simple and forward-compatible.
