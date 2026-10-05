# AWS Auto Scaling Group

Deploys an EC2 Auto Scaling group — the fleet manager that keeps a set of instances launched from a launch template at the desired size, replaces unhealthy members, spreads capacity across subnets, and scales on policies and schedules. The group is a pure orchestrator: WHAT launches lives in the referenced AwsLaunchTemplate; this group decides how many, where, and when. Scaling policies, scheduled actions, lifecycle hooks, and notifications are folded into the spec (they are sub-resources of exactly one group), and both IaC modules manage each as its own provider resource so adding or removing one is an in-place update — never a group replacement.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Auto Scaling Group** -- the fleet with its bounds, subnet spread, health model, capacity source (single template or mixed-instances policy), Capacity Reservation targeting, traffic sources (VPC Lattice / Classic ELB), and terminate-hook retention policy
- **Scaling Policies** -- one per `scalingPolicies` entry (target tracking, step, simple, or predictive — with pair, split, or fully customized forecast metrics); target tracking's underlying CloudWatch alarms are created and managed by AWS; `disabled` pauses a policy without deleting it
- **Scheduled Actions** -- one per `scheduledActions` entry: cron-driven or one-shot capacity changes
- **Lifecycle Hooks** -- one per `lifecycleHooks` entry: pause points at launch and termination; hooks flagged `applyAtLaunch` attach atomically at group creation so even the first instance is caught
- **Warm Pool** -- attached only when `warmPool` is configured; pre-initialized instances that cut scale-out latency
- **Notification Configurations** -- attached only when `notifications` names an SNS topic

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A launch template with an AMI** -- the group requires the template it references to carry an image. Reference an AwsLaunchTemplate Infra Component or pass a literal lt- ID.
- **Subnets in at least two availability zones** -- the fault-tolerance floor; the group rebalances across the zones its subnets cover.
- **Target groups** (for load-balanced fleets) -- reference AwsLbTargetGroup resources and pair them with the ELB health check type.

## Deploy

### Console

Open the deployment store, find **AWS Auto Scaling Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Web Service Behind ALB** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAutoScalingGroup
metadata:
  name: web-fleet
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  subnets:
    - valueFrom:
        kind: AwsSubnet
        name: private-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-b
        fieldPath: status.outputs.subnet_id
  launchTemplate:
    launchTemplateId:
      valueFrom:
        kind: AwsLaunchTemplate
        name: web-fleet-base
        fieldPath: status.outputs.launch_template_id
  minSize: 2
  maxSize: 10
  healthCheckType: ELB
  targetGroups:
    - valueFrom:
        kind: AwsLbTargetGroup
        name: web-tg
        fieldPath: status.outputs.target_group_arn
  instanceRefresh:
    strategy: Rolling
    preferences:
      minHealthyPercentage: 100
      maxHealthyPercentage: 110
      autoRollback: true
  scalingPolicies:
    - name: cpu-target
      policyType: TargetTrackingScaling
      targetTracking:
        targetValue: 60
        predefinedMetricType: ASGAverageCPUUtilization
```

```shell
planton apply -f auto-scaling-group.yaml
```

This runs the full production loop: a two-instance floor across two zones, ELB health checks replacing wedged processes, a CPU thermostat at 60%, and rolling surge rollouts (110%/100%) with auto-rollback whenever the launch template publishes a new version. An Infra Job tracks the provisioning in real time.

### InfraChart

When the fleet deploys alongside its network and launch template in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  subnets:
    - valueFrom:
        kind: AwsSubnet
        name: private-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-b
        fieldPath: status.outputs.subnet_id
  launchTemplate:
    launchTemplateId:
      valueFrom:
        kind: AwsLaunchTemplate
        name: web-fleet-base
        fieldPath: status.outputs.launch_template_id
  minSize: 2
  maxSize: 10
```

The InfraPipeline resolves the dependency graph, deploys the subnets and launch template first, then provisions the fleet against their outputs.

## Key Configuration

