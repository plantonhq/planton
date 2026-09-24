# AwsEksNodeGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsEksNodeGroupSpec defines a MANAGED EKS node group: an EC2 fleet that
AWS provisions, health-checks, and rolls for you, registered as workers
of an AwsEksCluster.

The node group composes onto its neighbors instead of embedding them:
the cluster attaches by reference (status.outputs.name), the node IAM
role is a referenced AwsIamRole that carries its own worker policies
(AmazonEKSWorkerNodePolicy, AmazonEC2ContainerRegistryReadOnly,
AmazonEKS_CNI_Policy), and launch mechanics can come from a referenced
AwsLaunchTemplate -- custom AMI, IMDSv2 posture, encrypted volumes,
extra tags -- rather than from the flat inline knobs.

Two configuration styles, honestly separated:
- INLINE: instance_types + disk_size_gb (+ optional remote_access).
  Simple fleets; AWS builds the launch mechanics.
- LAUNCH TEMPLATE: launch_template references an AwsLaunchTemplate.
  AWS then forbids disk_size_gb, instance_types, and remote_access here
  (they live in the template); those conflicts are enforced below.

The node group name comes from metadata.name (AWS limit: 63 characters).
Most instance-shaping fields are create-only in AWS (they replace the
group); labels/taints/scaling/update/repair settings update in place, and
version changes roll the group node by node.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEksNodeGroup
metadata:
  name: awseksnodegroup-demo
