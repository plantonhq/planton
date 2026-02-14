# Autoscaling Production Node Pool

This preset creates an autoscaling node pool for a DigitalOcean Kubernetes cluster. It provisions general-purpose nodes with Kubernetes labels for workload scheduling and automatic scaling between 2 and 6 nodes based on pod demand.

## When to Use

- Production application workloads with variable traffic patterns
- Adding dedicated capacity separate from the cluster's default node pool
- Teams practicing the "sacrificial default pool" pattern (keep default pool minimal, run workloads on additional pools)

## Key Configuration Choices

- **Autoscaling** (`autoScale: true`, 2-6 nodes) -- the cluster autoscaler adds/removes nodes based on pending pod scheduling.
- **General-purpose nodes** (`size: s-4vcpu-8gb`) -- balanced for typical web/API workloads. Use dedicated CPU instances (`c-*`) for compute-intensive workloads.
- **Workload label** (`labels: {workload: app}`) -- enables `nodeSelector` or `nodeAffinity` rules to schedule pods exclusively on this pool.
- **Cluster reference** -- uses `metadata.name` to identify the parent cluster.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cluster-name>` | Name of the parent DOKS cluster | `DigitalOceanKubernetesCluster` resource `metadata.name` |

## Related Presets

- **02-fixed-size** -- Use instead for stable workloads where autoscaling is unnecessary
