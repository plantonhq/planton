# AlicloudKubernetesNodePool Terraform Examples

Below are several examples demonstrating how to deploy ACK node pools with the OpenMCF Terraform module.

After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic Fixed-Size Pool

A minimal node pool with default settings.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesNodePool
metadata:
  name: basic-pool
spec:
  region: cn-hangzhou
  clusterId:
    value: c-abc123
  name: basic-pool
  vswitchIds:
    - value: vsw-aaa111
    - value: vsw-bbb222
  instanceTypes:
    - ecs.g7.xlarge
  desiredSize: 2
  keyName: my-keypair
```

This example:
- Creates a two-node pool across two AZs
- Uses defaults: AliyunLinux3, cloud_essd 120 GiB, PostPaid billing
- No auto-scaling (fixed-size)

---

## Auto-Scaling Pool with Managed Lifecycle

A production pool with auto-scaling, multiple instance types, and managed node operations.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesNodePool
metadata:
  name: compute-pool
  org: acme-corp
  env: production
spec:
  region: cn-hangzhou
  clusterId:
    value: c-abc123
  name: compute-pool
  vswitchIds:
    - value: vsw-aaa111
    - value: vsw-bbb222
  instanceTypes:
    - ecs.c7.xlarge
    - ecs.c7.2xlarge
  desiredSize: 3
  keyName: prod-keypair
  labels:
    workload-type: compute
  scalingConfig:
    enable: true
    minSize: 2
    maxSize: 20
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
```

This example:
- Two instance types for availability
- Auto-scales between 2 and 20 nodes
- Balanced AZ distribution
- Auto-repair and auto-upgrade enabled
- 200 GiB ESSD system disk at PL1

---

## Spot Instance Pool

A cost-optimized pool using spot instances with taints for workload isolation.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesNodePool
metadata:
  name: batch-spot
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
  desiredSize: 5
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
  scalingConfig:
    enable: true
    minSize: 0
    maxSize: 50
    type: spot
  multiAzPolicy: COST_OPTIMIZED
```

This configuration:
- Three instance types maximize spot availability
- Price limits prevent overpaying for spot capacity
- Taints ensure only batch workloads with tolerations run here
- Scales to zero when idle
- COST_OPTIMIZED fills the cheapest AZ first

---

## After Deploying

Verify the node pool using the Alibaba Cloud CLI:

```bash
# List node pools for a cluster
aliyun cs DescribeClusterNodePools --ClusterId <cluster-id>

# Get node pool details
aliyun cs DescribeClusterNodePoolDetail --ClusterId <cluster-id> --NodepoolId <node-pool-id>
```
