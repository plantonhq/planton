# Stockout-Resilient Default

The cluster-wide default class for a GKE cluster in a busy region. It lists four general-purpose machine families of the same size, so when Google runs out of one (`ZONE_RESOURCE_POOL_EXHAUSTED` across every zone of the region), GKE creates the next family's nodes instead of leaving pods pending.

## When to Use

- Any production GKE cluster, especially in us-central1, us-east1 and other regions that see machine-type stockouts.
- A cluster whose installs (cert-manager, Istio, operators) or rollouts must not stall because one machine type ran out.

## Key Configuration Choices

- **`name: default`.** It applies to every pod that selects no class. Turn on `clusterAutoscaling.defaultComputeClassEnabled` on the `GcpGkeCluster`, or the class only applies to pods that select it by name.
- **n2d, c3d, n2, t2d at 8 vCPUs.** Each takes pd-balanced boot disks and runs the same x86 images, so a pod scheduled on any of them behaves the same. Keep `minCores` equal across rules so a fallback node is not smaller than the pods expect. Avoid putting Hyperdisk-only families (C4, C4D, N4) in a list without `storage.bootDiskType: hyperdisk-balanced` on their rule.
- **`whenUnsatisfiable: ScaleUpAnyway`.** When all four families are out, GKE still creates a node from its default series, which keeps a default class from becoming the reason the cluster stalls. Use `DoNotScaleUp` instead when running on an unlisted machine is worse than waiting.
- **`nodePoolAutoCreation.enabled: true`.** GKE creates node pools for the families as needed. Without it the class can only use pools that already exist.
- **`activeMigration.optimizeRulePriority: true`.** Once n2d capacity returns, GKE moves pods back from the fallback families. Give your Deployments PodDisruptionBudgets so the move is gradual.

## Placeholders to Replace

None. Adjust the families and `minCores` to the shape your workloads were sized for.

## Related Presets

- **02-spot-first-on-demand-fallback**: the same idea for cost, Spot first.
- **04-reservation-first**: spend committed capacity before buying more.
