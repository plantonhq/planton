# AWS Batch Compute Environment

Deploys a managed AWS Batch compute environment: the elastic pool of compute (EC2 On-Demand, EC2 Spot, Fargate, or Fargate Spot) that AWS Batch scales up and down to run submitted jobs. The compute environment is one node of the Batch resource graph — jobs are submitted to an [AWS Batch Job Queue](/infra-catalog/aws-batch-job-queue) (which maps onto one or more compute environments in preference order) using an [AWS Batch Job Definition](/infra-catalog/aws-batch-job-definition) as the container blueprint.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Batch Compute Environment** -- a MANAGED compute environment with the specified resource type (EC2, SPOT, FARGATE, or FARGATE_SPOT), vCPU scaling limits, VPC networking, and optional instance type selection and allocation strategy
- **Optional EKS attachment** -- when `eksConfiguration` is set, the environment schedules jobs as Kubernetes pods on your existing EKS cluster instead of ECS tasks (Batch on EKS)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

Job queues and fair-share scheduling policies are **separate Infra Components** ([AwsBatchJobQueue](/infra-catalog/aws-batch-job-queue), [AwsBatchSchedulingPolicy](/infra-catalog/aws-batch-scheduling-policy)) that reference this environment by ARN — one queue can span a Spot environment with an On-Demand overflow, and an environment can be replaced behind a queue with zero queue downtime.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **VPC subnets** in one or more Availability Zones for compute resource placement. Private subnets are recommended. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef.
- **Security groups** -- recommended for all resource types; required for FARGATE and FARGATE_SPOT. Provide IDs directly or reference an AwsSecurityGroup Infra Component.
- **An IAM instance profile** (EC2 and SPOT only) -- grants the ECS agent on each instance permission to communicate with AWS Batch. Reference an AwsIamInstanceProfile Infra Component or provide the profile ARN.
- **A Spot Fleet IAM role** (SPOT with the BEST_FIT allocation strategy only) -- allows EC2 Spot Fleet to request and manage Spot instances. The modern capacity-optimized strategies do not use Spot Fleet and need no role.
- **An EKS cluster with a prepared namespace** (Batch on EKS only) -- the cluster must exist and the namespace must be RBAC-configured for AWS Batch before the environment is created.

## Deploy

### Console

