# AWS ECS Cluster

Deploys an ECS cluster: the logical boundary that groups services and tasks, decides where their containers run (Fargate, EC2 capacity providers, ECS Managed Instances, or a blend), and carries cluster-wide posture — Container Insights observability, ECS Exec auditing, Fargate storage encryption, and the Service Connect default namespace. Cost is driven by the tasks and instances the cluster schedules, not by the cluster itself.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ECS Cluster** — a managed cluster with optional Container Insights, ECS Exec audit configuration, customer-managed storage encryption, and a Service Connect default namespace
- **EC2 Capacity Providers** — one per `ec2CapacityProviders` entry, each wrapping a referenced auto-scaling group with ECS-managed scaling, termination protection, and draining; keyed by name so adding or removing one never disturbs the others
- **Managed Instances Capacity Providers** — one per `managedInstancesCapacityProviders` entry: ECS launches and retires the EC2 instances itself from attribute-based requirements (no auto-scaling group, AMI, or user data), bound to this cluster with an infrastructure role, an instance profile, and the subnets/security groups to launch into
- **Cluster Capacity Provider Associations** — the FARGATE / FARGATE_SPOT built-ins from `capacityProviders` plus every folded provider, associated together with the optional `defaultCapacityProviderStrategy`
- **AWS Tags** — resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An auto-scaling group** (only for EC2 capacity) — each `ec2CapacityProviders` entry references one. Its launch template must use an ECS-optimized AMI whose agent joins this cluster via user data. Enabling `managedTerminationProtection` additionally requires the group itself to enable new-instance scale-in protection.
- **Managed Instances IAM identities** (only for managed-instances capacity) — an infrastructure role trusting `ecs.amazonaws.com` with `AmazonECSInfrastructureRolePolicyForManagedInstances` (the identity applying the manifest needs `iam:PassRole` on it), and an instance profile for the launched instances.
- **KMS keys** (optional) — for encrypted ECS Exec sessions and/or customer-managed storage encryption. The Fargate ephemeral storage key's policy must grant the Fargate service principal decrypt/generate rights.
- **A CloudWatch log group and/or S3 bucket** (optional) — required when ECS Exec logging is `OVERRIDE`; the log group must already exist (ECS does not create it).
- **An AWS Cloud Map namespace** (optional) — required only when setting the Service Connect default namespace.

## Deploy

### Console

Open the deployment store, find **AWS ECS Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Fargate Cluster**, **Fargate Cost-Optimized Cluster**, or **EC2-Backed Cluster** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcsCluster
metadata:
  name: my-cluster
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  containerInsights: enhanced
  capacityProviders:
    - FARGATE
```

```shell
planton apply -f ecs-cluster.yaml
```

This creates a Fargate-ready cluster with enhanced Container Insights. An Infra Job tracks the provisioning in real time.

### InfraChart

When wiring EC2 capacity or encrypted exec sessions, use ValueFromRef to reference resources deployed in the same InfraPipeline:

```yaml
spec:
  ec2CapacityProviders:
    - name: general-purpose
      autoScalingGroupArn:
        valueFrom:
          kind: AwsAutoScalingGroup
          name: ecs-fleet
          fieldPath: status.outputs.autoscaling_group_arn
  executeCommandConfiguration:
    logging: DEFAULT
    kmsKeyId:
      valueFrom:
        kind: AwsKmsKey
        name: exec-encryption-key
        fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the auto-scaling group and KMS key first, then provisions the cluster with the resolved ARNs.

## Key Configuration

These are the most important decisions when configuring an ECS cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Capacity model** — Capacity is a spectrum. A serverless cluster associates the `FARGATE` / `FARGATE_SPOT` built-ins and never thinks about instances. An EC2-backed cluster defines `ec2CapacityProviders` — each wrapping a referenced auto-scaling group whose fleet ECS scales through managed scaling. A managed-instances cluster defines `managedInstancesCapacityProviders` — ECS launches and retires the EC2 instances itself from attribute-based requirements (memory, vCPUs, CPU manufacturer, accelerators), with a purchase model per provider (`ON_DEMAND`, `SPOT`, or `RESERVED` against capacity reservations). Services blend across all of it by provider name.

**Default capacity provider strategy** — What ECS uses when a service or run-task does not declare its own strategy. Every entry must name an associated provider (a Fargate built-in or a folded entry from either list), and only one entry may set a non-zero `base`. The classic cost posture: FARGATE base 1 / weight 1 plus FARGATE_SPOT weight 4 keeps one guaranteed on-demand task and runs ~80% of scaled capacity on Spot.