spec:
  region: us-west-2
  clusterName:
    value: awsekscluster-demo
  nodeRoleArn:
    value: arn:aws:iam::123456789012:role/EksNodeRole
  subnetIds:
    - value: subnet-0123456789abcdef0
    - value: subnet-0123456789abcdef1
  instanceTypes:
    - t3.medium
  amiType: AL2023_x86_64_STANDARD
  scaling:
    minSize: 1
    maxSize: 3
    desiredSize: 2
  diskSizeGb: 100
  labels:
    pool: demo
  updateConfig:
    maxUnavailable: 1
    updateStrategy: MINIMAL
  # Pre-initialized nodes that cut scale-out from minutes to seconds.
  # STOPPED pooled instances cost only their EBS storage.
  warmPoolConfig:
    poolState: STOPPED
    minSize: 1
    maxGroupPreparedCapacity: 2
    reuseOnScaleIn: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.clusterName` | `string \| valueFrom` | yes |  | AwsEksCluster (`status.outputs.name`) |
| `spec.nodeRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.launchTemplate` | `AwsEksNodeGroupLaunchTemplate` |  |  |  |
| `spec.launchTemplate.launchTemplateId` | `string \| valueFrom` | yes |  | AwsLaunchTemplate (`status.outputs.launch_template_id`) |
| `spec.launchTemplate.version` | `string` |  |  |  |
| `spec.instanceTypes` | `[]string` |  |  |  |
| `spec.amiType` | `string` |  |  |  |
| `spec.capacityType` | `enum` |  | `on_demand` |  |
| `spec.diskSizeGb` | `int32` |  | `100` |  |
| `spec.scaling` | `AwsEksNodeGroupScalingConfig` | yes |  |  |
| `spec.scaling.minSize` | `int32` |  |  |  |
| `spec.scaling.maxSize` | `int32` |  |  |  |
| `spec.scaling.desiredSize` | `int32` |  |  |  |
| `spec.remoteAccess` | `AwsEksNodeGroupRemoteAccess` |  |  |  |
| `spec.remoteAccess.ec2SshKey` | `string` |  |  |  |
| `spec.remoteAccess.sourceSecurityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.taints` | `[]AwsEksNodeGroupTaint` |  |  |  |
| `spec.taints[].key` | `string` | yes |  |  |
| `spec.taints[].value` | `string` |  |  |  |
| `spec.taints[].effect` | `string` | yes |  |  |
| `spec.updateConfig` | `AwsEksNodeGroupUpdateConfig` |  |  |  |
| `spec.updateConfig.maxUnavailable` | `int32` |  |  |  |
| `spec.updateConfig.maxUnavailablePercentage` | `int32` |  |  |  |
| `spec.updateConfig.updateStrategy` | `string` |  |  |  |
| `spec.nodeRepairConfig` | `AwsEksNodeGroupNodeRepairConfig` |  |  |  |
| `spec.nodeRepairConfig.enabled` | `bool` |  |  |  |
| `spec.nodeRepairConfig.maxParallelNodesRepairedCount` | `int32` |  |  |  |
| `spec.nodeRepairConfig.maxParallelNodesRepairedPercentage` | `int32` |  |  |  |
| `spec.nodeRepairConfig.maxUnhealthyNodeThresholdCount` | `int32` |  |  |  |
| `spec.nodeRepairConfig.maxUnhealthyNodeThresholdPercentage` | `int32` |  |  |  |
| `spec.nodeRepairConfig.overrides` | `[]AwsEksNodeGroupNodeRepairOverride` |  |  |  |
| `spec.nodeRepairConfig.overrides[].minRepairWaitTimeMins` | `int32` |  |  |  |
| `spec.nodeRepairConfig.overrides[].nodeMonitoringCondition` | `string` | yes |  |  |
| `spec.nodeRepairConfig.overrides[].nodeUnhealthyReason` | `string` | yes |  |  |
| `spec.nodeRepairConfig.overrides[].repairAction` | `string` | yes |  |  |
| `spec.version` | `string` |  |  |  |
| `spec.releaseVersion` | `string` |  |  |  |
| `spec.forceUpdateVersion` | `bool` |  |  |  |
| `spec.warmPoolConfig` | `AwsEksNodeGroupWarmPoolConfig` |  |  |  |
| `spec.warmPoolConfig.poolState` | `string` |  |  |  |
| `spec.warmPoolConfig.minSize` | `int32` |  |  |  |
| `spec.warmPoolConfig.maxGroupPreparedCapacity` | `int32` |  |  |  |
| `spec.warmPoolConfig.reuseOnScaleIn` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the node group is created in. Must match the cluster's
region. Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.clusterName

`string | valueFrom` · required

The EKS cluster the nodes register with. Reference an AwsEksCluster's
name output or pass a literal cluster name for a cluster managed
outside Planton. Create-only in AWS.

- references: AwsEksCluster (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEksCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.nodeRoleArn

`string | valueFrom` · required

The IAM role every node assumes. It must trust ec2.amazonaws.com and
carry the worker policies (AmazonEKSWorkerNodePolicy,
AmazonEC2ContainerRegistryReadOnly, AmazonEKS_CNI_Policy) -- attach
them on the AwsIamRole itself; this component never modifies a role it
merely references. AmazonEC2ContainerRegistryReadOnly is what lets the
kubelet pull private images from ECR in this account with no pull
secret: ECR issues only twelve-hour tokens, so the node's own identity
(or IRSA on the pod) is the only way a cluster ever pulls from ECR.
Reference an AwsIamRole's role_arn output or pass a literal ARN.
Create-only in AWS.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.subnetIds

`[]string | valueFrom` · required

The subnets nodes launch into -- typically the cluster VPC's private
subnets. One subnet is a legitimate zonal topology (e.g. a stateful
pool pinned to its EBS volumes' zone); use two-plus zones for fleets
that should survive a zone impairment. Reference AwsSubnet subnet_id
outputs or pass literal subnet IDs. Create-only in AWS. A node group
is a member of its cluster and lives there on a diagram; the subnets
are where its nodes attach, so the reference is access, not
placement -- otherwise a group on subnets its cluster does not name
would be drawn outside the cluster it belongs to.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.launchTemplate

`AwsEksNodeGroupLaunchTemplate`

Launch the nodes from an AwsLaunchTemplate instead of the inline
knobs: custom AMI + bootstrap user data, IMDSv2 enforcement, encrypted
or provisioned-IOPS volumes, extra ENI/tag configuration. When set,
AWS forbids disk_size_gb, instance_types, and remote_access on the
node group (enforced below); ami_type stays valid unless the template
pins a custom AMI.

### spec.launchTemplate.launchTemplateId

`string | valueFrom` · required

The launch template. Reference an AwsLaunchTemplate's
launch_template_id output or pass a literal template ID ("lt-...").

- references: AwsLaunchTemplate (`status.outputs.launch_template_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLaunchTemplate, name: <that resource's name>, fieldPath: status.outputs.launch_template_id}} -- a bare string does not parse

### spec.launchTemplate.version

