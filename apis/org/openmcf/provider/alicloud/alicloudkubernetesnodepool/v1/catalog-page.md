# AliCloud KubernetesNodePool

Deploys a worker node pool in an Alibaba Cloud ACK Managed Kubernetes cluster with configurable instance types, ESSD disk configuration, auto-scaling, managed lifecycle (auto-repair, auto-upgrade), spot instance support, and Kubernetes scheduling properties (labels, taints).

## What Gets Created

When you deploy an AliCloudKubernetesNodePool resource, OpenMCF provisions:

- **ACK Node Pool** — an `alicloud_cs_kubernetes_node_pool` resource containing a group of ECS worker nodes with shared instance configuration, scaling policy, and Kubernetes properties
- **ECS Instances** — worker nodes provisioned within the pool based on `desiredSize` or auto-scaler decisions
- **Auto Scaling Group** — backing scaling group for the node pool, used for auto-scaling operations

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **An existing ACK cluster** (AliCloudKubernetesCluster) to attach the node pool to
- **At least one VSwitch** in the same VPC as the parent cluster
- **An SSH key pair** or password for node access

## Quick Start

Create a file `node-pool.yaml`:

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKubernetesNodePool
metadata:
  name: my-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudKubernetesNodePool.my-pool
spec:
  region: cn-hangzhou
  clusterId:
    value: c-abc123
  name: my-pool
  vswitchIds:
    - value: vsw-aaa111
    - value: vsw-bbb222
  instanceTypes:
    - ecs.g7.xlarge
  desiredSize: 2
  keyName: my-keypair
```

Deploy:

```shell
openmcf apply -f node-pool.yaml
```

This creates a two-node pool with AliyunLinux3, 120 GiB cloud_essd system disks, across two Availability Zones.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region. Must match the parent cluster's region. | Required; non-empty |
| `clusterId` | `StringValueOrRef` | ACK cluster ID that this node pool belongs to. | Required |
| `clusterId.value` | `string` | Direct cluster ID value. | — |
| `clusterId.valueFrom` | `object` | Foreign key reference to an AliCloudKubernetesCluster resource. | Default kind: `AliCloudKubernetesCluster`, field: `status.outputs.cluster_id` |
| `name` | `string` | Node pool name. | Required; 1–63 characters |
| `vswitchIds` | `StringValueOrRef[]` | VSwitch IDs for worker node placement. Use distinct AZs for HA. | 1–5 items required |
| `instanceTypes` | `string[]` | ECS instance types. Multiple types improve availability. | At least 1 required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `desiredSize` | `int` | — | Number of nodes. For auto-scaling pools, sets the initial count. Range: 0–1000. |
| `imageType` | `string` | `AliyunLinux3` | OS image type. Values: `AliyunLinux`, `AliyunLinux3`, `AliyunLinux3Arm64`, `Ubuntu`, `CentOS`, `Windows`, `ContainerOS`, `Custom`, and others. |
| `systemDisk.category` | `string` | `cloud_essd` | System disk type: `cloud_efficiency`, `cloud_ssd`, `cloud_essd`, `cloud_auto`. |
| `systemDisk.size` | `int` | `120` | System disk size in GiB. Range: 40–500. |
| `systemDisk.performanceLevel` | `string` | — | ESSD performance level: `PL0`, `PL1`, `PL2`, `PL3`. Only for `cloud_essd`. |
| `systemDisk.encrypted` | `bool` | `false` | Encrypt the system disk. |
| `systemDisk.kmsKeyId` | `string` | — | KMS key ID for disk encryption. |
| `dataDisks` | `DataDisk[]` | `[]` | Additional data disks per node. Each requires `size` (40–32767 GiB). |
| `securityGroupIds` | `StringValueOrRef[]` | Cluster default | Security groups for nodes. Immutable after creation. Can reference AliCloudSecurityGroup. |
| `internetMaxBandwidthOut` | `int` | `0` | Max outbound bandwidth in Mbps. >0 allocates a public IP. Range: 0–100. |
| `internetChargeType` | `string` | `PayByTraffic` | Public internet billing: `PayByBandwidth` or `PayByTraffic`. |
| `keyName` | `string` | — | SSH key pair name. Mutually exclusive with `password`. |
| `password` | `string` | — | SSH password. Mutually exclusive with `keyName`. Sensitive. |
| `labels` | `map<string, string>` | `{}` | Kubernetes labels for pod scheduling (nodeSelector, affinity). |
| `taints` | `Taint[]` | `[]` | Kubernetes taints. Each has `key`, `value`, `effect` (NoSchedule/PreferNoSchedule/NoExecute). |
| `cpuPolicy` | `string` | `none` | CPU management: `none` (CFS) or `static` (pin exclusive containers to CPUs). |
| `runtimeName` | `string` | Provider default | Container runtime: `containerd`, `Sandboxed-Container.runv`. |
| `runtimeVersion` | `string` | Latest | Container runtime version. |
| `unschedulable` | `bool` | `false` | Mark new nodes as unschedulable until manually uncordoned. |
| `userData` | `string` | — | Base64-encoded boot script. Max 16 KB before encoding. |
| `installCloudMonitor` | `bool` | `true` | Install Alibaba Cloud CloudMonitor agent. |
| `scalingConfig` | `object` | — | Auto-scaling configuration. |
| `scalingConfig.enable` | `bool` | `true` | Enable auto-scaling. |
| `scalingConfig.minSize` | `int` | — | Minimum node count. Range: 0–1000. |
| `scalingConfig.maxSize` | `int` | — | Maximum node count. Range: 0–2000. |
| `scalingConfig.type` | `string` | `cpu` | Instance classification: `cpu`, `gpu`, `gpushare`, `spot`. |
| `multiAzPolicy` | `string` | — | Multi-AZ distribution: `PRIORITY`, `COST_OPTIMIZED`, `BALANCE`. |
| `management` | `object` | — | Managed lifecycle settings. |
| `management.enable` | `bool` | `true` | Enable managed node pool features. |
| `management.autoRepair` | `bool` | — | Auto-replace unhealthy nodes. |
| `management.autoUpgrade` | `bool` | — | Auto-upgrade kubelet on cluster version change. |
| `management.maxUnavailable` | `int` | `1` | Max nodes unavailable during managed operations. Range: 0–1000. |
| `spotStrategy` | `string` | `NoSpot` | Spot strategy: `NoSpot`, `SpotWithPriceLimit`, `SpotAsPriceGo`. |
| `spotPriceLimits` | `SpotPriceLimit[]` | `[]` | Per-type price caps. Each has `instanceType` and `priceLimit` (CNY/hour). |
| `instanceChargeType` | `string` | `PostPaid` | Billing: `PostPaid` (pay-as-you-go) or `PrePaid` (subscription). |
| `period` | `int` | — | Subscription months (1, 2, 3, 6, 12). Required for PrePaid. |
| `autoRenew` | `bool` | — | Auto-renew subscription. |
| `autoRenewPeriod` | `int` | — | Auto-renewal period in months (1, 2, 3, 6, 12). |
| `tags` | `map<string, string>` | `{}` | Tags applied to ECS instances. |
| `resourceGroupId` | `string` | Default group | Resource group for organizational grouping. |
| `ramRoleName` | `string` | Cluster default | RAM role for worker nodes. Immutable after creation. |

## Examples

### Development Pool

A minimal fixed-size pool for development workloads.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKubernetesNodePool
metadata:
  name: dev-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudKubernetesNodePool.dev-pool
spec:
  region: cn-hangzhou
  clusterId:
    value: c-abc123
  name: dev-pool
  vswitchIds:
    - value: vsw-aaa111
    - value: vsw-bbb222
  instanceTypes:
    - ecs.g7.xlarge
  desiredSize: 2
  keyName: dev-keypair
```

