# AwsBatchComputeEnvironment

Deploy and manage AWS Batch compute environments with bundled job queues and optional fair-share scheduling policies.

## Overview

AWS Batch manages the provisioning and scaling of compute resources for batch processing workloads. This component creates a MANAGED compute environment — the infrastructure layer where batch jobs execute — along with one or more job queues that route submitted jobs to the compute.

**Supported compute types:**

| Type | Description | Use Case |
|------|-------------|----------|
| `FARGATE` | Serverless containers | Zero infrastructure management |
| `FARGATE_SPOT` | Serverless at Spot pricing | Cost-sensitive serverless workloads |
| `EC2` | On-demand EC2 instances | GPU, custom AMI, sustained capacity |
| `SPOT` | EC2 Spot instances | Large-scale cost optimization |

**Bundled resources:**
- **Compute environment** — the infrastructure (instance types, vCPU limits, VPC networking)
- **Job queues** — routing layer (priority, state, time-limit actions)
- **Scheduling policy** (optional) — fair-share capacity distribution across teams/workloads

**Not included:** Job definitions — they represent application-level workloads with independent lifecycles and should be managed separately.

## Prerequisites

- An AWS VPC with private subnets (use `AwsVpc`)
- A security group allowing outbound access (use `AwsSecurityGroup`)
- For EC2/SPOT: an ECS instance profile IAM role (use `AwsIamRole`)
- For SPOT: a Spot Fleet IAM role (use `AwsIamRole`)

## Quick Start

### Minimal (Fargate)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsBatchComputeEnvironment
metadata:
  name: my-batch-env
spec:
  computeResources:
    type: FARGATE
    maxVcpus: 256
    subnetIds:
      - value: subnet-abc123
      - value: subnet-def456
    securityGroupIds:
      - value: sg-abc123
  jobQueues:
    - name: default
      priority: 1
```

### Production (EC2 with multi-queue)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsBatchComputeEnvironment
metadata:
  name: production-batch
spec:
  computeResources:
    type: EC2
    maxVcpus: 512
    minVcpus: 0
    instanceTypes:
      - m5.xlarge
      - c5.xlarge
      - optimal
    allocationStrategy: BEST_FIT_PROGRESSIVE
    subnetIds:
      - value: subnet-abc123
      - value: subnet-def456
    securityGroupIds:
      - value: sg-abc123
    instanceRole:
      value: arn:aws:iam::123456789012:instance-profile/ecsInstanceRole
  updatePolicy:
    terminateJobsOnUpdate: false
    jobExecutionTimeoutMinutes: 60
  jobQueues:
    - name: critical
      priority: 10
    - name: background
      priority: 1
      jobStateTimeLimitActions:
        - action: CANCEL
          maxTimeSeconds: 7200
          reason: "Job stuck in RUNNABLE for over 2 hours"
          state: RUNNABLE
```

## Spec Fields

### Top-Level

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `state` | string | No | `ENABLED` | Compute environment state: `ENABLED` or `DISABLED` |
| `serviceRole` | StringValueOrRef | No | Service-linked role | IAM role for AWS Batch API calls |
| `computeResources` | object | **Yes** | — | Infrastructure configuration |
| `updatePolicy` | object | No | — | Controls behavior during infrastructure updates |
| `jobQueues` | list | **Yes** (min 1) | — | Job queues routing to this compute environment |
| `schedulingPolicy` | object | No | — | Fair-share scheduling policy |

### Compute Resources

| Field | Type | Required | Applies To | Description |
|-------|------|----------|------------|-------------|
| `type` | string | **Yes** | All | `EC2`, `SPOT`, `FARGATE`, `FARGATE_SPOT` |
| `maxVcpus` | int | **Yes** | All | Maximum vCPU capacity |
| `minVcpus` | int | No | EC2/SPOT | Minimum vCPUs (default: 0) |
| `desiredVcpus` | int | No | EC2/SPOT | Initial desired vCPUs |
| `subnetIds` | list | **Yes** | All | VPC subnets for compute resources |
| `securityGroupIds` | list | No | All | VPC security groups |
| `instanceTypes` | list | No | EC2/SPOT | EC2 instance types (e.g., `optimal`) |
| `allocationStrategy` | string | No | EC2/SPOT | Instance selection strategy |
| `instanceRole` | StringValueOrRef | **Yes**\* | EC2/SPOT | ECS instance profile ARN |
| `ec2KeyPair` | string | No | EC2/SPOT | SSH key pair name |
| `bidPercentage` | int | No | SPOT | Max % of On-Demand price (0-100) |
| `spotIamFleetRole` | StringValueOrRef | **Yes**\* | SPOT | Spot Fleet IAM role ARN |
| `launchTemplate` | object | No | EC2/SPOT | Custom launch template |
| `ec2Configurations` | list | No | EC2/SPOT | AMI customization (max 2) |
| `resourceTags` | map | No | EC2/SPOT | Tags for launched compute resources |

\* Required for the indicated compute type.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `compute_environment_arn` | string | ARN of the compute environment |
| `compute_environment_name` | string | Name of the compute environment |
| `ecs_cluster_arn` | string | ARN of the underlying ECS cluster |
| `status` | string | Compute environment status |
| `job_queue_arns` | map | Queue name → ARN mapping |
| `scheduling_policy_arn` | string | Scheduling policy ARN (if created) |

## Presets

| Preset | Description |
|--------|-------------|
| `01-fargate-batch` | Serverless Fargate, single queue, minimal config |
| `02-ec2-managed-batch` | EC2 with optimal instances, two priority queues |
| `03-spot-cost-optimized-batch` | Spot instances, fair-share scheduling, cost-optimized |

## Deferred to v2

- UNMANAGED compute environments (user-managed compute)
- EKS-based compute environments (Kubernetes namespace integration)
- External scheduling policy references (ARN-based, shared across components)
- Placement groups
