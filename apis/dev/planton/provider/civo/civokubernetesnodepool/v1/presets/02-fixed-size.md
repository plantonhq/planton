# Fixed-Size Node Pool

This preset creates a static 3-node pool with no autoscaling. Suitable for workloads with predictable, steady resource requirements where the overhead of autoscaler decisions is unnecessary.

## When to Use

- Workloads with consistent, predictable resource demands
- Stateful services (databases, message brokers) that should not be disrupted by scale-down events
- Environments where cost predictability is more important than elastic scaling

## Key Configuration Choices

- **No autoscaling** (`autoScale` omitted) -- fixed node count for predictable capacity and cost
- **3 nodes** (`nodeCount: 3`) -- provides pod spread across nodes for resilience; adjust to match your workload
- **Medium nodes** (`size: g4s.kube.medium`) -- balanced for general workloads; use `g4s.kube.large` for memory-intensive services

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cluster-name>` | Name of the target CivoKubernetesCluster | `CivoKubernetesCluster` metadata.name |

## Related Presets

- **01-autoscaling** -- Use instead when workload demand varies and cost optimization through scale-down is desired