`string`

Which template version the nodes launch from: a numeric version for a
hard pin, "$Default", or "$Latest". Empty keeps the template's default
version. Changing it rolls the group onto the new version -- the
template-driven fleet-rollout mechanism.

### spec.instanceTypes

`[]string`

The EC2 instance types AWS may launch (e.g. ["m6i.large"]). Several
types is a Spot best practice (pool diversity); On-Demand groups use
the first type. Empty keeps the AWS default (t3.medium). Mutually
exclusive with launch_template. Create-only in AWS.

### spec.amiType

`string`

The EKS-optimized AMI family: "AL2023_x86_64_STANDARD" /
"AL2023_ARM_64_STANDARD" (current default generation),
"AL2023_x86_64_NVIDIA" / "AL2023_ARM_64_NVIDIA" /
"AL2023_x86_64_NEURON" (accelerated), the Bottlerocket families
(container-optimized, incl. FIPS/NVIDIA variants), the Windows Core/
Full families (2019/2022/2025), legacy AL2 ("AL2_x86_64",
"AL2_x86_64_GPU", "AL2_ARM_64"), or "CUSTOM" (launch template with
your own AMI). Empty lets AWS pick from the instance types. With a
launch template that pins a custom AMI, leave this empty. Create-only
in AWS.

### spec.capacityType

`enum`

Purchase model for the fleet: on_demand (default), spot (interruptible
at steep discount -- pair with several instance_types), or
capacity_block (pre-purchased ML capacity reservations). Create-only
in AWS.

- default: `on_demand`

Allowed values (use exactly as shown):

- `on_demand` -- Standard On-Demand instances (default).
- `spot` -- Spot instances -- interruptible spare capacity at a steep discount.
- `capacity_block` -- Pre-purchased EC2 Capacity Blocks for ML workloads.

### spec.diskSizeGb

`int32`

