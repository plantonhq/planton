# Basic Read Pool

This preset adds a single-node READ_POOL instance to an existing AlloyDB cluster for offloading read traffic from the bundled primary.

## When to Use

- Dev/staging read scaling where cost matters more than HA
- Workloads that can tolerate single-zone placement

## Key Configuration Choices

- **READ_POOL (default)** — scales reads independently of the cluster's bundled primary
- **nodeCount: 1 + ZONAL** — lowest-cost read pool shape
- **cpuCount: 2** — smallest practical compute for AlloyDB read nodes

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `my-gcp-project-123` | GCP project ID |
| `my-orders-cluster` | Your GcpAlloydbCluster resource name |
| `orders-read-pool` | Instance ID within the cluster |

## Related Presets

- **02-read-pool-ha** — regional pool with two nodes
- **03-read-pool-production** — connectors, TLS, and query insights

## Related Kinds

- [GcpAlloydbCluster](/docs/catalog/gcp/gcpalloydbcluster) — the cluster this instance attaches to
- [GcpAlloydbUser](/docs/catalog/gcp/gcpalloydbuser) — application credentials on the same cluster
