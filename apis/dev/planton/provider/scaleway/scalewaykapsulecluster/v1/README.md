# ScalewayKapsuleCluster

A managed Kubernetes cluster on Scaleway's Kapsule platform with an embedded default node pool.

## What It Provisions

This is a **composite resource** that bundles two Scaleway resources:

1. **`scaleway_k8s_cluster`** -- The managed Kubernetes control plane (API server, etcd, scheduler, controller-manager). Fully managed by Scaleway with automatic failover and updates.
2. **`scaleway_k8s_pool`** -- A default node pool that provides immediate compute capacity. The cluster is usable out of the box.

Additional node pools with different instance types, GPU nodes, or workload isolation can be added via separate [ScalewayKapsulePool](../scalewaykapsulepool/v1/) resources.

## Key Features

- **Managed Control Plane** -- Scaleway operates etcd, API server, and control plane components with built-in HA.
- **CNI Selection** -- Choose between Cilium (eBPF, recommended) and Calico for pod networking.
- **Auto-Upgrade** -- Automatic patch version upgrades during configurable maintenance windows.
- **Cluster Autoscaler** -- Cluster-wide autoscaler configuration (scale-down delays, utilization thresholds, expander strategies) applied to all autoscaling-enabled pools.
- **Private Network Isolation** -- Cluster runs on a dedicated Private Network for secure node-to-node and node-to-control-plane communication.
- **Dedicated Control Planes** -- Optional dedicated tiers (`kapsule-dedicated-4/8/16`) for production workloads requiring isolated API server resources.

## Dependencies

### Upstream (this resource consumes)

| Resource | Field | Purpose |
|----------|-------|---------|
| [ScalewayPrivateNetwork](../scalewayprivatenetwork/v1/) | `private_network_id` (required, StringValueOrRef) | Network isolation for cluster nodes |

### Downstream (these resources consume this)

| Resource | Field | Purpose |
|----------|-------|---------|
| [ScalewayKapsulePool](../scalewaykapsulepool/v1/) | `cluster_id` (StringValueOrRef) | Additional node pools reference this cluster |
| Kubernetes addons (via infra charts) | `kubeconfig`, `apiserver_url`, `cluster_ca_certificate` | Configure Kubernetes provider for addon deployment |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `cluster_id` | Regional cluster ID -- primary reference for downstream pools |
| `kubeconfig` | Raw kubeconfig file content (sensitive) |
| `apiserver_url` | Kubernetes API server URL |
| `cluster_ca_certificate` | API server CA certificate (base64, sensitive) |
| `wildcard_dns` | DNS wildcard for ready nodes |
| `default_pool_id` | ID of the embedded default node pool |

## Important Constraints

- **CNI is immutable** -- Changing `cni` after creation recreates the cluster (data loss).
- **Private Network is immutable** -- Changing `private_network_id` recreates the cluster.
- **Pod/Service CIDRs are immutable** -- `pod_cidr` and `service_cidr` cannot be changed after creation.
- **Node type is immutable per pool** -- To change instance types, create a new pool and migrate workloads.
- **Version drift is ignored** -- Both Pulumi and Terraform ignore version changes to accommodate auto-upgrade. Manual version upgrades require temporary lifecycle override.

## Infra Chart Composition

In the `kapsule-environment` infra chart, this resource sits at Layer 2:

```
ScalewayVpc (Layer 0)
  └── ScalewayPrivateNetwork (Layer 1)
        └── ScalewayKapsuleCluster (Layer 2)
              └── ScalewayKapsulePool (Layer 3, additional pools)
                    └── Kubernetes Addons (Layer 4, cert-manager, ingress-nginx, etc.)
```

## Scaleway Documentation

- [Kapsule Overview](https://www.scaleway.com/en/docs/containers/kubernetes/quickstart/)
- [Kapsule Pricing](https://www.scaleway.com/en/pricing/?tags=available,managedkubernetes)
- [Cluster API Reference](https://www.scaleway.com/en/developers/api/kubernetes/)
- [CNI Comparison](https://www.scaleway.com/en/docs/containers/kubernetes/reference-content/cni-overview/)
