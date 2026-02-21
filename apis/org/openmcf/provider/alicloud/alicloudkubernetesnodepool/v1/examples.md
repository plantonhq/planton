# AliCloudKubernetesNodePool Examples

## Minimal Fixed-Size Pool

A basic node pool with a fixed number of nodes, suitable for development or
small workloads.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKubernetesNodePool
metadata:
  name: dev-pool
spec:
  region: cn-hangzhou
  clusterId: c-abc123
  name: dev-pool
  vswitchIds:
    - vsw-aaa111
    - vsw-bbb222
  instanceTypes:
    - ecs.g7.xlarge
  desiredSize: 2
  keyName: dev-keypair
```

## Auto-Scaling Pool

A node pool with auto-scaling enabled. The cluster auto-scaler adjusts the
node count between 2 and 20 based on pending pod resource requests.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKubernetesNodePool
metadata:
  name: compute-pool
spec:
  region: cn-hangzhou
  clusterId: c-abc123
  name: compute-pool
  vswitchIds:
    - vsw-az-a
    - vsw-az-b
    - vsw-az-c
  instanceTypes:
    - ecs.g7.2xlarge
    - ecs.g7.xlarge
  desiredSize: 3
  imageType: AliyunLinux3
  systemDisk:
    category: cloud_essd
    size: 200
    performanceLevel: PL1
  dataDisk:
    - category: cloud_essd
      size: 500
      performanceLevel: PL1
  keyName: prod-keypair
  labels:
    workloadType: compute
    tier: standard
  scalingConfig:
    enable: true
    minSize: 2
    maxSize: 20
  multiAzPolicy: BALANCE
  management:
    enable: true
    autoRepair: true
    autoUpgrade: true
    maxUnavailable: 2
  installCloudMonitor: true
  tags:
    team: platform
```

## Production Spot Pool

A cost-optimized production pool using spot instances with price limits. Multiple
instance types maximize availability across spot pools. Falls back to on-demand
when spot capacity is unavailable.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKubernetesNodePool
metadata:
  name: spot-pool
  org: acme-corp
  env: production
spec:
  region: cn-hangzhou
  clusterId: c-prod-001
  name: spot-compute-pool
  vswitchIds:
    - vsw-prod-a
    - vsw-prod-b
    - vsw-prod-c
  instanceTypes:
    - ecs.g7.xlarge
    - ecs.g7.2xlarge
    - ecs.c7.xlarge
    - ecs.c7.2xlarge
  desiredSize: 5
  imageType: AliyunLinux3
  systemDisk:
    category: cloud_essd
    size: 120
    encrypted: true
  dataDisk:
    - category: cloud_essd
      size: 200
      encrypted: "true"
  securityGroupIds:
    - sg-prod-workers
  keyName: prod-keypair
  labels:
    workloadType: batch
    costModel: spot
  taints:
    - key: spot-instance
      value: "true"
      effect: PreferNoSchedule
  cpuPolicy: none
  runtimeName: containerd
  scalingConfig:
    enable: true
    minSize: 2
    maxSize: 50
    type: spot
  multiAzPolicy: COST_OPTIMIZED
  management:
    enable: true
    autoRepair: true
    autoUpgrade: true
    maxUnavailable: 5
  spotStrategy: SpotWithPriceLimit
  spotPriceLimits:
    - instanceType: ecs.g7.xlarge
      priceLimit: "0.98"
    - instanceType: ecs.g7.2xlarge
      priceLimit: "1.96"
    - instanceType: ecs.c7.xlarge
      priceLimit: "0.85"
    - instanceType: ecs.c7.2xlarge
      priceLimit: "1.70"
  installCloudMonitor: true
  tags:
    team: platform
    costCenter: infra-001
  resourceGroupId: rg-acme-prod
```