Open the deployment store, find **AWS Batch Compute Environment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Fargate Batch** preset in the [Presets](#presets) tab to pre-populate a serverless configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBatchComputeEnvironment
metadata:
  name: data-processing
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  computeResources:
    type: FARGATE
    maxVcpus: 256
    subnetIds:
      - value: "subnet-0a1b2c3d4e5f00001"
      - value: "subnet-0a1b2c3d4e5f00002"
    securityGroupIds:
      - value: "sg-0a1b2c3d4e5f00001"
```

```shell
planton apply -f batch-compute-environment.yaml
```

This creates a Fargate compute environment with a 256 vCPU ceiling. To start submitting jobs, create an AwsBatchJobQueue that references this environment's ARN. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the compute environment to a VPC and security group deployed in the same InfraPipeline:

```yaml
spec:
  computeResources:
    subnetIds:
      - valueFrom:
          kind: AwsSubnet
          name: private-a
          fieldPath: status.outputs.subnet_id
      - valueFrom:
          kind: AwsSubnet
          name: private-b
          fieldPath: status.outputs.subnet_id
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: batch-sg
          fieldPath: status.outputs.security_group_id
    instanceRole:
      valueFrom:
        kind: AwsIamInstanceProfile
        name: batch-instance-profile
        fieldPath: status.outputs.instance_profile_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, and instance profile first, then provisions the Batch compute environment with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Batch compute environment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Resource type** -- `FARGATE` provides serverless, zero-management compute -- AWS handles all infrastructure. Use `EC2` for GPU instances, custom AMIs, or sustained throughput. Use `SPOT` for interruption-tolerant workloads that trade reclaim risk for lower instance cost. `FARGATE_SPOT` combines serverless convenience with the Spot purchase model.

**vCPU scaling** -- `computeResources.maxVcpus` caps total concurrent capacity — it is the one sizing knob AWS allows updating on every environment. For EC2/SPOT, `minVcpus` controls the always-on baseline (keep it 0 to scale to zero when idle). Fargate types only use `maxVcpus` since AWS manages scaling dynamically.

**Instance selection** -- EC2 and SPOT types support `instanceTypes` (e.g., `["m5.xlarge", "c5.xlarge"]`) or `"optimal"` to let AWS Batch select from the C, M, and R families. `allocationStrategy` controls selection from the full six-strategy set — note that only `BEST_FIT_PROGRESSIVE`, `SPOT_CAPACITY_OPTIMIZED`, and `SPOT_PRICE_CAPACITY_OPTIMIZED` support in-place infrastructure updates; the others (including the AWS default `BEST_FIT`) force replacement on most compute changes.

**Custom launch surfaces** -- `launchTemplate` carries custom AMIs, user data, and IMDSv2 posture; `ec2Configurations` picks the image family (use `ECS_AL2_NVIDIA` alongside GPU instance types); `placementGroup` co-locates tightly-coupled multi-node jobs; `resourceTags` land on the launched EC2 instances for cost attribution.

**Batch on EKS** -- set `eksConfiguration` to schedule jobs as Kubernetes pods on an existing EKS cluster. Both fields are create-time only, and the namespace must be RBAC-prepared before deployment.

**Update policy** -- for EC2/SPOT in-place updates, `updatePolicy` decides what happens to running jobs when instances are replaced: wait for them to finish (`jobExecutionTimeoutMinutes`, AWS default 30) or terminate them immediately (`terminateJobsOnUpdate`).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** (optional) | `serviceRole` | `status.outputs.role_arn` |
| **AwsSubnet** | `computeResources.subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `computeResources.securityGroupIds` | `status.outputs.security_group_id` |
| **AwsIamInstanceProfile** (EC2/SPOT only) | `computeResources.instanceRole` | `status.outputs.instance_profile_arn` |
| **AwsIamRole** (SPOT + BEST_FIT only) | `computeResources.spotIamFleetRole` | `status.outputs.role_arn` |
| **AwsLaunchTemplate** (optional) | `computeResources.launchTemplate.launchTemplateId` | `status.outputs.launch_template_id` |
| **AwsEksCluster** (Batch on EKS only) | `eksConfiguration.eksClusterArn` | `status.outputs.cluster_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `compute_environment_arn` | Compute environment ARN | AwsBatchJobQueue `computeEnvironmentOrder` rows, IAM policies |
| `compute_environment_name` | Compute environment name | CLI commands, monitoring dashboards |
| `ecs_cluster_arn` | Underlying ECS cluster ARN managed by AWS Batch | Monitoring, debugging compute capacity |
| `status` | Compute environment status (VALID, INVALID) | Health checks, operational monitoring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Serverless batch processing** -- Fargate compute with 256 max vCPUs. Zero infrastructure management for container-based batch jobs. Start from the **Fargate Batch** preset.

**EC2 managed batch** -- On-demand EC2 instances with optimal instance selection, 0-512 vCPU scaling, and an update policy that waits for running jobs. Production configuration for GPU workloads or custom AMI requirements. Start from the **EC2 Managed Batch** preset.

**Spot cost-optimized batch** -- Spot instances with capacity-optimized allocation and 1024 max vCPUs. Maximizes throughput per dollar for large-scale, interruption-tolerant processing. Pair with an AwsBatchJobQueue that overflows to an On-Demand environment. Start from the **Spot Cost-Optimized Batch** preset.

## Works With

- [**AWS Batch Job Queue**](/infra-catalog/aws-batch-job-queue) -- routes submitted jobs onto this environment (and up to two others) in preference order
- [**AWS Batch Job Definition**](/infra-catalog/aws-batch-job-definition) -- the container blueprint jobs are submitted from
- [**AWS Batch Scheduling Policy**](/infra-catalog/aws-batch-scheduling-policy) -- fair-share capacity division for queues mapped onto this environment
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides subnets for compute resource placement across Availability Zones
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls network access for compute resources
- [**AWS IAM Instance Profile**](/infra-catalog/aws-iam-instance-profile) -- wraps the ECS instance role for EC2/SPOT environments
- [**AWS Launch Template**](/infra-catalog/aws-launch-template) -- custom AMIs, user data, and instance hardening
- [**AWS EKS Cluster**](/infra-catalog/aws-eks-cluster) -- the pod-scheduling target for Batch on EKS