These are the most important decisions when configuring an auto-scaling group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Single template XOR mixed instances** -- Exactly one capacity source: `launchTemplate` for the common one-type fleet, or `mixedInstancesPolicy` blending instance types and purchase options — an On-Demand base for guaranteed capacity, a Spot majority above it, drawn from several pools so no single interruption hurts. An explicit `onDemandPercentageAboveBaseCapacity: 0` means all-Spot above the base.

**Bounds, not a headcount** -- `minSize` is the availability floor, `maxSize` the cost ceiling. Leave `desiredCapacity` unset (0): the fleet starts at the minimum and scaling policies govern — a pinned count is re-asserted on every apply and fights the autoscaler.

**ELB health checks whenever target groups exist** -- EC2 status checks only see a dead machine; the target group's health check sees a dead process. Registering `targetGroups` without `healthCheckType: ELB` is the classic silent failure — AWS accepts it, and a wedged process serves errors forever.

**Instance refresh is the rollout mechanism** -- With `instanceRefresh.strategy: Rolling`, a launch-template version change rolls the fleet in health-bounded waves. `minHealthyPercentage: 100` + `maxHealthyPercentage: 110` gives launch-before-terminate (the fleet never dips); `autoRollback: true` plus watched alarms makes failed rollouts undo themselves; checkpoint percentages stage a canary.

**Spot safety nets** -- `capacityRebalance: true` replaces at-risk Spot instances before the two-minute notice; the `AllocationStrategy` termination policy keeps scale-in on the fleet's preferred pools.

**Graceful drain** -- `protectFromScaleIn` plus a terminating lifecycle hook lets the workload finish its jobs, release protection, and complete the hook — nothing is lost on scale-in. The modern hook needs no SNS target: EventBridge rules watch lifecycle events.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnets[]` | `status.outputs.subnet_id` |
| **AwsLaunchTemplate** | `launchTemplate.launchTemplateId` | `status.outputs.launch_template_id` |
| **AwsLaunchTemplate** (mixed base / per-override) | `mixedInstancesPolicy.launchTemplate.launchTemplateId`, `mixedInstancesPolicy.overrides[].launchTemplate.launchTemplateId` | `status.outputs.launch_template_id` |
| **AwsLbTargetGroup** | `targetGroups[]` | `status.outputs.target_group_arn` |
| **AwsCloudwatchAlarm** (refresh watch) | `instanceRefresh.preferences.alarms[]` | `status.outputs.alarm_name` |
| **AwsSnsTopic** (hooks + notifications) | `lifecycleHooks[].notificationTargetArn`, `notifications.topic` | `status.outputs.topic_arn` |
| **AwsIamRole** (hook publish role) | `lifecycleHooks[].roleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `autoscaling_group_name` | The group name (metadata.name) | CloudWatch dimensions (AutoScalingGroupName), ECS capacity providers, AWS CLI operations |
| `autoscaling_group_arn` | Amazon Resource Name of the group | IAM policies and EventBridge rules scoped to this group |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Web service behind an ALB** -- Two-zone fleet, ELB health checks, a CPU thermostat, and surge rollouts. Start from the **Web Service Behind ALB** preset.

**Spot mixed fleet** -- An On-Demand base of two with an all-Spot majority across four instance pools, capacity rebalance on. Start from the **Spot Mixed Fleet** preset.

**Scheduled scale** -- Business-hours scale-up and overnight scale-to-zero on cron, with a stopped warm pool for fast mornings. Start from the **Scheduled Scale with Warm Pool** preset.

**Reserved fleet** -- Fill pre-purchased Capacity Reservations first, place capacity reservations-then-balanced, retain failed-drain instances for post-mortems, and attach the warm-up hook atomically at creation. Start from the **Reserved Fleet with Guaranteed Capacity** preset.

## Works With

- [**AWS Launch Template**](/infra-catalog/aws-launch-template) -- the blueprint every instance launches from; publishing a version rolls this fleet via instance refresh
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- the zones capacity spreads across
- [**AWS LB Target Group**](/infra-catalog/aws-lb-target-group) -- the traffic contract; pair with ELB health checks
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) -- fleet lifecycle notifications and lifecycle-hook delivery
- [**AWS CloudWatch Alarm**](/infra-catalog/aws-cloudwatch-alarm) -- step-scaling triggers and instance-refresh watch alarms
