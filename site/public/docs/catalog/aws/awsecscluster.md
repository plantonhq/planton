---
title: "ECS Cluster"
description: "ECS Cluster deployment documentation"
icon: "package"
order: 100
componentName: "awsecscluster"
---

# AWS ECS Cluster

Deploys an AWS ECS cluster configured for Fargate workloads with optional capacity provider strategies, CloudWatch Container Insights, and ECS Exec auditing. The component handles capacity provider attachment and default strategy configuration for cost-optimized task placement.

## What Gets Created

When you deploy an AwsEcsCluster resource, OpenMCF provisions:

- **ECS Cluster** — an `ecs.Cluster` resource with the specified name, optional Container Insights setting, and optional ECS Exec configuration
- **Cluster Capacity Providers** — an `ecs.ClusterCapacityProviders` resource (created only when `capacityProviders` is specified) that attaches FARGATE and/or FARGATE_SPOT providers with an optional default strategy defining base/weight distribution

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An existing VPC and subnets** if you plan to deploy ECS services into this cluster (the cluster itself does not require networking)
- **A KMS key** if enabling encrypted ECS Exec sessions
- **A CloudWatch log group** or **S3 bucket** if using OVERRIDE logging for ECS Exec

## Quick Start

Create a file `ecs-cluster.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: my-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEcsCluster.my-cluster
spec: {}
```

Deploy:

```shell
openmcf apply -f ecs-cluster.yaml
```

This creates a basic ECS cluster with no capacity providers attached and Container Insights disabled. Services deployed into this cluster will need to specify their own launch type.

## Configuration Reference

### Required Fields

The `AwsEcsClusterSpec` has no strictly required fields. An empty `spec: {}` creates a valid cluster. All fields below are optional.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enableContainerInsights` | `bool` | `false` | Enables CloudWatch Container Insights for cluster-level monitoring. Recommended for production (incurs CloudWatch costs). |
| `capacityProviders` | `string[]` | `[]` | Capacity providers to attach. Valid values: `FARGATE`, `FARGATE_SPOT`. Items must be unique. |
| `defaultCapacityProviderStrategy` | `CapacityProviderStrategy[]` | `[]` | Base/weight distribution for tasks across capacity providers. See sub-fields below. |
| `defaultCapacityProviderStrategy[].capacityProvider` | `string` | — | Name of the capacity provider. Valid values: `FARGATE`, `FARGATE_SPOT`. |
| `defaultCapacityProviderStrategy[].base` | `int32` | `0` | Minimum number of tasks guaranteed on this provider. Must be >= 0. |
| `defaultCapacityProviderStrategy[].weight` | `int32` | — | Relative weight for scaling beyond the base. Must be > 0. |
| `executeCommandConfiguration.logging` | `enum` | `LOGGING_UNSPECIFIED` | Logging behavior for ECS Exec. Valid values: `LOGGING_UNSPECIFIED` (exec disabled), `DEFAULT`, `NONE`, `OVERRIDE`. |
| `executeCommandConfiguration.kmsKeyId` | `string` | `""` | KMS key ID for encrypting exec session data. |
| `executeCommandConfiguration.logConfiguration.cloudWatchLogGroupName` | `string` | `""` | CloudWatch log group for exec audit logs. Only used when logging is `OVERRIDE`. |
| `executeCommandConfiguration.logConfiguration.cloudWatchEncryptionEnabled` | `bool` | `false` | Encrypt CloudWatch logs using the specified KMS key. |
| `executeCommandConfiguration.logConfiguration.s3BucketName` | `string` | `""` | S3 bucket for exec audit logs. Only used when logging is `OVERRIDE`. |
| `executeCommandConfiguration.logConfiguration.s3KeyPrefix` | `string` | `""` | S3 key prefix for organizing exec log files. |
| `executeCommandConfiguration.logConfiguration.s3EncryptionEnabled` | `bool` | `false` | Encrypt S3 logs using the specified KMS key. |

## Examples

### Fargate-Only Cluster with Container Insights

A production cluster using only on-demand Fargate capacity with monitoring enabled:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: prod-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEcsCluster.prod-cluster
spec:
  enableContainerInsights: true
  capacityProviders:
    - FARGATE
```

### Cost-Optimized Cluster with Spot Capacity

A cluster that uses 20% on-demand Fargate for baseline stability and 80% Fargate Spot for cost savings on scaled tasks:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: cost-optimized-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEcsCluster.cost-optimized-cluster
spec:
  enableContainerInsights: true
  capacityProviders:
    - FARGATE
    - FARGATE_SPOT
  defaultCapacityProviderStrategy:
    - capacityProvider: FARGATE
      base: 1
      weight: 1
    - capacityProvider: FARGATE_SPOT
      base: 0
      weight: 4
```

### Cluster with ECS Exec and Default Logging

Enables ECS Exec for debugging containers with AWS-managed logging:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: debug-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEcsCluster.debug-cluster
spec:
  enableContainerInsights: true
  capacityProviders:
    - FARGATE
  executeCommandConfiguration:
    logging: DEFAULT
```

### Full-Featured Cluster with Custom Exec Logging

Production cluster with Spot capacity, Container Insights, and ECS Exec audit logs sent to both CloudWatch and S3 with KMS encryption:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: full-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEcsCluster.full-cluster
spec:
  enableContainerInsights: true
  capacityProviders:
    - FARGATE
    - FARGATE_SPOT
  defaultCapacityProviderStrategy:
    - capacityProvider: FARGATE
      base: 1
      weight: 1
    - capacityProvider: FARGATE_SPOT
      base: 0
      weight: 3
  executeCommandConfiguration:
    logging: OVERRIDE
    kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/abcd-1234-efgh-5678
    logConfiguration:
      cloudWatchLogGroupName: /ecs/exec-audit
      cloudWatchEncryptionEnabled: true
      s3BucketName: my-ecs-exec-logs
      s3KeyPrefix: exec-sessions/
      s3EncryptionEnabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_name` | `string` | Name of the created ECS cluster |
| `cluster_arn` | `string` | ARN of the created ECS cluster |
| `cluster_capacity_providers` | `string[]` | List of capacity providers associated with the cluster (exported per-index) |

## Related Components

- [AwsEcsService](/docs/catalog/aws/awsecsservice) — deploys Fargate services into this cluster
- [AwsAlb](/docs/catalog/aws/awsalb) — provides load balancing for services running in the cluster
- [AwsVpc](/docs/catalog/aws/awsvpc) — provides the network infrastructure for ECS services
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides encryption keys for ECS Exec session data
