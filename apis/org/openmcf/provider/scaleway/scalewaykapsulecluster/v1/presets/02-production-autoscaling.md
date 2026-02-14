# Production Kubernetes Cluster with Autoscaling

This preset creates a production-grade Scaleway Kapsule cluster with autoscaling, automatic patch upgrades on Sunday mornings, private-only nodes, and autohealing. The default pool uses PRO2-S instances for guaranteed performance and scales between 2 and 10 nodes based on workload demand.

## When to Use

- Production applications requiring elastic compute capacity
- Workloads with variable traffic patterns that benefit from automatic node scaling
- Teams that want hands-off Kubernetes maintenance with automatic patch upgrades

## Key Configuration Choices

- **Shared control plane** (`type: kapsule`) -- sufficient for most production workloads; upgrade to `kapsule-dedicated-8` for isolated API server SLA
- **Cilium CNI** (`cni: cilium`) -- recommended eBPF-based networking
- **PRO2-S nodes** (`nodeType: PRO2-S`) -- 2 vCPU, 8 GB RAM; production-optimized with guaranteed resources
- **Autoscaling** (`autoScale: true`, 2-10 nodes) -- the cluster autoscaler adds nodes when pods are pending and removes underutilized nodes
- **Auto-upgrade** (Sunday 3:00 AM UTC) -- Scaleway automatically applies Kubernetes patch versions during the maintenance window
- **Autohealing enabled** (`autohealing: true`) -- unhealthy nodes are automatically detected and replaced
- **Private nodes** (`publicIpDisabled: true`) -- nodes have no public IPs; requires a Public Gateway or NAT for outbound internet access
- **Cleanup on delete** (`deleteAdditionalResources: true`) -- prevents orphaned LBs and PVCs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-private-network-id>` | UUID of the Private Network for cluster networking | Scaleway console or `ScalewayPrivateNetwork` status outputs |

## Related Presets

- **01-dev-minimal** -- Use instead for development with a small fixed-size pool and no auto-upgrade