### Production Auto-Scaling Pool

A production pool with auto-scaling, managed lifecycle, and multiple instance types.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKubernetesNodePool
metadata:
  name: prod-compute
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.AliCloudKubernetesNodePool.prod-compute
spec:
  region: cn-hangzhou
  clusterId:
    valueFrom:
      kind: AliCloudKubernetesCluster
      name: prod-cluster
      field: status.outputs.cluster_id
  name: prod-compute
  vswitchIds:
    - valueFrom:
        kind: AliCloudVswitch
        name: node-vsw-a
        field: status.outputs.vswitch_id
    - valueFrom:
        kind: AliCloudVswitch
        name: node-vsw-b
        field: status.outputs.vswitch_id
  instanceTypes:
    - ecs.c7.xlarge
    - ecs.c7.2xlarge
  desiredSize: 5
  keyName: prod-keypair
  labels:
    workload-type: compute
    team: platform
  scalingConfig:
    enable: true
    minSize: 3
    maxSize: 30
  multiAzPolicy: BALANCE
  management:
    enable: true
    autoRepair: true
    autoUpgrade: true
    maxUnavailable: 1
  systemDisk:
    category: cloud_essd
    size: 200
    performanceLevel: PL1
    encrypted: true
  tags:
    cost-center: infra-001
```

### Spot Batch Processing Pool

A cost-optimized pool using spot instances with taints for batch workload isolation.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKubernetesNodePool
metadata:
  name: batch-spot
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AliCloudKubernetesNodePool.batch-spot
spec:
  region: cn-hangzhou
  clusterId:
    value: c-abc123
  name: batch-spot
  vswitchIds:
    - value: vsw-aaa111
    - value: vsw-bbb222
  instanceTypes:
    - ecs.g7.xlarge
    - ecs.g7.2xlarge
    - ecs.c7.xlarge
  desiredSize: 0
  keyName: my-keypair
  spotStrategy: SpotWithPriceLimit
  spotPriceLimits:
    - instanceType: ecs.g7.xlarge
      priceLimit: "0.98"
    - instanceType: ecs.g7.2xlarge
      priceLimit: "1.96"
    - instanceType: ecs.c7.xlarge
      priceLimit: "0.85"
  taints:
    - key: workload-type
      value: batch
      effect: NoSchedule
  labels:
    workload-type: batch
  scalingConfig:
    enable: true
    minSize: 0
    maxSize: 50
    type: spot
  multiAzPolicy: COST_OPTIMIZED
  systemDisk:
    size: 200
  dataDisks:
    - category: cloud_essd
      size: 500
      name: batch-data
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `node_pool_id` | `string` | ACK node pool ID assigned by Alibaba Cloud |
| `scaling_group_id` | `string` | Auto Scaling group ID associated with this node pool |

## Related Components

- [AliCloudKubernetesCluster](/docs/catalog/alicloud/alicloudkubernetescluster) — the parent cluster that this node pool belongs to
- [AliCloudVswitch](/docs/catalog/alicloud/alicloudvswitch) — provides VSwitches for worker node placement
- [AliCloudSecurityGroup](/docs/catalog/alicloud/alicloudsecuritygroup) — controls network access for worker nodes
