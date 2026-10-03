# AWS ECS Service

Runs a task definition as a long-lived service: the ECS scheduler keeps the desired number of task copies running in a cluster, replaces unhealthy ones, rolls deployments, and wires tasks into load balancers and service discovery. The service composes onto its neighbors instead of bundling them — the task definition (WHAT runs) is a referenced AwsEcsTaskDefinition, the cluster (WHERE it runs) is a referenced AwsEcsCluster, and traffic arrives through referenced AwsLbTargetGroup nodes that AwsLbListener / AwsLbListenerRule nodes route into. Because the task-definition reference resolves to a revision-carrying ARN, registering a new revision (a new image tag) rolls the service on its next deployment — the deploy pipeline is the resource graph itself.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ECS Service** -- the scheduler keeping the desired count of tasks running, with its networking, load-balancer registrations, deployment guards, and discovery wiring
- **Application Auto Scaling resources** -- the scalable target and target-tracking policies (with their AWS-managed CloudWatch alarms) when autoscaling is configured
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the service

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **ECS Cluster** -- an AwsEcsCluster referenced by its `cluster_arn` output.
- **Task Definition** -- an AwsEcsTaskDefinition referenced by its `task_definition_arn` output (the recommended wiring — each new revision rolls the service).
- **Networking** -- AwsSubnet and AwsSecurityGroup resources for the task ENIs, referenced by their outputs.
- **Target Groups** (request-serving services) -- AwsLbTargetGroup resources (target type `ip`), each already associated with an AwsLbListener / AwsLbListenerRule.

### AWS Account

- **ECS permissions** -- the credentials used by the Provider Connection must have `ecs:CreateService`, `ecs:UpdateService`, `ecs:DeleteService`, and `ecs:DescribeServices`, plus `iam:PassRole` on the task definition's roles and `application-autoscaling:*` for autoscaling.
- **Listener association** -- AWS requires each referenced target group to already be associated with a load-balancer listener at service creation; deploy the listener (rules) first.
- **Egress for private tasks** -- tasks in private subnets reach ECR and AWS APIs through a NAT gateway or VPC endpoints.

## Deploy

### Console

