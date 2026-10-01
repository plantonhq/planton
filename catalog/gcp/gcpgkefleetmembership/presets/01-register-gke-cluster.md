# Register a GKE Cluster

## Use Case

Bring an existing GKE cluster, created without a fleet, into the fleet so scopes and fleet features can use it.

## When to Use

- A cluster created outside Planton, or before the fleet existed
- Never for a cluster that sets `fleetProject` -- it is already registered

## What This Creates

- The Fleet API on the fleet host project
- A membership for the cluster, named after it

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `gkeCluster` | a literal cluster path | A `GcpGkeCluster` reference (its `cluster_id`). |
| `membershipId` | `metadata.name` | A different membership ID. |
