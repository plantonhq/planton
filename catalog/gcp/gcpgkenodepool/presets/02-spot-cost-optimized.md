# Spot Cost-Optimized Pool

This preset creates a scale-to-zero Spot pool for fault-tolerant workloads: deeply discounted capacity (60-91%) that costs nothing while idle, fenced off from general workloads by a taint.

## When to Use

- Batch jobs, CI runners, queue workers, and other retry-tolerant workloads
- Bursty compute that should not pay for idle nodes
- Cost-optimizing a cluster after the on-demand primary pool is in place

## Key Configuration Choices

- **`spot: true`** — the current preemptible model (no 24-hour lifetime); nodes can be reclaimed with 30 seconds notice
- **Scale-to-zero (`minNodes: 0`)** — the pool exists only while tolerating workloads need it
- **`locationPolicy: ANY`** — lets the autoscaler hunt Spot capacity across all zones and prefer unused reservations, cutting preemption risk
- **Taint + label pair** — `workload-class=batch:NO_SCHEDULE` keeps everything without a matching toleration off Spot nodes; workloads opt in explicitly

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gke-cluster` | Your `GcpGkeCluster` resource name | Your cluster manifest |

## Related Presets

- **01-on-demand-autoscaling** — the guaranteed-capacity primary pool to pair with
- **03-gpu-accelerated** — GPU nodes (also commonly run on Spot)

## Related Kinds

- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — the control plane this pool attaches to