Open the deployment store, find **AWS ECS Service**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Web Service** preset in the [Presets](#presets) tab for a working baseline.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcsService
metadata:
  name: api
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  clusterArn:
    valueFrom:
      kind: AwsEcsCluster
      name: platform
      fieldPath: status.outputs.cluster_arn
  taskDefinition:
    valueFrom:
      kind: AwsEcsTaskDefinition
      name: api
      fieldPath: status.outputs.task_definition_arn
  desiredCount: 2
  launchType: FARGATE
  network:
    subnets:
      - valueFrom:
          kind: AwsSubnet
          name: private-subnet-a
          fieldPath: status.outputs.subnet_id
      - valueFrom:
          kind: AwsSubnet
          name: private-subnet-b
          fieldPath: status.outputs.subnet_id
    securityGroups:
      - valueFrom:
          kind: AwsSecurityGroup
          name: api-tasks
          fieldPath: status.outputs.security_group_id
  loadBalancers:
    - targetGroupArn:
        valueFrom:
          kind: AwsLbTargetGroup
          name: api
          fieldPath: status.outputs.target_group_arn
      containerName: api
      containerPort: 8080
  deploymentCircuitBreaker:
    enable: true
    rollback: true
```

```shell
planton apply -f ecs-service.yaml
```

This keeps two copies of the `api` task running behind the target group, guarded by the circuit breaker. An Infra Job tracks the provisioning in real time.

### InfraChart

When the service deploys alongside its cluster and task definition in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  clusterArn:
    valueFrom:
      kind: AwsEcsCluster
      name: platform
      fieldPath: status.outputs.cluster_arn
  taskDefinition:
    valueFrom:
      kind: AwsEcsTaskDefinition
      name: api
      fieldPath: status.outputs.task_definition_arn
  desiredCount: 2
  launchType: FARGATE
  network:
    subnets:
      - valueFrom:
          kind: AwsSubnet
          name: private-subnet-a
          fieldPath: status.outputs.subnet_id
    securityGroups:
      - valueFrom:
          kind: AwsSecurityGroup
          name: api-tasks
          fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the cluster, task definition, and network first, then starts the service against their outputs.

## Key Configuration

These are the most important decisions when configuring an ECS service. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The composition rolls deploys** -- reference the task definition by its ARN output and every new revision (a new image tag) rolls the service on its next deployment. A literal `family:revision` pins one exact revision instead.

**Capacity is a spectrum** -- name a `launchType` (FARGATE is the default posture) OR blend providers with `capacityProviderStrategy` (e.g. FARGATE base 1/weight 1 + FARGATE_SPOT weight 4 keeps one guaranteed on-demand task and runs ~80% of scaled capacity on Spot). The two are mutually exclusive.

**Deployments are guarded in layers** -- the circuit breaker (enable + rollback) stops a rollout whose tasks keep failing; alarm gating fails a deployment when a referenced CloudWatch alarm fires; the BLUE_GREEN strategy adds an alternate target group, canary/linear traffic shifting, bake time, and Lambda lifecycle hooks for the most sensitive services.

**Load-balancer wiring only registers IPs** -- the target group, listener, and rules are their own first-class resources; each `loadBalancers` entry tells ECS which container/port registers into which target group. Workers and queue consumers leave it empty.

**Service Connect over legacy registries** -- Service Connect gives sidecar-free service-to-service discovery ("call http://orders") with telemetry, retries, per-request access logs (`accessLogConfiguration`, TEXT or JSON), header-based test-traffic routing for blue/green (`clientAlias.testTrafficRules`), and optional Private-CA TLS. Keep legacy Cloud Map `serviceRegistries` only where consumers already resolve its DNS name.

**VPC Lattice instead of a load balancer** -- each `vpcLatticeConfigurations` entry registers the tasks' named port into a Lattice target group, with an infrastructure role ECS assumes to manage the registrations -- cross-VPC and cross-account traffic without provisioning a load balancer in every VPC.

**Managed EBS volumes close a two-kind contract** -- `volumeConfiguration.name` must match a `configureAtLaunch` volume in the task definition; ECS then attaches a fresh EBS volume per task, tagged via `tagSpecifications` (without them the volumes carry no cost-allocation tags) and optionally hydrated from a snapshot at a chosen `volumeInitializationRate`.

**Autoscaling scales the count** -- target-tracking on CPU, memory, and/or requests-per-target (the AwsAlb / AwsLbTargetGroup `arn_suffix` outputs scope the request metric). The desired count only seeds the initial size once autoscaling owns the service.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEcsCluster** | `clusterArn` | `status.outputs.cluster_arn` |
| **AwsEcsTaskDefinition** | `taskDefinition` | `status.outputs.task_definition_arn` |
| **AwsSubnet** | `network.subnets[]` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `network.securityGroups[]` | `status.outputs.security_group_id` |
| **AwsLbTargetGroup** | `loadBalancers[].targetGroupArn` | `status.outputs.target_group_arn` |
| **AwsLbListenerRule** | `loadBalancers[].advancedConfiguration.*ListenerRule` | `status.outputs.rule_arn` |
| **AwsCloudwatchAlarm** | `alarms.alarmNames[]` | `status.outputs.alarm_name` |
| **AwsKmsKey** | `serviceConnect.services[].tls.kmsKey` | `status.outputs.key_arn` |
| **AwsIamRole** | `volumeConfiguration.managedEbsVolume.roleArn`, `vpcLatticeConfigurations[].roleArn` | `status.outputs.role_arn` |
| **AwsAlb** | `autoscaling.requestsPerTarget.loadBalancerArnSuffix` | `status.outputs.arn_suffix` |
| **AwsLbTargetGroup** | `autoscaling.requestsPerTarget.targetGroupArnSuffix` | `status.outputs.arn_suffix` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_arn` | Amazon Resource Name of the service | Auditing, event rules, and support tooling |
| `service_name` | The service's name | Cross-referencing with `aws ecs describe-services` |
| `cluster_arn` | The cluster the service runs in | Deployment verification |
| `task_definition_arn` | The task-definition revision the service deployed | Recording exactly which revision is live |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Web service behind an ALB** -- Fargate tasks in private subnets registering into a target group the listener routes to, circuit breaker on. The bread-and-butter shape. Start from the **Web Service** preset.

**Background worker** -- no load-balancer wiring at all; the service just keeps N copies of a queue consumer running, scaled by CPU.

**Cost-blended fleet** -- a FARGATE base with FARGATE_SPOT weight for scale-out capacity — the on-demand floor covers interruptions. Start from the **Cost-Optimized Spot Blend** preset.

**Blue/green with canary shifting** -- the BLUE_GREEN strategy with an alternate target group, staged traffic shifting, bake time, and alarm gating for the most sensitive services. Start from the **Blue/Green with Canary** preset.

**Stateful task storage** -- managed EBS volumes attached per task, closing the `configureAtLaunch` contract with the task definition. Start from the **Stateful Service with Managed EBS** preset.

## Works With

- [**AWS ECS Task Definition**](/infra-catalog/aws-ecs-task-definition) -- WHAT runs, referenced by `taskDefinition`; new revisions roll the service.
- [**AWS ECS Cluster**](/infra-catalog/aws-ecs-cluster) -- WHERE it runs, referenced by `clusterArn`.
- [**AWS LB Target Group**](/infra-catalog/aws-lb-target-group) -- the service registers task IPs into it; its `arn_suffix` also scopes request-based autoscaling.
- [**AWS LB Listener**](/infra-catalog/aws-lb-listener) and [**AWS LB Listener Rule**](/infra-catalog/aws-lb-listener-rule) -- the routing graph that delivers traffic to the target group.
- [**AWS Subnet**](/infra-catalog/aws-subnet) and [**AWS Security Group**](/infra-catalog/aws-security-group) -- the task ENIs' placement and firewall.
- [**AWS CloudWatch Alarm**](/infra-catalog/aws-cloudwatch-alarm) -- gates deployments via `alarms.alarmNames`.
- [**AWS ALB**](/infra-catalog/aws-alb) -- its `arn_suffix` output scopes request-based autoscaling.
