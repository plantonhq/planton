# Cost-Optimized Spot Instance Node Pool

This preset creates a node pool using spot instances with price caps for significant cost savings (typically 60-90% off on-demand pricing). Four instance types across three AZs maximize spot pool diversity, reducing the probability of simultaneous capacity reclamation. A taint prevents workloads that have not explicitly opted in to spot scheduling from landing on these nodes.

## When to Use

- Batch processing, CI/CD jobs, and data pipelines that tolerate interruptions
- Non-critical workloads where cost savings outweigh availability guarantees
- Supplementary capacity alongside an on-demand node pool for burst scaling
- Workloads with built-in retry logic and checkpointing

## Key Configuration Choices

- **SpotWithPriceLimit** (`spotStrategy: SpotWithPriceLimit`) -- Spot instances with per-type price caps. When the market price exceeds the cap, no new instances are created at that price; the auto-scaler tries other instance types instead. Price limits are set at roughly 70% of on-demand pricing as a starting point.
- **Four instance types** (`ecs.g7.xlarge`, `ecs.g7.2xlarge`, `ecs.c7.xlarge`, `ecs.c7.2xlarge`) -- Diversifying across general-purpose (g7) and compute-optimized (c7) families with two sizes each gives the spot auto-scaler four independent capacity pools. More pools means higher availability.
- **COST_OPTIMIZED multi-AZ** (`multiAzPolicy: COST_OPTIMIZED`) -- Allocates nodes in the cheapest AZ first, maximizing cost savings. Unlike BALANCE, nodes may cluster in one AZ when prices diverge.
- **Scale to zero** (`minSize: 0`) -- The pool can shrink to zero when no spot-tolerant workloads are pending, eliminating idle costs entirely.
- **Taint** (`spot-instance=true:PreferNoSchedule`) -- A soft taint that discourages the scheduler from placing regular workloads on spot nodes. Pods that tolerate spot instances (via a matching toleration) are scheduled normally. PreferNoSchedule is used instead of NoSchedule to allow graceful overflow when on-demand pools are full.
- **Label** (`costModel: spot`) -- Enables node selectors and pod affinity rules that explicitly target spot capacity.
- **Higher maxUnavailable** (`maxUnavailable: 5`) -- Spot nodes are inherently transient. Allowing more simultaneous unavailable nodes during repairs and reclamations avoids queuing delays.
- **No auto-upgrade** -- Spot nodes are short-lived; upgrading them is wasteful. New nodes created by the auto-scaler automatically use the latest AMI and kubelet version.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code matching the parent cluster | Your cluster's region |
| `<your-cluster-id>` | ACK cluster ID | `AlicloudKubernetesCluster` stack outputs |
| `<vswitch-id-zone-a>` | VSwitch in first AZ | `AlicloudVswitch` stack outputs |
| `<vswitch-id-zone-b>` | VSwitch in second AZ | `AlicloudVswitch` stack outputs |
| `<vswitch-id-zone-c>` | VSwitch in third AZ | `AlicloudVswitch` stack outputs |
| `<your-ssh-key-pair>` | ECS SSH key pair name | ECS console or your key management system |
| `<your-team>` | Team or business unit | Your organizational structure |

## Related Presets

- **01-general-purpose-autoscaling** -- Use for production workloads that require on-demand instance availability guarantees
- **02-fixed-size-development** -- Use for small development pools without auto-scaling
