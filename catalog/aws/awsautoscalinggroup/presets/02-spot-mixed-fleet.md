# Spot Mixed Fleet

This preset runs the classic cost architecture: two guaranteed On-Demand
instances as the base, everything above them on Spot
(`onDemandPercentageAboveBaseCapacity: 0`), drawn from four interchangeable
instance types so no single pool interruption can starve the fleet.
Typical savings on the Spot majority run 60-90% versus On-Demand.

## When to Use

- Interruption-tolerant workloads: queue consumers, CI runners, batch
  processing, stateless web capacity above a guaranteed core
- Any fleet where cost matters more than per-instance longevity

## Key Configuration Choices

- **`onDemandBaseCapacity: 2`** -- the always-on core that survives even a
  total Spot drought
- **`price-capacity-optimized`** -- the AWS-recommended Spot strategy:
  weighs price AND interruption likelihood, instead of chasing the
  cheapest (and most volatile) pool
- **Four override types** -- m6i/m5/m5a/m4 are interchangeable for most
  workloads; more pools = fewer simultaneous interruptions
- **`capacityRebalance: true`** -- replaces at-risk Spot instances
  proactively, before the two-minute interruption notice
- **`AllocationStrategy` termination policy** -- scale-in keeps the fleet
  on its preferred pools

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<fleet-name>` | Name for the group | Your workload's name (e.g., `workers`) |
| `<aws-region>` | AWS region code (e.g., `us-east-1`) | Your deployment region |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |
| `<launch-template-resource-name>` | Name of the AwsLaunchTemplate resource | Your template manifest's `metadata.name` |

## Common Additions

- Replace explicit overrides with one `instanceRequirements` override
  (memory/vCPU ranges) to widen the pool set even further
- Add `weightedCapacity` per override when blending different sizes
- Add a queue-depth target-tracking policy (metric-math backlog per
  instance) for consumer fleets

## Related Presets

- **01-web-service-behind-alb** -- load-balanced fleet with ELB health and rolling refresh
- **03-scheduled-scale** -- time-based capacity with a warm pool
