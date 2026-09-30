# AwsEcsCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsEcsClusterSpec defines an ECS cluster: the logical boundary that
groups services and tasks, decides where their containers run (Fargate,
EC2 capacity providers, or a blend), and carries cluster-wide posture --
Container Insights observability, ECS Exec auditing, Fargate storage
encryption, and the Service Connect default namespace.

Capacity is a spectrum. A serverless cluster associates the AWS-managed
FARGATE / FARGATE_SPOT providers and never thinks about instances. An
EC2-backed cluster defines ec2_capacity_providers -- each wrapping a
referenced AwsAutoScalingGroup whose fleet ECS scales up and down through
managed scaling. A managed-instances cluster defines
managed_instances_capacity_providers -- ECS launches and retires the EC2
instances itself from attribute-based requirements, no auto-scaling group
to own. Services blend across all of it by provider name in their
capacity_provider_strategy. The cluster itself is free; only the tasks
and instances it schedules cost money.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcsCluster
metadata:
  name: ecs-cluster-demo
spec:
  region: us-west-2
  containerInsights: enhanced
  capacityProviders:
    - FARGATE
    - FARGATE_SPOT
  # ECS Managed Instances: ECS launches and retires the EC2 instances
  # itself from these attribute-based requirements -- no auto-scaling
  # group, AMI, or user data to manage.
  managedInstancesCapacityProviders:
    - name: mi-general
      infrastructureRoleArn:
        value: arn:aws:iam::123456789012:role/ecsMiInfrastructureRole
      instanceLaunchTemplate:
        ec2InstanceProfileArn:
          value: arn:aws:iam::123456789012:instance-profile/ecsMiInstanceProfile
        networkConfiguration:
          subnets:
            - value: subnet-0a1b2c3d4e5f60789
            - value: subnet-0f9e8d7c6b5a43210
          # Required: CreateCapacityProvider rejects a managed-instances
          # network configuration without security groups (no VPC-default
          # fall-back on this path).
          securityGroups:
            - value: sg-0a1b2c3d4e5f60001
        instanceRequirements:
          memoryMib:
            min: 2048
          vcpuCount:
            min: 1
        monitoring: BASIC
  defaultCapacityProviderStrategy:
    - capacityProvider: FARGATE
      base: 1
      weight: 1
    - capacityProvider: FARGATE_SPOT
      weight: 4
  executeCommandConfiguration:
    logging: DEFAULT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.containerInsights` | `string` |  |  |  |
| `spec.capacityProviders` | `[]string` |  |  |  |
| `spec.ec2CapacityProviders` | `[]AwsEcsClusterEc2CapacityProvider` |  |  |  |
| `spec.ec2CapacityProviders[].name` | `string` | yes |  |  |
| `spec.ec2CapacityProviders[].autoScalingGroupArn` | `string \| valueFrom` | yes |  | AwsAutoScalingGroup (`status.outputs.autoscaling_group_arn`) |
| `spec.ec2CapacityProviders[].managedScaling` | `AwsEcsClusterManagedScaling` |  |  |  |
| `spec.ec2CapacityProviders[].managedScaling.status` | `string` |  |  |  |
| `spec.ec2CapacityProviders[].managedScaling.targetCapacity` | `int32` |  |  |  |
| `spec.ec2CapacityProviders[].managedScaling.minimumScalingStepSize` | `int32` |  |  |  |
| `spec.ec2CapacityProviders[].managedScaling.maximumScalingStepSize` | `int32` |  |  |  |
| `spec.ec2CapacityProviders[].managedScaling.instanceWarmupPeriodSeconds` | `int32` |  |  |  |
| `spec.ec2CapacityProviders[].managedTerminationProtection` | `string` |  |  |  |
| `spec.ec2CapacityProviders[].managedDraining` | `string` |  |  |  |
| `spec.defaultCapacityProviderStrategy` | `[]AwsEcsClusterCapacityProviderStrategy` |  |  |  |
| `spec.defaultCapacityProviderStrategy[].capacityProvider` | `string` | yes |  |  |
| `spec.defaultCapacityProviderStrategy[].base` | `int32` |  |  |  |
| `spec.defaultCapacityProviderStrategy[].weight` | `int32` |  |  |  |
| `spec.executeCommandConfiguration` | `AwsEcsClusterExecuteCommandConfiguration` |  |  |  |
| `spec.executeCommandConfiguration.logging` | `string` |  |  |  |
| `spec.executeCommandConfiguration.logConfiguration` | `AwsEcsClusterExecuteCommandLogConfiguration` |  |  |  |
| `spec.executeCommandConfiguration.logConfiguration.cloudWatchLogGroupName` | `string` |  |  |  |
| `spec.executeCommandConfiguration.logConfiguration.cloudWatchEncryptionEnabled` | `bool` |  |  |  |
| `spec.executeCommandConfiguration.logConfiguration.s3BucketName` | `string` |  |  |  |
| `spec.executeCommandConfiguration.logConfiguration.s3KeyPrefix` | `string` |  |  |  |
| `spec.executeCommandConfiguration.logConfiguration.s3BucketEncryptionEnabled` | `bool` |  |  |  |
| `spec.executeCommandConfiguration.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.managedStorageConfiguration` | `AwsEcsClusterManagedStorageConfiguration` |  |  |  |
| `spec.managedStorageConfiguration.fargateEphemeralStorageKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.managedStorageConfiguration.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.serviceConnectNamespaceArn` | `string` |  |  |  |
| `spec.managedInstancesCapacityProviders` | `[]AwsEcsClusterManagedInstancesCapacityProvider` |  |  |  |
| `spec.managedInstancesCapacityProviders[].name` | `string` | yes |  |  |
| `spec.managedInstancesCapacityProviders[].infrastructureRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate` | `AwsEcsClusterManagedInstancesLaunchTemplate` | yes |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.ec2InstanceProfileArn` | `string \| valueFrom` | yes |  | AwsIamInstanceProfile (`status.outputs.instance_profile_arn`) |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration` | `AwsEcsClusterManagedInstancesNetworkConfiguration` | yes |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration.subnets` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration.securityGroups` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.capacityOptionType` | `string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.capacityReservations` | `AwsEcsClusterManagedInstancesCapacityReservations` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.capacityReservations.reservationPreference` | `string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.capacityReservations.reservationGroupArn` | `string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements` | `AwsEcsClusterManagedInstancesRequirements` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryMib` | `AwsEcsClusterIntRange` | yes |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryMib.min` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryMib.max` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.vcpuCount` | `AwsEcsClusterIntRange` | yes |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.vcpuCount.min` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.vcpuCount.max` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.allowedInstanceTypes` | `[]string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.excludedInstanceTypes` | `[]string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.instanceGenerations` | `[]string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.cpuManufacturers` | `[]string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.bareMetal` | `string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.burstablePerformance` | `string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.requireHibernateSupport` | `bool` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.spotMaxPricePercentageOverLowestPrice` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.maxSpotPriceAsPercentageOfOptimalOnDemandPrice` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.onDemandMaxPricePercentageOverLowestPrice` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.localStorage` | `string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.localStorageTypes` | `[]string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.totalLocalStorageGb` | `AwsEcsClusterDoubleRange` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.totalLocalStorageGb.min` | `double` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.totalLocalStorageGb.max` | `double` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryGibPerVcpu` | `AwsEcsClusterDoubleRange` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryGibPerVcpu.min` | `double` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryGibPerVcpu.max` | `double` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkInterfaceCount` | `AwsEcsClusterIntRange` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkInterfaceCount.min` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkInterfaceCount.max` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkBandwidthGbps` | `AwsEcsClusterDoubleRange` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkBandwidthGbps.min` | `double` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkBandwidthGbps.max` | `double` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.baselineEbsBandwidthMbps` | `AwsEcsClusterIntRange` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.baselineEbsBandwidthMbps.min` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.baselineEbsBandwidthMbps.max` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorCount` | `AwsEcsClusterIntRange` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorCount.min` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorCount.max` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorManufacturers` | `[]string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorNames` | `[]string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorTypes` | `[]string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorTotalMemoryMib` | `AwsEcsClusterIntRange` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorTotalMemoryMib.min` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorTotalMemoryMib.max` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.useLocalStorage` | `bool` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.monitoring` | `string` |  |  |  |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.storageSizeGib` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].scaleInAfterSeconds` | `int32` |  |  |  |
| `spec.managedInstancesCapacityProviders[].propagateTags` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.containerInsights

