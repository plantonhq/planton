# Spot Cost-Optimized Pool

This preset runs an interruptible batch/burst pool on Spot capacity at a
steep discount, with the two Spot survival practices built in: several
similar instance types for pool diversity, and a taint so only
interruption-tolerant workloads schedule onto it.

## When to Use

- Batch jobs, CI runners, queue workers, and other retry-friendly
  workloads
- Burst capacity alongside an On-Demand base pool

## Key Configuration Choices

- **Three similar `instanceTypes`** -- Spot interruptions hit one
  capacity pool at a time; diversity keeps the fleet alive when one pool
  is reclaimed
- **`capacityType: spot`** -- typically 60-90% cheaper than On-Demand
- **The `node-lifecycle=spot` taint + matching label** -- workloads must
  tolerate the taint to land here, so a Spot reclaim never takes down a
  pod that could not handle it
- **`minSize: 0`** -- the pool scales to nothing when idle

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<node-group-name>` | Name for the pool | Your naming convention (e.g., `batch-spot`) |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) | Your deployment region |
| `<cluster-resource-name>` | Name of the AwsEksCluster resource | Your cluster manifest's `metadata.name` |
| `<node-role-resource-name>` | Name of the AwsIamRole with the worker policies | Your role manifest's `metadata.name` |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |

## Common Additions

- `nodeRepairConfig.enabled: true` to auto-replace unhealthy survivors
- A cluster autoscaler (or Karpenter) to drive `desiredSize` from queue
  depth instead of a fixed count

## Related Presets

- **01-on-demand-general** -- the predictable-capacity workhorse pool
- **03-launch-template** -- custom launch mechanics for either capacity
  type
