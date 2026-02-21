# Sharded Cluster

This preset creates a sharded OCI Cache (Redis) cluster with 3 shards and 3 nodes per shard (9 nodes total). Data is automatically distributed across shards using Redis Cluster's hash slot mechanism, providing horizontal scaling for both reads and writes. Use this when the dataset exceeds single-node memory or write throughput is the bottleneck.

## When to Use

- Large datasets that exceed the memory capacity of a single Redis node
- Write-heavy workloads where a single primary cannot sustain the required throughput
- Real-time analytics pipelines with high-cardinality sorted sets or HyperLogLogs
- Multi-tenant caching where tenant data is naturally distributed across hash slots

## Key Configuration Choices

- **Sharded mode** (`clusterMode: sharded`) -- enables Redis Cluster protocol with automatic hash slot distribution across shards. Applications must use a Redis Cluster-aware client (all modern clients support this).
- **3 shards** (`shardCount: 3`) -- data is distributed across 3 shards, tripling both memory capacity and write throughput compared to a single-shard setup. Each shard owns ~5,461 of Redis's 16,384 hash slots.
- **3 nodes per shard** (`nodeCount: 3`) -- one primary and two replicas per shard, matching the non-sharded preset's HA pattern within each shard.
- **16 GB per node** (`nodeMemoryInGbs: 16`) -- 48 GB total cache capacity across 3 shard primaries (16 GB x 3 shards). Each shard handles roughly one-third of the keyspace.
- **NSG-protected** (`nsgIds`) -- restricts network access via security rules (ports 6379 and 16379 for cluster bus).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the cluster | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<private-subnet-ocid>` | OCID of the private subnet for the cluster | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<cache-nsg-ocid>` | OCID of the NSG allowing Redis traffic (ports 6379 and 16379) | OCI Console > Networking > NSGs, or `OciSecurityGroup` outputs |

## Related Presets

- **01-non-sharded-cluster** -- Use instead when the dataset fits in a single node and simpler operations are preferred
