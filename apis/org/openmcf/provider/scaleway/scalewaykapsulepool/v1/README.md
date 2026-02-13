# ScalewayKapsulePool

An additional worker node pool for an existing Scaleway Kapsule Kubernetes cluster.

## What It Provisions

This is a **standalone resource** (not composite) that creates a single Scaleway Kubernetes node pool (`scaleway_k8s_pool`). The pool provides compute capacity for Kubernetes workloads, independently scalable and configurable from the cluster's default pool.

Use this resource when you need:

- **Different instance types** -- GPU pools, high-memory pools, CPU-optimized pools alongside the default pool.
- **Workload isolation** -- Dedicated pools for specific teams, environments, or workload categories using Kubernetes labels and taints.
- **Independent scaling** -- Pools that autoscale based on their own workload demands, with different min/max boundaries.
- **Zone-specific placement** -- Pools pinned to specific availability zones for data locality or compliance.

## Key Features

- **First-Class Labels and Taints** -- Kubernetes node labels and taints are structured fields in the spec, not raw strings. Under the hood, they are applied via Scaleway's Cloud Controller Manager (CCM) tag convention, automatically synced to Kubernetes nodes.
- **Cluster Autoscaler Integration** -- Per-pool autoscaling that works with the cluster-level autoscaler configuration set on the parent ScalewayKapsuleCluster.
- **Autohealing** -- Automatic detection and replacement of unhealthy nodes.
- **Zone Placement** -- Optional zone pinning for multi-AZ architectures.
- **Anti-Affinity** -- Placement group support for spreading nodes across hypervisors.
- **Kubelet Arguments** -- Power-user escape hatch for custom kubelet configuration.

## How Labels and Taints Work

Scaleway does not have native label/taint fields on node pools. Instead, the Scaleway Cloud Controller Manager reads pool tags and syncs them to Kubernetes nodes:

| Spec Field | Tag Format | K8s Result |
|---|---|---|
| `kubernetes_labels: {workload: gpu}` | `noprefix=workload=gpu` | Node label `workload=gpu` |
| `taints: [{key: nvidia.com/gpu, value: "true", effect: NoSchedule}]` | `taint=noprefix=nvidia.com/gpu=true:NoSchedule` | Taint `nvidia.com/gpu=true:NoSchedule` |

The `noprefix=` convention ensures labels and taints use exactly the keys you specify (no `k8s.scaleway.com/` prefix). This abstraction is transparent -- you work with structured fields, and the IaC module generates the correct tags.

Reference: [Scaleway CCM Tag Documentation](https://github.com/scaleway/scaleway-cloud-controller-manager/blob/master/docs/tags.md)

## Dependencies

### Upstream (this resource consumes)

| Resource | Field | Purpose |
|----------|-------|---------|
| [ScalewayKapsuleCluster](../scalewaykapsulecluster/v1/) | `cluster_id` (required, StringValueOrRef) | Parent cluster for this node pool |

### Downstream (these resources consume this)

| Resource | Field | Purpose |
|----------|-------|---------|
| Kubernetes addons (via infra charts) | `runs_on` relationship in metadata | Topology tracking for addon deployment |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `pool_id` | Regional pool ID for management and monitoring |
| `pool_version` | Actual Kubernetes version on pool nodes |
| `current_size` | Actual node count (may differ from spec when autoscaling) |

## Important Constraints

- **Node type is immutable** -- Changing `node_type` recreates the pool. Create a new pool and migrate workloads instead.
- **Container runtime is immutable** -- Changing `container_runtime` recreates the pool.
- **Zone is immutable** -- Changing `zone` recreates the pool.
- **Placement group is immutable** -- Changing `placement_group_id` recreates the pool.
- **Public IP setting is immutable** -- Changing `public_ip_disabled` recreates the pool.
- **Region must match cluster** -- The pool's region must match the parent cluster's region.
- **Size ignored when autoscaling** -- When `auto_scale` is true, updates to `size` are ignored by the provider; the autoscaler controls the actual count.

## Infra Chart Composition

In the `kapsule-environment` infra chart, this resource sits at Layer 3:

```
ScalewayVpc (Layer 0)
  └── ScalewayPrivateNetwork (Layer 1)
        └── ScalewayKapsuleCluster (Layer 2, includes default pool)
              └── ScalewayKapsulePool (Layer 3, additional pools)
                    └── Kubernetes Addons (Layer 4, cert-manager, ingress-nginx, etc.)
```

Pools are wired to the cluster via `valueFrom`:

```yaml
spec:
  clusterId:
    valueFrom:
      kind: ScalewayKapsuleCluster
      name: "{{ values.cluster_name }}"
      fieldPath: status.outputs.cluster_id
```

## Scaleway Documentation

- [Kapsule Node Pools](https://www.scaleway.com/en/docs/containers/kubernetes/how-to/manage-a-pool/)
- [Instance Types](https://www.scaleway.com/en/pricing/?tags=available,compute)
- [CCM Tag Convention](https://github.com/scaleway/scaleway-cloud-controller-manager/blob/master/docs/tags.md)
- [Placement Groups](https://www.scaleway.com/en/docs/compute/instances/how-to/use-placement-groups/)
