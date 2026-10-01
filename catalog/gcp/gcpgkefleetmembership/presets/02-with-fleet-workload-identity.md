# With Fleet Workload Identity

## Use Case

Register a cluster from another project and turn on fleet Workload Identity, so its workloads authenticate as identities in the fleet's shared workload identity pool.

## When to Use

- A cluster in another project joining a central fleet
- Workloads that should use the same identities across every cluster in the fleet

## What This Creates

- The Fleet API on the fleet host project
- A membership for the cluster with its OIDC issuer

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `issuer` | the cluster's issuer | `https://container.googleapis.com/v1/` followed by the cluster's `cluster_id` (the `locations/` form). |
| `gkeCluster` | a literal cluster path | A `GcpGkeCluster` reference. |