`string`

CloudWatch Container Insights for the cluster:
"enabled" -- metrics and logs at the cluster/service/task level.
"enhanced" -- adds container-level observability with automatic
  dashboards (recommended for production; higher CloudWatch cost).
"disabled" -- no Insights telemetry.
Unset keeps the account's default setting. Updatable in place.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["enabled","enhanced","disabled"]}}

### spec.capacityProviders

`[]string`

The AWS-managed serverless capacity providers to associate:
"FARGATE" and/or "FARGATE_SPOT". These are built into every account
-- associating them here is what lets services in this cluster name
them in a capacity_provider_strategy. EC2 capacity is defined
separately in ec2_capacity_providers; both sets associate onto the
cluster together.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["FARGATE","FARGATE_SPOT"]}}}}

### spec.ec2CapacityProviders

`[]AwsEcsClusterEc2CapacityProvider`

EC2 capacity providers, each wrapping a referenced auto-scaling
group. ECS's managed scaling drives the group's desired count from
task demand -- you size the ASG's bounds, ECS turns instances on and
off. Each entry materializes as its own capacity provider resource
(keyed by name, so adding or removing one never disturbs the others)
and is automatically associated with the cluster alongside the
Fargate built-ins. Services reference entries by name in their
capacity_provider_strategy.

