# Warm Pool Node Group

This preset keeps pre-initialized nodes on standby so scale-out serves
pods in seconds instead of the minutes a cold boot takes (instance
launch + AMI boot + kubelet registration + image pulls). Worth its cost
exactly when boot time dominates your scale-out latency.

## When to Use

- Latency-sensitive services whose traffic spikes faster than a node
  cold-boots (checkout flows, game servers, live events)
- Fleets with heavy container images or long agent/cache warmup
- Batch systems whose queues surge and drain on a schedule

## Key Configuration Choices

- **`poolState: STOPPED`** -- the cost sweet spot: pooled nodes are
  fully initialized then stopped, billing only EBS storage until
  needed. `RUNNING` joins instantly at full instance price;
  `HIBERNATED` restores RAM from disk (fast JVM/cache warmup, storage
  cost only)
- **`minSize: 2`** -- the standby floor: how many warm nodes are always
  ready regardless of scale-in
- **`maxGroupPreparedCapacity: 4`** -- caps total pooled capacity; omit
  it to let AWS default the ceiling to the gap between the group's
  `maxSize` and its desired capacity. Explicit `0` means "nothing
  beyond minSize"
- **`reuseOnScaleIn: true`** -- scaled-in nodes go back to the pool
  instead of terminating, so the warm boot is reused
- **Warm pools update in place** -- adding one to a live group later is
  a plain update, not a group replacement

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<node-group-name>` | Name for the node group | Your environment naming convention |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) | Your deployment region |
| `<cluster-resource-name>` | Name of the AwsEksCluster this pool joins | Your cluster manifest's `metadata.name` |
| `<node-role-resource-name>` | Name of the AwsIamRole with the three worker policies | Your role manifest's `metadata.name` |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |

## Common Additions

- `taints` to reserve the warm capacity for the workload that needs it
- `labels` so the deployment's `nodeSelector` lands pods on warm nodes
- A `launch_template` reference when the warm nodes need a custom AMI
  or user-data warmup script (move `instanceTypes`/`diskSizeGb` into
  the template)

## Related Presets

- **01-on-demand-general** -- the baseline pool without standby capacity
- **02-spot-cost-optimized** -- cost-first capacity; note Spot and warm
  pools solve opposite problems (savings vs latency)
