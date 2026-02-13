# ScalewayKapsulePool Examples

Practical examples for deploying additional node pools to Scaleway Kapsule clusters. Each example is a complete, deployable manifest.

## Table of Contents

- [Example 1: Basic Fixed-Size Pool](#example-1-basic-fixed-size-pool)
- [Example 2: Autoscaling Production Pool](#example-2-autoscaling-production-pool)
- [Example 3: GPU Pool with Taints](#example-3-gpu-pool-with-taints)
- [Example 4: Multi-Pool Architecture](#example-4-multi-pool-architecture)
- [Example 5: Infra Chart Composition](#example-5-infra-chart-composition)

---

## Example 1: Basic Fixed-Size Pool

A simple, fixed-size pool for general workloads. No autoscaling, no labels, no taints.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: app-workers
spec:
  region: fr-par
  clusterId:
    value: "fr-par/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  nodeType: GP1-XS
  size: 3
  autohealing: true
  publicIpDisabled: true
```

**Estimated cost**: ~150 EUR/month (3x GP1-XS nodes).

---

## Example 2: Autoscaling Production Pool

A production pool that scales based on workload demand. Labeled for workload scheduling, with autohealing and private networking.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: prod-workers
  org: mycompany
  env: production
spec:
  region: fr-par
  clusterId:
    value: "fr-par/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  nodeType: PRO2-M
  size: 5
  autoScale: true
  minSize: 3
  maxSize: 10
  autohealing: true
  publicIpDisabled: true
  rootVolumeSizeInGb: 100
  kubernetesLabels:
    workload: application
    tier: backend
    env: production
  upgradePolicy:
    maxSurge: 1
    maxUnavailable: 0
```

**Estimated cost**: ~240-800 EUR/month (3-10x PRO2-M nodes depending on autoscaler).

**Using labels for pod placement:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-api
spec:
  replicas: 3
  template:
    spec:
      nodeSelector:
        workload: application
        tier: backend
      containers:
        - name: api
          image: myapp:latest
```

---

## Example 3: GPU Pool with Taints

A specialized GPU pool for ML/AI workloads. Taints prevent non-GPU pods from consuming expensive hardware. Labels enable targeted scheduling.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: gpu-workers
spec:
  region: fr-par
  clusterId:
    value: "fr-par/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  nodeType: GPU-3070-S
  size: 2
  autoScale: true
  minSize: 0
  maxSize: 4
  autohealing: true
  publicIpDisabled: true
  kubernetesLabels:
    hardware: gpu
    workload: ml
  taints:
    - key: nvidia.com/gpu
      value: "true"
      effect: NoSchedule
```

**Estimated cost**: Variable (GPU instances, scales to zero when idle).

**Pods must tolerate the GPU taint to schedule on these nodes:**

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: training-job
spec:
  template:
    spec:
      tolerations:
        - key: nvidia.com/gpu
          operator: Equal
          value: "true"
          effect: NoSchedule
      nodeSelector:
        hardware: gpu
      containers:
        - name: trainer
          image: tensorflow/tensorflow:latest-gpu
          resources:
            limits:
              nvidia.com/gpu: 1
      restartPolicy: Never
```

---

## Example 4: Multi-Pool Architecture

A production architecture with separate pools for different workload classes. Each pool has specific instance types, labels, and (optionally) taints.

### Pool A: Application Workers

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: app-workers
  org: mycompany
  env: production
spec:
  region: fr-par
  clusterId:
    value: "fr-par/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  nodeType: PRO2-S
  size: 5
  autoScale: true
  minSize: 3
  maxSize: 10
  autohealing: true
  publicIpDisabled: true
  kubernetesLabels:
    workload: application
    tier: web
```

### Pool B: Batch Processing (CPU-Optimized)

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: batch-workers
  org: mycompany
  env: production
spec:
  region: fr-par
  clusterId:
    value: "fr-par/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  nodeType: GP1-S
  size: 2
  autoScale: true
  minSize: 0
  maxSize: 8
  autohealing: true
  publicIpDisabled: true
  kubernetesLabels:
    workload: batch
    compute: cpu-optimized
  taints:
    - key: workload
      value: batch
      effect: PreferNoSchedule
```

### Pool C: Zone-Specific High-Availability

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: ha-workers-par1
  org: mycompany
  env: production
spec:
  region: fr-par
  zone: fr-par-1
  clusterId:
    value: "fr-par/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  nodeType: PRO2-M
  size: 3
  autoScale: true
  minSize: 2
  maxSize: 5
  autohealing: true
  publicIpDisabled: true
  kubernetesLabels:
    topology: zone-1
    ha-group: primary
```

---

## Example 5: Infra Chart Composition

Shows how pools compose with the cluster in an infra chart template. The `clusterId` is wired via `valueFrom` to create a dependency edge.

```yaml
# Kapsule Cluster (Layer 2)
---
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsuleCluster
metadata:
  name: "{{ values.env }}-cluster"
spec:
  region: "{{ values.region }}"
  kubernetesVersion: "{{ values.kubernetes_version }}"
  cni: cilium
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: "{{ values.env }}-network"
      fieldPath: status.outputs.private_network_id
  deleteAdditionalResources: true
  defaultNodePool:
    nodeType: "{{ values.default_node_type }}"
    size: {{ values.default_node_count }}
    autohealing: true
    publicIpDisabled: true

# Application Workers Pool (Layer 3)
---
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: "{{ values.env }}-app-workers"
spec:
  region: "{{ values.region }}"
  clusterId:
    valueFrom:
      kind: ScalewayKapsuleCluster
      name: "{{ values.env }}-cluster"
      fieldPath: status.outputs.cluster_id
  nodeType: "{{ values.app_node_type }}"
  size: {{ values.app_node_count }}
  autoScale: {{ values.app_auto_scale }}
  {% if values.app_auto_scale | bool %}
  minSize: {{ values.app_min_nodes }}
  maxSize: {{ values.app_max_nodes }}
  {% endif %}
  autohealing: true
  publicIpDisabled: true
  kubernetesLabels:
    workload: application
    env: "{{ values.env }}"

# GPU Workers Pool (Layer 3, conditional)
{% if values.enable_gpu_pool | default(false) | bool %}
---
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: "{{ values.env }}-gpu-workers"
spec:
  region: "{{ values.region }}"
  clusterId:
    valueFrom:
      kind: ScalewayKapsuleCluster
      name: "{{ values.env }}-cluster"
      fieldPath: status.outputs.cluster_id
  nodeType: "{{ values.gpu_node_type }}"
  size: {{ values.gpu_node_count | default(1) }}
  autoScale: true
  minSize: 0
  maxSize: {{ values.gpu_max_nodes | default(4) }}
  autohealing: true
  publicIpDisabled: true
  kubernetesLabels:
    hardware: gpu
    workload: ml
  taints:
    - key: nvidia.com/gpu
      value: "true"
      effect: NoSchedule
{% endif %}
```

---

## Common Operations

### Get Pool Status

```bash
# Get detailed pool information
openmcf get scalewaykapsulepool <pool-name> -o yaml

# Get pool outputs (ID, version, current size)
openmcf stack-outputs --manifest pool.yaml
```

### Scale Pool

```bash
# Edit manifest to change size (or min/max for autoscaling pools)
# Then apply changes
openmcf pulumi up --manifest pool.yaml --yes
```

### Delete Pool

```bash
# Delete pool (drains nodes first)
openmcf delete -f pool.yaml
```

---

## Best Practices

### Pool Sizing

1. **Start small** -- Begin with minimal nodes and enable autoscaling.
2. **Separate workloads** -- Use multiple pools for different workload types.
3. **Right-size instances** -- Match instance types to actual resource needs.
4. **Scale-to-zero for batch** -- Use `minSize: 0` for non-critical batch pools.

### Labels and Taints

1. **Use labels for selection** -- Pods use `nodeSelector` or node affinity.
2. **Use taints for isolation** -- Prevent unwanted pods on specialized nodes.
3. **Keep conventions consistent** -- Use the same label keys across pools and clusters.
4. **Document taint requirements** -- Ensure application teams know which tolerations are needed.

### Production Recommendations

1. **Enable autohealing** -- Always set `autohealing: true` for production pools.
2. **Disable public IPs** -- Set `publicIpDisabled: true` for security.
3. **Configure upgrade policy** -- Use `maxSurge: 1, maxUnavailable: 0` for zero-downtime upgrades.
4. **Set pod resource requests** -- The autoscaler needs resource requests to make scaling decisions.
5. **Use placement groups** -- Spread nodes across hypervisors for high availability.