Root EBS volume size per node, in GiB. 0 keeps the AWS default (20
GiB Linux / 50 GiB Windows); 100 is a comfortable production default
for image-heavy workloads. Mutually exclusive with launch_template
(size the template's block device instead). Create-only in AWS.

- default: `100`

### spec.scaling

`AwsEksNodeGroupScalingConfig` · required

Node counts the group scales between. desired_size is where the group
starts (and what AWS holds until an autoscaler moves it); min 0 with
desired 0 expresses an intentionally dormant pool.

- rule: {"required":true}
- rule: max_size must be greater than or equal to min_size
- rule: desired_size must be between min_size and max_size

### spec.scaling.minSize

`int32`

The floor the group never shrinks below. 0 is valid -- a pool that
scales to zero when idle.

- rule: {"int32":{"gte":0}}

### spec.scaling.maxSize

`int32`

The ceiling the group never grows above. At least 1.

- rule: {"int32":{"gte":1}}

### spec.scaling.desiredSize

`int32`

The node count AWS creates and maintains (until a cluster autoscaler
adjusts it). Must sit within [min_size, max_size].

- rule: {"int32":{"gte":0}}

### spec.remoteAccess

`AwsEksNodeGroupRemoteAccess`

SSH/security-group access to the nodes. Immutable in AWS -- changing
it replaces the group. Mutually exclusive with launch_template (put
key and security groups in the template instead).

### spec.remoteAccess.ec2SshKey

`string`

The name of an existing EC2 key pair enabling SSH to the nodes.

- rule: {"string":{"maxLen":"255"}}

### spec.remoteAccess.sourceSecurityGroupIds

`[]string | valueFrom`

Security groups allowed to reach the nodes over SSH. Empty with an
ec2_ssh_key set means AWS opens port 22 to 0.0.0.0/0 -- always scope
this when enabling SSH. Reference AwsSecurityGroup security_group_id
outputs or pass literal IDs.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Kubernetes labels applied to every node (visible to schedulers and
nodeSelectors). Updates in place.

- rule: {"map":{"keys":{"string":{"maxLen":"63"}},"values":{"string":{"maxLen":"63"}}}}

### spec.taints

`[]AwsEksNodeGroupTaint`

Kubernetes taints applied to every node -- the reservation mechanism
that keeps ordinary pods off dedicated capacity (GPU pools, ingress
tiers) until a pod tolerates the taint. Updates in place.

- rule: {"repeated":{"maxItems":"50"}}
- rule: effect must be 'NO_SCHEDULE', 'NO_EXECUTE', or 'PREFER_NO_SCHEDULE'

### spec.taints[].key

`string` · required

Taint key (e.g. "dedicated"). Required.

- rule: {"required":true,"string":{"maxLen":"63"}}

### spec.taints[].value

`string`

Taint value (e.g. "gpu"). Optional -- key-only taints are valid.

- rule: {"string":{"maxLen":"63"}}

### spec.taints[].effect

`string` · required

The scheduling effect: "NO_SCHEDULE" (new pods need a toleration),
"PREFER_NO_SCHEDULE" (soft), or "NO_EXECUTE" (also evicts running
pods without a toleration). Required.

- rule: {"required":true}

### spec.updateConfig

`AwsEksNodeGroupUpdateConfig`

How aggressively version updates roll nodes (surge/unavailability
budget). Unset keeps AWS defaults (1 node unavailable, DEFAULT
strategy).

- rule: set exactly one of max_unavailable or max_unavailable_percentage
- rule: max_unavailable must be between 1 and 100 when set
- rule: max_unavailable_percentage must be between 1 and 100 when set
- rule: update_strategy must be 'DEFAULT' or 'MINIMAL' when set

### spec.updateConfig.maxUnavailable

`int32`

Maximum number of nodes updated (unavailable) at once, 1-100.
Exactly one of max_unavailable or max_unavailable_percentage must be
set.

### spec.updateConfig.maxUnavailablePercentage

`int32`

Maximum percentage of nodes updated at once, 1-100. Exactly one of
max_unavailable or max_unavailable_percentage must be set.

### spec.updateConfig.updateStrategy

`string`

The rollout strategy: "DEFAULT" (respect the unavailability budget
only) or "MINIMAL" (additionally launch replacements before
terminating -- the surge, capacity-safe rollout). Empty keeps the AWS
default (DEFAULT).

### spec.nodeRepairConfig

`AwsEksNodeGroupNodeRepairConfig`

Automatic replacement/reboot of nodes the cluster reports unhealthy --
AWS's managed node auto-repair.

- rule: max_parallel_nodes_repaired_count and max_parallel_nodes_repaired_percentage are mutually exclusive
- rule: max_unhealthy_node_threshold_count and max_unhealthy_node_threshold_percentage are mutually exclusive
- rule: max_parallel_nodes_repaired_percentage must be between 1 and 100 when set
- rule: max_unhealthy_node_threshold_percentage must be between 1 and 100 when set

### spec.nodeRepairConfig.enabled

`bool`

Turn node auto-repair on.

### spec.nodeRepairConfig.maxParallelNodesRepairedCount

`int32`

Maximum number of nodes repaired in parallel. Mutually exclusive with
max_parallel_nodes_repaired_percentage. 0 keeps the AWS default.

- rule: {"int32":{"gte":0}}

### spec.nodeRepairConfig.maxParallelNodesRepairedPercentage

`int32`

Maximum percentage of nodes repaired in parallel, 1-100. Mutually
exclusive with max_parallel_nodes_repaired_count.

### spec.nodeRepairConfig.maxUnhealthyNodeThresholdCount

`int32`

Repair pauses when more than this many nodes are unhealthy (a signal
the problem is systemic, not per-node). Mutually exclusive with
max_unhealthy_node_threshold_percentage. 0 keeps the AWS default.

- rule: {"int32":{"gte":0}}

### spec.nodeRepairConfig.maxUnhealthyNodeThresholdPercentage

`int32`

Percentage form of the unhealthy-node pause threshold, 1-100.
Mutually exclusive with max_unhealthy_node_threshold_count.

### spec.nodeRepairConfig.overrides

`[]AwsEksNodeGroupNodeRepairOverride`

Per-condition overrides of the repair action and wait time.

- rule: repair_action must be 'Replace', 'Reboot', or 'NoAction'

### spec.nodeRepairConfig.overrides[].minRepairWaitTimeMins

`int32`

Minutes to wait after the condition is observed before repairing.
Required, at least 1.

- rule: {"int32":{"gte":1}}

### spec.nodeRepairConfig.overrides[].nodeMonitoringCondition

`string` · required

The node monitoring condition this override applies to (e.g.
"AcceleratedHardwareReady"). Required.

- rule: {"required":true}

### spec.nodeRepairConfig.overrides[].nodeUnhealthyReason

`string` · required

The unhealthy reason this override applies to. Required.

- rule: {"required":true}

### spec.nodeRepairConfig.overrides[].repairAction

`string` · required

What to do: "Replace", "Reboot", or "NoAction". Required.

- rule: {"required":true}

### spec.version

`string`

The Kubernetes version of the nodes, e.g. "1.31". Empty follows the
cluster's version at creation. Set it to pin nodes during a control-
plane upgrade, then bump to roll them; nodes may trail the control
plane by up to two minors during an upgrade window.

### spec.releaseVersion

`string`

The exact EKS-optimized AMI release to run (e.g.
"1.31.3-20241109"), for byte-identical fleets and controlled AMI
rollouts. Empty keeps the latest release for `version`. Changing it
rolls the group.

### spec.forceUpdateVersion

`bool`

Force a version update even if pods cannot be drained within their
disruption budgets (otherwise the update fails and rolls back). Only
consulted while version/release_version/launch_template change.

### spec.warmPoolConfig

`AwsEksNodeGroupWarmPoolConfig`

A pool of pre-initialized instances that cuts scale-out latency from
minutes to seconds -- worth its cost exactly when boot time (AMI +
bootstrap + image pulls) dominates how fast new capacity serves
pods. Updates in place.

- rule: pool_state must be 'STOPPED', 'RUNNING', or 'HIBERNATED' when set

### spec.warmPoolConfig.poolState

`string`

The state pooled instances wait in: "STOPPED" (AWS default --
near-zero compute cost, seconds to start), "RUNNING" (instant but
full price), or "HIBERNATED" (RAM restored from disk -- fast
JVM/cache warmup without running cost). Empty keeps the AWS default.

### spec.warmPoolConfig.minSize

`int32`

Minimum number of instances always kept in the pool.

- rule: {"int32":{"gte":0}}

### spec.warmPoolConfig.maxGroupPreparedCapacity

`int32` · optional (explicit presence)

Ceiling on pool size. Unset keeps the AWS default: the gap between
the group's max_size and its desired capacity (AWS represents that
default internally as -1 -- it never appears here). Explicit 0 is
meaningful (no prepared capacity beyond min_size), which is why this
field is optional.

