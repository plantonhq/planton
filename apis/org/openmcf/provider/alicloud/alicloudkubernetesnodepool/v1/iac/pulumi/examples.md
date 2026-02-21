# AlicloudKubernetesNodePool Pulumi Examples

## CLI

```bash
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

---

## Minimal Node Pool

A basic fixed-size node pool with default settings.

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

**Key Points:**
- Single instance type, two nodes, two AZs
- Default system disk: 120 GiB cloud_essd
- Default image: AliyunLinux3
- No auto-scaling (fixed-size pool)

---

## Auto-Scaling Pool with Labels

A node pool with auto-scaling enabled and Kubernetes labels for pod scheduling.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesNodePool
metadata:
  name: compute-pool
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
  keyName: my-keypair
  labels:
    workload-type: compute
    team: platform
  scalingConfig:
    enable: true
    minSize: 2
    maxSize: 20
    type: cpu
  multiAzPolicy: BALANCE
  management:
    enable: true
    autoRepair: true
    autoUpgrade: true
    maxUnavailable: 1
  tags:
    env: production
```

**Key Points:**
- Two instance types for availability during capacity shortages
- Auto-scaling from 2 to 20 nodes based on pending pod demands
- BALANCE policy distributes nodes evenly across AZs
- Managed lifecycle with auto-repair and auto-upgrade
- Labels enable pod scheduling via `nodeSelector` or affinity

---

## Spot Instance Pool with Taints

A cost-optimized pool using spot instances with taints for workload isolation.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesNodePool
metadata:
  name: batch-spot-pool
spec:
  region: cn-hangzhou
  clusterId:
    value: c-abc123
  name: batch-spot-pool
  vswitchIds:
    - value: vsw-aaa111
    - value: vsw-bbb222
  instanceTypes:
    - ecs.g7.xlarge
    - ecs.g7.2xlarge
    - ecs.c7.xlarge
  desiredSize: 5
  keyName: my-keypair
  spotStrategy: SpotAsPriceGo
  taints:
    - key: workload-type
      value: batch
      effect: NoSchedule
  labels:
    workload-type: batch
    instance-lifecycle: spot
  scalingConfig:
    enable: true
    minSize: 0
    maxSize: 50
    type: spot
  multiAzPolicy: COST_OPTIMIZED
  systemDisk:
    category: cloud_essd
    size: 200
  dataDisks:
    - category: cloud_essd
      size: 500
      name: batch-data
```

**Key Points:**
- Three instance types for maximum spot availability
- SpotAsPriceGo for lowest cost (market price)
- Taints ensure only batch workloads with matching tolerations run here
- COST_OPTIMIZED prefers the cheapest AZ
- Scales to zero when no batch jobs are pending
- 200 GiB system disk and 500 GiB data disk for batch data

---

**Next Steps:**

- See [README.md](./README.md) for CLI flows and debugging instructions
- See [overview.md](./overview.md) for module architecture details