**Container Insights** — `containerInsights` accepts `enabled`, `enhanced` (container-level observability with automatic dashboards — the production choice), or `disabled`. Unset keeps the account default. Updatable in place.

**ECS Exec auditing** — `executeCommandConfiguration` controls where interactive `aws ecs execute-command` sessions are logged cluster-wide: `DEFAULT` rides each task's own awslogs configuration, `OVERRIDE` sends every session to the CloudWatch group and/or S3 bucket in `logConfiguration` (required with OVERRIDE, forbidden otherwise), and `NONE` disables session logging. Without the block, exec still works where services enable it — sessions are simply not centrally audited.

**Storage encryption** — `managedStorageConfiguration` applies customer-managed KMS keys to Fargate ephemeral task storage and other ECS-managed storage. Unset keeps AWS-owned keys (data is still encrypted).

**Service Connect namespace** — `serviceConnectNamespaceArn` names the Cloud Map namespace Service Connect uses by default for services in this cluster, letting a whole environment share one mesh namespace without per-service wiring.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsAutoScalingGroup** (per EC2 provider) | `ec2CapacityProviders[].autoScalingGroupArn` | `status.outputs.autoscaling_group_arn` |
| **AwsIamRole** (per managed-instances provider) | `managedInstancesCapacityProviders[].infrastructureRoleArn` | `status.outputs.role_arn` |
| **AwsIamInstanceProfile** (per managed-instances provider) | `managedInstancesCapacityProviders[].instanceLaunchTemplate.ec2InstanceProfileArn` | `status.outputs.instance_profile_arn` |
| **AwsSubnet** (per managed-instances provider) | `managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration.subnets[]` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (per managed-instances provider — AWS requires at least one; there is no VPC-default fall-back on this path) | `managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration.securityGroups[]` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `executeCommandConfiguration.kmsKeyId` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional) | `managedStorageConfiguration.fargateEphemeralStorageKmsKeyId` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional) | `managedStorageConfiguration.kmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_name` | Name of the ECS cluster | The AWS CLI and the ECS agent's `ECS_CLUSTER` setting |
| `cluster_arn` | ARN of the cluster | The join key — an AwsEcsService's `clusterArn` references this |
| `capacity_provider_names` | Every provider associated with the cluster (built-ins + folded names) | The vocabulary services can use in a capacity provider strategy |
| `capacity_provider_arns` | ARNs of the folded capacity providers (empty for Fargate-only clusters) | IAM policies, verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard Fargate cluster** — On-demand Fargate capacity with Container Insights. The starting point for production ECS deployments where Spot interruptions are not acceptable. Start from the **Standard Fargate Cluster** preset.

**Cost-optimized Fargate cluster** — Fargate + Fargate Spot with a weighted default strategy that runs ~80% of scaled tasks on Spot while guaranteeing one on-demand task. Start from the **Fargate Cost-Optimized Cluster** preset.

**EC2-backed cluster** — Wraps an auto-scaling group as a named capacity provider for workloads needing GPUs, special instance families, or EC2 unit economics; keeps the Fargate lane alongside. Start from the **EC2-Backed Cluster** preset.

**Managed-instances cluster** — EC2 economics with Fargate-grade hands-off: ECS owns the fleet from attribute-based requirements, no auto-scaling group or AMI pipeline to maintain. Start from the **Managed Instances Cluster** preset.

## Works With

- [**AWS Auto Scaling Group**](/infra-catalog/aws-auto-scaling-group) — provides the instance fleets behind EC2 capacity providers
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) / [**AWS IAM Instance Profile**](/infra-catalog/aws-iam-instance-profile) — the infrastructure and instance identities behind managed-instances capacity
- [**AWS Subnet**](/infra-catalog/aws-subnet) / [**AWS Security Group**](/infra-catalog/aws-security-group) — where managed instances launch and what guards them
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — encrypts ECS Exec sessions and ECS-managed storage
- [**AWS ECS Task Definition**](/infra-catalog/aws-ecs-task-definition) — the workload blueprints services deploy into this cluster
- [**AWS ECS Service**](/infra-catalog/aws-ecs-service) — runs and scales tasks inside this cluster, referencing its `cluster_arn` output
