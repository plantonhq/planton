# Spot First, On-Demand Fallback

A class for workloads that tolerate interruption. It runs them on Spot VMs across three families, and only buys on-demand capacity when Spot has run out everywhere, then moves the pods back to Spot when it returns.

## When to Use

- Batch jobs, queue consumers, CI runners and stateless services behind a load balancer.
- A cost-sensitive cluster that cannot afford pods sitting pending while Spot is unavailable.

## Key Configuration Choices

- **Three Spot families before any on-demand rule.** Spot capacity runs out per family and per zone. Listing several families keeps the discount available far more often than one.
- **On-demand rules last.** They are the fallback, not an extra. Without them, a Spot shortage leaves the pods pending.
- **`activeMigration.optimizeRulePriority: true`.** When Spot capacity returns, GKE moves pods off the on-demand nodes, so the fallback costs only as long as the shortage lasts.
- **`consolidationDelayMinutes: 5`.** Idle nodes go away five minutes after their last pod, which suits bursty work.
- **Selected by name.** Pods opt in with `cloud.google.com/compute-class: spot-first`. GKE adds the toleration for the class's taint, so nothing else lands on these nodes.

Spot VMs can be reclaimed with 30 seconds' notice. Keep stateful and singleton workloads on an on-demand class.

## Placeholders to Replace

None.

## Related Presets

- **01-stockout-resilient-default**: the on-demand cluster default.