- rule: capacity provider names may not start with 'aws', 'ecs', or 'fargate' (reserved by AWS)
- rule: managed_termination_protection must be 'ENABLED' or 'DISABLED' when set
- rule: managed_draining must be 'ENABLED' or 'DISABLED' when set

### spec.ec2CapacityProviders[].name

`string` · required

The capacity provider name -- what services put in their
capacity_provider_strategy. 1-255 characters: letters, digits,
hyphens, underscores; must not start with "aws", "ecs", or "fargate"
(AWS reserves those prefixes).

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9_-]{1,255}$"}}

### spec.ec2CapacityProviders[].autoScalingGroupArn

`string | valueFrom` · required

The auto-scaling group that provides the instances. Reference an
AwsAutoScalingGroup's autoscaling_group_arn output or pass a literal
ARN. The group's launch template decides the instance shape (use an
ECS-optimized AMI whose agent joins this cluster via user data);
ECS's managed scaling then drives the group's desired capacity
between the group's own min/max bounds. ForceNew: changing the group
replaces the provider.

- references: AwsAutoScalingGroup (`status.outputs.autoscaling_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsAutoScalingGroup, name: <that resource's name>, fieldPath: status.outputs.autoscaling_group_arn}} -- a bare string does not parse

### spec.ec2CapacityProviders[].managedScaling

`AwsEcsClusterManagedScaling`

ECS-managed scaling of the auto-scaling group. Leave unset to keep
AWS's defaults (managed scaling enabled, target capacity 100); set
it to tune headroom and scaling step bounds.

- rule: status must be 'ENABLED' or 'DISABLED' when set
- rule: target_capacity must be between 1 and 100 when set
- rule: minimum_scaling_step_size must be between 1 and 10000 when set
- rule: maximum_scaling_step_size must be between 1 and 10000 when set
- rule: instance_warmup_period_seconds must be between 0 and 10000
- rule: maximum_scaling_step_size must be greater than or equal to minimum_scaling_step_size when both are set

### spec.ec2CapacityProviders[].managedScaling.status

`string`

Managed scaling on or off: "ENABLED" (AWS default -- ECS sizes the
group) or "DISABLED" (you size the group yourself; ECS only places
tasks on what exists).

### spec.ec2CapacityProviders[].managedScaling.targetCapacity

`int32`

Utilization target for the group, 1-100 percent. 100 (AWS default)
runs instances fully packed; a lower value (e.g. 80) keeps headroom
so new tasks place without waiting for an instance launch.

### spec.ec2CapacityProviders[].managedScaling.minimumScalingStepSize

`int32`

Smallest scale-out step, 1-10000 instances. AWS default: 1.

### spec.ec2CapacityProviders[].managedScaling.maximumScalingStepSize

`int32`

Largest scale-out step, 1-10000 instances. AWS default: 10000.

### spec.ec2CapacityProviders[].managedScaling.instanceWarmupPeriodSeconds

`int32`

Seconds a newly launched instance warms up before counting toward
capacity metrics, 0-10000. AWS default: 300.

### spec.ec2CapacityProviders[].managedTerminationProtection

`string`

Protect instances running non-daemon tasks from scale-in
termination: "ENABLED" or "DISABLED". Enabling it requires the
auto-scaling group itself to enable new-instance scale-in protection
(protect_from_scale_in on the group) -- AWS rejects the provider
otherwise. The safe default for task-dense clusters.

### spec.ec2CapacityProviders[].managedDraining

`string`

Gracefully drain tasks off instances the group is terminating:
"ENABLED" (AWS default) or "DISABLED". Draining is what makes
scale-in and instance refresh invisible to services.

### spec.defaultCapacityProviderStrategy

`[]AwsEcsClusterCapacityProviderStrategy`

The cluster's default capacity provider strategy -- what ECS uses
when a service or run-task does not declare its own strategy. Name
any associated provider: the Fargate built-ins, an
ec2_capacity_providers entry, or a managed_instances_capacity_providers
entry. Example: FARGATE base 1 / weight 1 + FARGATE_SPOT weight 4
keeps one guaranteed On-Demand task and runs ~80% of scaled capacity
on Spot. Known first-apply caveat when naming a managed-instances
entry created in the SAME apply: the strategy PUT can race the
provider's seconds-long provisioning (AWS rejects it with "not in an
ACTIVE state" until it finishes); a re-apply succeeds. Naming
built-ins or EC2 providers has no such window.

### spec.defaultCapacityProviderStrategy[].capacityProvider

`string` · required

The capacity provider: "FARGATE", "FARGATE_SPOT", or the name of an
ec2_capacity_providers entry.

- rule: {"required":true}

### spec.defaultCapacityProviderStrategy[].base

`int32`

Minimum number of tasks guaranteed on this provider before weights
apply. Only one entry of the strategy may set a non-zero base.

- rule: {"int32":{"lte":100000,"gte":0}}

### spec.defaultCapacityProviderStrategy[].weight

`int32`

Relative share of tasks beyond the bases. Example: FARGATE weight 1 +
FARGATE_SPOT weight 4 runs ~80% of scaled tasks on Spot.

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.executeCommandConfiguration

`AwsEcsClusterExecuteCommandConfiguration`

ECS Exec auditing for the cluster: where interactive exec sessions
(`aws ecs execute-command`) are logged and how session traffic is
encrypted. Without this block, exec sessions still work when a
service enables them -- they are simply not centrally audited.

- rule: logging must be 'DEFAULT', 'OVERRIDE', or 'NONE' when set
- rule: logging 'OVERRIDE' requires log_configuration with at least one destination
- rule: log_configuration only applies when logging is 'OVERRIDE'

### spec.executeCommandConfiguration.logging

`string`

Log destination behavior for exec sessions:
"DEFAULT" (AWS default) -- sessions log to the task's own awslogs
  configuration.
"OVERRIDE" -- sessions log to the destinations in log_configuration.
"NONE" -- exec works but sessions are not logged (avoid outside
  sandboxes; unaudited interactive access defeats compliance).

### spec.executeCommandConfiguration.logConfiguration

`AwsEcsClusterExecuteCommandLogConfiguration`

Custom destinations for exec session logs. Only used (and required)
when logging is "OVERRIDE".

- rule: provide cloud_watch_log_group_name and/or s3_bucket_name

### spec.executeCommandConfiguration.logConfiguration.cloudWatchLogGroupName

`string`

The CloudWatch log group session logs are sent to. The group must
already exist -- ECS does not create it.

### spec.executeCommandConfiguration.logConfiguration.cloudWatchEncryptionEnabled

`bool`

Require the CloudWatch log group to be KMS-encrypted; the send fails
if it is not. Pair with an encrypted log group for compliance
postures.

### spec.executeCommandConfiguration.logConfiguration.s3BucketName

`string`

The S3 bucket session logs are written to.

### spec.executeCommandConfiguration.logConfiguration.s3KeyPrefix

`string`

Key prefix for session log objects within the bucket.

### spec.executeCommandConfiguration.logConfiguration.s3BucketEncryptionEnabled

`bool`

Require the S3 bucket to be encrypted; the write fails if it is not.

### spec.executeCommandConfiguration.kmsKeyId

`string | valueFrom`

A KMS key to encrypt the exec session traffic between client and
container. Reference an AwsKmsKey's key_arn output or pass a literal
key ARN/ID. Unset uses TLS without customer-managed encryption.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.managedStorageConfiguration

`AwsEcsClusterManagedStorageConfiguration`

Customer-managed KMS encryption for Fargate ephemeral task storage
and managed storage -- the compliance posture for regulated
workloads. Unset uses AWS-owned keys (data is still encrypted).

### spec.managedStorageConfiguration.fargateEphemeralStorageKmsKeyId

`string | valueFrom`

The KMS key encrypting Fargate ephemeral task storage. Reference an
AwsKmsKey's key_arn output or pass a literal key ARN. The key policy
must grant the Fargate service principal decrypt/generate rights --
AWS rejects the cluster configuration otherwise.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.managedStorageConfiguration.kmsKeyId

`string | valueFrom`

The KMS key for other ECS-managed storage. Reference an AwsKmsKey's
key_arn output or pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.serviceConnectNamespaceArn

`string`

The AWS Cloud Map namespace (by ARN) that Service Connect uses by
default for services in this cluster. Services can override it;
setting it here is what lets a whole environment share one service
mesh namespace without per-service wiring.

### spec.managedInstancesCapacityProviders

`[]AwsEcsClusterManagedInstancesCapacityProvider`

ECS Managed Instances capacity providers: ECS launches, patches, and
retires the EC2 instances itself -- you describe the compute by
attributes (vCPUs, memory, accelerators) and the network to launch
into, and ECS owns the fleet end to end (no auto-scaling group, no
AMI, no user data). Each entry materializes as its own capacity
provider resource that AWS binds to this cluster at creation --
unlike EC2 providers there is no association step
(PutClusterCapacityProviders neither attaches nor detaches
managed-instances providers); services reference entries by name in
their capacity_provider_strategy. Requires an infrastructure role the
ECS service principal can assume and an instance profile for the
launched instances.

- rule: capacity provider names may not start with 'aws', 'ecs', or 'fargate' (reserved by AWS)
- rule: propagate_tags must be 'CAPACITY_PROVIDER' or 'NONE' when set

### spec.managedInstancesCapacityProviders[].name

`string` · required

The capacity provider name -- what services put in their
capacity_provider_strategy. 1-255 characters: letters, digits,
hyphens, underscores; must not start with "aws", "ecs", or "fargate"
(AWS reserves those prefixes).

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9_-]{1,255}$"}}

### spec.managedInstancesCapacityProviders[].infrastructureRoleArn

`string | valueFrom` · required

The infrastructure role ECS assumes to launch, patch, and retire the
managed instances. Reference an AwsIamRole's role_arn output or pass
a literal ARN. The role must trust the ecs.amazonaws.com service
principal and carry AmazonECSInfrastructureRolePolicyForManagedInstances
(or equivalent); the caller applying the manifest needs iam:PassRole
on it.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate

`AwsEcsClusterManagedInstancesLaunchTemplate` · required

What the launched instances look like: instance profile, network
placement, and the attribute-based requirements ECS resolves into
concrete instance types.

- rule: {"required":true}
- rule: capacity_option_type must be 'ON_DEMAND', 'SPOT', or 'RESERVED' when set
- rule: capacity_option_type 'RESERVED' requires capacity_reservations
- rule: capacity_reservations is only legal when capacity_option_type is 'RESERVED'
- rule: reservation_preference 'RESERVATIONS_ONLY' or 'RESERVATIONS_FIRST' requires instance_requirements
- rule: monitoring must be 'BASIC' or 'DETAILED' when set
- rule: storage_size_gib must be at least 1 when set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.ec2InstanceProfileArn

`string | valueFrom` · required

The instance profile attached to every launched instance -- the
instance-side identity (the ECS agent's permissions come from here).
Reference an AwsIamInstanceProfile's instance_profile_arn output or
pass a literal ARN.

- references: AwsIamInstanceProfile (`status.outputs.instance_profile_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamInstanceProfile, name: <that resource's name>, fieldPath: status.outputs.instance_profile_arn}} -- a bare string does not parse

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration

`AwsEcsClusterManagedInstancesNetworkConfiguration` · required

Where the managed instances launch: the subnets (required) and
security groups applied to each instance.

- rule: {"required":true}

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration.subnets

`[]string | valueFrom` · required

Subnets the instances launch into -- span at least two AZs for
availability. Reference AwsSubnet subnet_id outputs or pass literal
subnet IDs.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration.securityGroups

`[]string | valueFrom` · required

Security groups applied to each instance -- at least one is REQUIRED.
Reference AwsSecurityGroup security_group_id outputs or pass literal
group IDs. Unlike EC2 launch paths there is NO fall-back to the VPC
default group: AWS's CreateCapacityProvider rejects a managed-instances
network configuration without security groups (ClientException
"must specify a Network Configuration that contain security groups"),
even though the Terraform provider's schema marks the argument
optional -- the contract lives only server-side.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.capacityOptionType

`string`

Purchase model for the launched capacity: "ON_DEMAND" (AWS default),
"SPOT", or "RESERVED" (draw from capacity reservations --
capacity_reservations must then be set). Changing this replaces the
capacity provider; everything else in the launch template updates in
place.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.capacityReservations

`AwsEcsClusterManagedInstancesCapacityReservations`

Which capacity reservations RESERVED capacity draws from. Only legal
(and required) when capacity_option_type is "RESERVED".

- rule: reservation_preference must be 'RESERVATIONS_ONLY', 'RESERVATIONS_FIRST', or 'RESERVATIONS_EXCLUDED' when set
- rule: reservation_group_arn is only legal when reservation_preference is 'RESERVATIONS_ONLY'

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.capacityReservations.reservationPreference

`string`

How reservations are used:
"RESERVATIONS_ONLY" -- launch only into reservations (pair with
  reservation_group_arn to scope which ones).
"RESERVATIONS_FIRST" -- prefer reservations, overflow to on-demand.
"RESERVATIONS_EXCLUDED" -- never consume reservations.
RESERVATIONS_ONLY and RESERVATIONS_FIRST require instance_requirements
on the launch template.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.capacityReservations.reservationGroupArn

`string`

A capacity-reservation group ARN scoping which reservations to use.
Only legal when reservation_preference is "RESERVATIONS_ONLY".

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements

`AwsEcsClusterManagedInstancesRequirements`

Attribute-based instance requirements -- describe the compute
(memory, vCPUs, accelerators, price protection) and ECS resolves
matching instance types at launch. Required when
capacity_reservations uses a RESERVATIONS_ONLY or RESERVATIONS_FIRST
preference.

- rule: allowed_instance_types and excluded_instance_types are mutually exclusive
- rule: spot_max_price_percentage_over_lowest_price and max_spot_price_as_percentage_of_optimal_on_demand_price are mutually exclusive
- rule: bare_metal must be 'included', 'excluded', or 'required' when set
- rule: burstable_performance must be 'included', 'excluded', or 'required' when set
- rule: local_storage must be 'included', 'excluded', or 'required' when set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryMib

`AwsEcsClusterIntRange` · required

Required. Memory per instance, in MiB. min is required; leave max
unset (0) for no upper bound.

- rule: {"required":true}
- rule: max must be greater than or equal to min when both are set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryMib.min

`int32`

Lower bound, inclusive.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryMib.max

`int32`

Upper bound, inclusive. 0 means no upper bound.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.vcpuCount

`AwsEcsClusterIntRange` · required

Required. vCPUs per instance. min is required; leave max unset (0)
for no upper bound.

- rule: {"required":true}
- rule: max must be greater than or equal to min when both are set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.vcpuCount.min

`int32`

Lower bound, inclusive.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.vcpuCount.max

`int32`

Upper bound, inclusive. 0 means no upper bound.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.allowedInstanceTypes

`[]string`

Allow-list of instance types or families, with wildcards
("m5.large", "m5.*", "c*"). At most 400 entries. Mutually exclusive
with excluded_instance_types.

- rule: {"repeated":{"maxItems":"400"}}

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.excludedInstanceTypes

`[]string`

Deny-list of instance types or families, with wildcards. At most 400
entries. Mutually exclusive with allowed_instance_types.

- rule: {"repeated":{"maxItems":"400"}}

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.instanceGenerations

`[]string`

Instance generations to include: "current" and/or "previous". AWS
default: any generation matching the other requirements.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.cpuManufacturers

`[]string`

CPU manufacturers to include: "intel", "amd", "amazon-web-services"
(Graviton), "apple". AWS default: any.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.bareMetal

`string`

Bare-metal eligibility: "included", "excluded" (AWS default), or
"required".

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.burstablePerformance

`string`

Burstable (T-family) eligibility: "included", "excluded" (AWS
default), or "required".

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.requireHibernateSupport

`bool`

Only instance types that support hibernation.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.spotMaxPricePercentageOverLowestPrice

`int32`

Spot price protection: exclude types whose Spot price exceeds the
identified lowest-priced type's Spot price by more than this
percentage. Mutually exclusive with
max_spot_price_as_percentage_of_optimal_on_demand_price.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.maxSpotPriceAsPercentageOfOptimalOnDemandPrice

`int32`

Spot price protection anchored to On-Demand: exclude types whose
Spot price exceeds this percentage of the optimal type's On-Demand
price. Mutually exclusive with
spot_max_price_percentage_over_lowest_price.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.onDemandMaxPricePercentageOverLowestPrice

`int32`

On-Demand price protection: exclude types whose On-Demand price
exceeds the identified lowest-priced type's by more than this
percentage. AWS default: 20.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.localStorage

`string`

Instance-store (local disk) eligibility: "included" (AWS default),
"excluded", or "required".

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.localStorageTypes

`[]string`

Local storage technologies when instance-store is in play: "hdd"
and/or "ssd".

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.totalLocalStorageGb

`AwsEcsClusterDoubleRange`

Total local (instance-store) storage, in GB.

- rule: max must be greater than or equal to min when both are set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.totalLocalStorageGb.min

`double`

Lower bound, inclusive.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.totalLocalStorageGb.max

`double`

Upper bound, inclusive. 0 means no upper bound.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryGibPerVcpu

`AwsEcsClusterDoubleRange`

Memory-to-vCPU ratio, in GiB per vCPU -- a compact way to say
"memory optimized" (min 8) or "compute optimized" (max 2) without
naming families.

- rule: max must be greater than or equal to min when both are set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryGibPerVcpu.min

`double`

Lower bound, inclusive.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.memoryGibPerVcpu.max

`double`

Upper bound, inclusive. 0 means no upper bound.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkInterfaceCount

`AwsEcsClusterIntRange`

Number of network interfaces the type must support.

- rule: max must be greater than or equal to min when both are set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkInterfaceCount.min

`int32`

Lower bound, inclusive.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkInterfaceCount.max

`int32`

Upper bound, inclusive. 0 means no upper bound.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkBandwidthGbps

`AwsEcsClusterDoubleRange`

Network bandwidth, in Gbps.

- rule: max must be greater than or equal to min when both are set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkBandwidthGbps.min

`double`

Lower bound, inclusive.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.networkBandwidthGbps.max

`double`

Upper bound, inclusive. 0 means no upper bound.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.baselineEbsBandwidthMbps

`AwsEcsClusterIntRange`

Baseline (non-burst) EBS bandwidth, in Mbps.

- rule: max must be greater than or equal to min when both are set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.baselineEbsBandwidthMbps.min

`int32`

Lower bound, inclusive.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.baselineEbsBandwidthMbps.max

`int32`

Upper bound, inclusive. 0 means no upper bound.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorCount

`AwsEcsClusterIntRange`

Number of accelerators (GPUs, FPGAs, inference chips). Set min 1 to
require accelerated types; to EXCLUDE accelerators, leave this unset
and rely on accelerator_types being empty.

- rule: max must be greater than or equal to min when both are set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorCount.min

`int32`

Lower bound, inclusive.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorCount.max

`int32`

Upper bound, inclusive. 0 means no upper bound.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorManufacturers

`[]string`

Accelerator manufacturers: "nvidia", "amd", "amazon-web-services",
"xilinx", "habana".

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorNames

`[]string`

Specific accelerator models (e.g. "a100", "v100", "t4",
"inferentia", "radeon-pro-v520").

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorTypes

`[]string`

Accelerator categories: "gpu", "fpga", "inference".

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorTotalMemoryMib

`AwsEcsClusterIntRange`

Total accelerator memory, in MiB.

- rule: max must be greater than or equal to min when both are set

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorTotalMemoryMib.min

`int32`

Lower bound, inclusive.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.instanceRequirements.acceleratorTotalMemoryMib.max

`int32`

Upper bound, inclusive. 0 means no upper bound.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.useLocalStorage

`bool` · optional (explicit presence)

Use instance-store (local NVMe) volumes for container storage on
instance types that have them. Unset keeps AWS's default placement.

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.monitoring

`string`

CloudWatch monitoring detail for the launched instances: "BASIC"
(AWS default) or "DETAILED" (1-minute metrics, billed).

### spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.storageSizeGib

`int32`

Root EBS volume size for each launched instance, in GiB (>= 1).
Unset keeps AWS's default size.

### spec.managedInstancesCapacityProviders[].scaleInAfterSeconds

`int32` · optional (explicit presence)

Seconds an empty managed instance idles before ECS scales it in,
0-3600; -1 disables scale-in entirely (instances stay until
terminated another way). Unset keeps AWS's default optimization.

- rule: {"int32":{"lte":3600,"gte":-1}}

### spec.managedInstancesCapacityProviders[].propagateTags

`string`

Propagate the capacity provider's tags to the EC2 instances ECS
launches: "CAPACITY_PROVIDER" or "NONE". Unset keeps AWS's default.

## Validation Rules

- `default_strategy_names_associated_providers`: every default_capacity_provider_strategy entry must name an associated provider -- a capacity_providers built-in (FARGATE/FARGATE_SPOT), an ec2_capacity_providers entry, or a managed_instances_capacity_providers entry
- `default_strategy_single_base`: only one default_capacity_provider_strategy entry may set a non-zero base
- `ec2_capacity_provider_names_unique`: ec2_capacity_providers entries must have unique names
- `managed_instances_capacity_provider_names_unique`: managed_instances_capacity_providers entries must have unique names
- `capacity_provider_names_unique_across_types`: a managed_instances_capacity_providers entry may not reuse an ec2_capacity_providers name -- both lists share one provider namespace

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEcsCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_name` | `string` | The cluster name (mirrors metadata.name). What the AWS CLI and the ECS agent's ECS_CLUSTER setting address. |
| `status.outputs.cluster_arn` | `string` | The ARN of the cluster. The join key: an AwsEcsService's cluster_arn references this output. |
| `status.outputs.capacity_provider_names` | `[]string` | Every capacity provider associated with the cluster -- the Fargate built-ins plus the names of the folded EC2 capacity providers. The vocabulary services can use in a capacity_provider_strategy. |
| `status.outputs.capacity_provider_arns` | `[]string` | The ARNs of the EC2 capacity providers this cluster defines (empty for Fargate-only clusters), in the order declared in the spec. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.ec2CapacityProviders[].autoScalingGroupArn` | AwsAutoScalingGroup | `status.outputs.autoscaling_group_arn` |
| `spec.executeCommandConfiguration.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.managedStorageConfiguration.fargateEphemeralStorageKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.managedStorageConfiguration.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.managedInstancesCapacityProviders[].infrastructureRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.ec2InstanceProfileArn` | AwsIamInstanceProfile | `status.outputs.instance_profile_arn` |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration.subnets` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsEcsService | `spec.clusterArn` | `status.outputs.cluster_arn` |
| AwsEventBridgeScheduler | `spec.target.arn` | `status.outputs.cluster_arn` |

## See Also

- [Overview](../README.md)
