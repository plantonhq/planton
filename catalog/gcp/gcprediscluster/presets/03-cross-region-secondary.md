# Cross-Region Secondary

## Use Case

Disaster recovery for a production cache: a second Memorystore for Redis Cluster in another region that continuously replicates from the primary and serves reads there, ready to be promoted if the primary's region fails. Deploy the primary first (the **Production Cache** preset, optionally listing this cluster under `secondaryClusters`), then this secondary naming the primary.

## When to Use

- A cache whose loss in a regional outage would take the service down
- Serving reads to clients in a second region with local latency
- A planned regional migration: promote the secondary, then repoint clients

## What This Creates

- A three-shard, one-replica cluster in `europe-west1` matching the primary's shape
- `clusterRole: SECONDARY` replicating from `orders-cache`, referenced by its `name` output
- IAM authentication and TLS matching the primary, Private Service Connect endpoints in the EU VPC
- Deletion protection on and `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `region` | `europe-west1` | The DR region; must differ from the primary's. |
| `primaryCluster.cluster` | `orders-cache` | Your primary cluster, by reference or full resource path. |
| `shardCount` | `3` | Must equal the primary's shard count. |
| `pscConfigs[].network` | `prod-vpc-eu` | The VPC clients in the DR region read from; its `GcpServiceConnectionPolicy` must exist first. |

A secondary is read-only until promoted; roles are exchanged by an in-place update on both clusters during a planned switchover. Two clusters bill as two clusters, plus inter-region replication traffic.
