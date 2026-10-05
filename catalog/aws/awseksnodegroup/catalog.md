# AWS EKS Node Group

Deploys a managed EKS node group — an EC2 fleet AWS provisions, health-checks, and rolls for you, registered as workers of an existing EKS cluster. The kind supports inline instance configuration or launch-template-driven fleets, On-Demand/Spot/Capacity-Block purchase models, auto-scaling bounds, Kubernetes labels and taints, controlled version rollouts, and managed node auto-repair.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EKS Managed Node Group** -- EC2 worker nodes attached to the referenced EKS cluster, placed in your subnets, assuming the referenced IAM role, with the launch configuration, capacity type, scaling bounds, scheduling metadata, and update/repair policy you declare
- **Auto Scaling Group** -- managed by EKS behind the node group, scaling between `scaling.minSize` and `scaling.maxSize`
- **SSH Security Group** -- only when `remoteAccess.ec2SshKey` is set without source security groups; AWS opens port 22 wide, so always scope SSH when enabling it
- **Warm Pool** -- only when `warmPoolConfig` is set; pre-initialized instances held stopped, running, or hibernated so scale-out joins in seconds
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An existing EKS cluster** to register with. Reference an AwsEksCluster's `name` output via ValueFromRef, or provide a literal cluster name for a cluster managed outside Planton.
- **An IAM role** trusting `ec2.amazonaws.com` with the worker policies attached (`AmazonEKSWorkerNodePolicy`, `AmazonEC2ContainerRegistryReadOnly`, `AmazonEKS_CNI_Policy`) — attach them on the role itself; this component never modifies a role it merely references.
- **At least one subnet** — typically the cluster VPC's private subnets. One subnet is a legitimate zonal topology; use two-plus zones for fleets that should survive a zone impairment.
- **A launch template** (optional) -- for custom AMIs, IMDSv2 enforcement, or encrypted volumes. Reference an AwsLaunchTemplate Infra Component; AWS then forbids the inline instance/disk/SSH knobs.

## Deploy

### Console

Open the deployment store, find **AWS EKS Node Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **On-Demand General Pool** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEksNodeGroup
metadata:
  name: general-workers
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  clusterName:
    value: "platform-cluster"
  nodeRoleArn:
    value: "arn:aws:iam::123456789012:role/EksWorkerNodeRole"
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  instanceTypes: [m6i.large]
  diskSizeGb: 100
  scaling:
    minSize: 2
    maxSize: 5
    desiredSize: 2
```

```shell
planton apply -f eks-node-group.yaml
```

This creates an On-Demand node group with m6i.large instances scaling between 2 and 5 nodes on 100 GiB root disks, no SSH access, no labels or taints. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the node group to an EKS cluster, IAM role, and subnets deployed in the same InfraPipeline:

```yaml
spec:
  clusterName:
    valueFrom:
      kind: AwsEksCluster
      name: platform-cluster
      fieldPath: status.outputs.name
  nodeRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: eks-worker-role
      fieldPath: status.outputs.role_arn
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-b
        fieldPath: status.outputs.subnet_id
```

The InfraPipeline resolves the dependency graph, deploys the subnets, IAM role, and EKS cluster first, then provisions the node group with the resolved values.

## Key Configuration

These are the most important decisions when configuring an EKS node group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Configuration style** -- Inline (`instanceTypes` + `diskSizeGb` + optional `remoteAccess`) for simple fleets, or `launchTemplate` referencing an AwsLaunchTemplate for custom AMIs, IMDSv2 posture, and encrypted volumes. The styles are mutually exclusive: with a template, AWS forbids the inline knobs — the template owns them.

**Instance types** -- The EC2 types AWS may launch. Empty keeps the AWS default (t3.medium). On-Demand groups use the first type; Spot fleets should list several similar types (m6i.large, m5.large, m6a.large) for capacity-pool diversity. Create-only.

**Capacity type** -- `on_demand` (default) for predictable availability, `spot` for fault-tolerant workloads that can absorb a 2-minute reclaim notice in exchange for lower instance cost, or `capacity_block` for pre-purchased ML capacity. Create-only.

**Scaling** -- `minSize` 0 is a valid dormant pool; min ≥ 2 keeps a production fleet alive through a node failure. `desiredSize` is where the group starts and what AWS holds until a cluster autoscaler moves it.

**Scheduling** -- `labels` let nodeSelectors and affinity target the group; `taints` reserve it for workloads that explicitly tolerate them (GPU pools, ingress tiers). Both update in place.

**Version rollouts** -- `version` pins the node Kubernetes minor (empty follows the cluster at creation); `releaseVersion` pins the exact EKS-optimized AMI build; `updateConfig` sets the unavailability budget (node count XOR percentage) and the `MINIMAL` surge strategy for capacity-safe rolls.

**Node auto-repair** -- `nodeRepairConfig` lets EKS replace or reboot unhealthy nodes within parallelism and unhealthy-threshold bounds, with per-condition overrides (e.g. Reboot on GPU XID errors instead of Replace).

**Warm pools** -- `warmPoolConfig` keeps pre-initialized nodes on standby: `poolState` picks the cost/latency trade-off (`STOPPED` bills only EBS; `RUNNING` joins instantly at full price; `HIBERNATED` restores RAM from disk), `minSize` sets the standby floor, `maxGroupPreparedCapacity` caps the pool (omit for AWS's default ceiling, explicit 0 for nothing beyond the floor), and `reuseOnScaleIn` returns scaled-in nodes to the pool. Updates in place.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEksCluster** | `clusterName` | `status.outputs.name` |
| **AwsIamRole** | `nodeRoleArn` | `status.outputs.role_arn` |
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsLaunchTemplate** (optional) | `launchTemplate.launchTemplateId` | `status.outputs.launch_template_id` |
| **AwsSecurityGroup** (optional) | `remoteAccess.sourceSecurityGroupIds` | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `nodegroup_name` | EKS node group name | Monitoring, kubectl node group identification |
| `nodegroup_arn` | Node group Amazon Resource Name | EKS access entries, IAM policies |
| `asg_name` | Auto Scaling Group name | CloudWatch alarms, ASG-level tooling |
| `remote_access_sg_id` | SSH security group ID (when SSH is enabled without explicit source groups) | Finding and tightening the generated rule |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**On-demand general purpose** -- m6i.large instances, 2–5 node scaling, 100 GiB disks, multi-AZ placement. The standard starting point for workloads that need predictable capacity. Start from the **On-Demand General Pool** preset.

**Spot cost-optimized** -- Several similar instance types on Spot, a `node-lifecycle: spot` label, and a taint if only Spot-tolerant workloads should land there. Pair with an On-Demand group for critical services. Start from the **Spot Cost-Optimized Pool** preset.

**Template-driven fleet** -- A `launchTemplate` reference with a pinned version: custom AMI, IMDSv2 required, encrypted volumes. Roll the fleet by publishing a new template version and bumping the pin. Start from the **Launch-Template Pool** preset.

**Warm standby capacity** -- A `warmPoolConfig` with `poolState: STOPPED`, a small standby floor, and `reuseOnScaleIn: true`: scale-out serves traffic in seconds while pooled nodes bill only their EBS storage. Start from the **Warm Pool Node Group** preset.

## Works With

- [**AWS EKS Cluster**](/infra-catalog/aws-eks-cluster) -- provides the Kubernetes control plane this node group registers with
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the worker node role with the EKS and ECR policies
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnets for worker node placement
- [**AWS Launch Template**](/infra-catalog/aws-launch-template) -- provides custom launch mechanics for template-driven fleets
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- scopes SSH access when remote access is enabled