- rule: {"int32":{"gte":0}}

### spec.warmPoolConfig.reuseOnScaleIn

`bool`

Return scaled-in instances to the pool instead of terminating them
-- reuse the warm boot instead of paying for it again.

## Validation Rules

- `launch_template_excludes_instance_types`: instance_types cannot be set with launch_template; define instance type(s) in the launch template
- `launch_template_excludes_disk_size`: disk_size_gb cannot be set with launch_template; size the template's block device instead
- `launch_template_excludes_remote_access`: remote_access cannot be set with launch_template; configure key and security groups in the template
- `version_format`: version must be a Kubernetes minor version of 1.24 or later, e.g. '1.31'
- `ami_type_valid`: ami_type must be a supported EKS AMI family (AL2023_*, BOTTLEROCKET_*, WINDOWS_*, AL2_*, or CUSTOM)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEksNodeGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.nodegroup_name` | `string` | nodegroup_name is the name of the managed node group. |
| `status.outputs.nodegroup_arn` | `string` | nodegroup_arn is the Amazon Resource Name of the node group -- the identifier EKS access entries and IAM policies reference. |
| `status.outputs.asg_name` | `string` | asg_name is the name of the EC2 Auto Scaling group AWS manages behind the node group -- the hook for ASG-level tooling (activity history, suspended processes, custom CloudWatch metrics). |
| `status.outputs.remote_access_sg_id` | `string` | remote_access_sg_id is the ID of the security group AWS creates for SSH access when remote_access is enabled without explicit source security groups. Empty when remote access is off or scoped to provided groups. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.clusterName` | AwsEksCluster | `status.outputs.name` |
| `spec.nodeRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.launchTemplate.launchTemplateId` | AwsLaunchTemplate | `status.outputs.launch_template_id` |
| `spec.remoteAccess.sourceSecurityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |

## See Also

- [Overview](../README.md)
