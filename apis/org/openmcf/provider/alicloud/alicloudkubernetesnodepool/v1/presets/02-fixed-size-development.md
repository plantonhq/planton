# Fixed-Size Development Node Pool

This preset creates a small, fixed-size node pool for development and testing. Two nodes across two availability zones provide basic resilience without the complexity of auto-scaling or managed lifecycle features. The minimal configuration keeps costs low and setup simple.

## When to Use

- Development and testing environments
- Proof-of-concept clusters with predictable, low resource requirements
- Environments where auto-scaling is unnecessary or undesirable
- Quick iteration without waiting for node scale-up events

## Key Configuration Choices

- **Fixed size** (`desiredSize: 2`, no `scalingConfig`) -- Two nodes provide basic resilience without auto-scaler overhead. The node count stays constant regardless of pod pressure.
- **Single instance type** (`ecs.g7.xlarge`) -- 4 vCPUs, 16 GB RAM per node. Sufficient for development workloads. A single type simplifies capacity planning for dev environments.
- **80 GB system disk** -- Smaller than the 120 GB production default. Adequate for development where fewer container images are cached simultaneously.
- **Two AZs** -- Minimum for basic zone fault tolerance without over-provisioning.
- **No management features** -- Auto-repair and auto-upgrade are omitted to keep the pool simple and avoid unexpected node recycling during development sessions.
- **No labels or taints** -- Development pools typically run mixed workloads without scheduling constraints.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code matching the parent cluster | Your cluster's region |
| `<your-cluster-id>` | ACK cluster ID | `AliCloudKubernetesCluster` stack outputs |
| `<vswitch-id-zone-a>` | VSwitch in first AZ | `AliCloudVswitch` stack outputs |
| `<vswitch-id-zone-b>` | VSwitch in second AZ | `AliCloudVswitch` stack outputs |
| `<your-ssh-key-pair>` | ECS SSH key pair name | ECS console or your key management system |

## Related Presets

- **01-general-purpose-autoscaling** -- Use for production with auto-scaling, managed lifecycle, and multi-AZ balance
- **03-cost-optimized-spot** -- Use for batch workloads where spot pricing reduces costs further
