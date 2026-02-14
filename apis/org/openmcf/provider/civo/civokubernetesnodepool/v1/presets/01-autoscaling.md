# Autoscaling Node Pool

This preset creates an autoscaling node pool that dynamically adjusts between 2 and 5 nodes based on workload demand. Starting at 2 nodes ensures baseline capacity while allowing the cluster autoscaler to add nodes during traffic spikes and remove them when idle.

## When to Use

- Production workloads with variable traffic patterns
- Applications that experience periodic load spikes (e.g., business hours, batch jobs)
- Environments where cost optimization requires scaling down during low-demand periods

## Key Configuration Choices

- **Autoscaling enabled** (`autoScale: true`) -- the cluster autoscaler manages node count automatically
- **2-5 node range** (`minNodes: 2`, `maxNodes: 5`) -- baseline HA at 2 nodes, burst capacity up to 5; adjust based on your peak-to-trough ratio
- **Medium nodes** (`size: g4s.kube.medium`) -- balanced CPU/RAM per node; prefer more smaller nodes over fewer large ones for better bin-packing
- **Initial count** (`nodeCount: 2`) -- starts at the minimum; the autoscaler scales up within seconds when pods are pending

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cluster-name>` | Name of the target CivoKubernetesCluster | `CivoKubernetesCluster` metadata.name |

## Related Presets

- **02-fixed-size** -- Use instead when workload is predictable and autoscaling adds unnecessary complexity
