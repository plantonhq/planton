# Non-Sharded Cluster

This preset creates a non-sharded OCI Cache (Redis) cluster with 3 nodes: one primary and two replicas. The primary handles all writes while replicas serve read traffic and provide automatic failover. This is the standard deployment for application caching, session storage, and rate limiting where the dataset fits within a single node's memory.

## When to Use

- Application-level caching (database query results, API responses, computed values)
- HTTP session storage for stateless web application deployments
- Rate limiting and leaky bucket counters for API gateways
- Pub/sub messaging for real-time features (chat, notifications)
- Any Redis workload where the dataset fits in a single node's memory (up to 32 GB per node)

## Key Configuration Choices

- **Non-sharded mode** (`clusterMode: nonsharded`) -- all data resides on a single primary node. Simpler to operate than sharded mode and supports all Redis commands without cross-slot restrictions. Choose sharded mode only when the dataset exceeds single-node memory.
- **3 nodes** (`nodeCount: 3`) -- one primary plus two replicas. Two replicas ensure continued read scaling and failover capability even if one replica is lost. Minimum recommended for production.
- **8 GB per node** (`nodeMemoryInGbs: 8`) -- provides 8 GB of cache capacity (primary dataset) with the same dataset replicated on each replica. Adjust based on expected cache working set.
- **Redis 7.1.1** (`softwareVersion: V7.1.1`) -- latest available version with Redis Functions, ACL improvements, and performance optimizations.
- **NSG-protected** (`nsgIds`) -- restricts network access to the Redis cluster via security rules (port 6379).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the cluster | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<private-subnet-ocid>` | OCID of the private subnet for the cluster | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<cache-nsg-ocid>` | OCID of the NSG allowing Redis traffic (port 6379) | OCI Console > Networking > NSGs, or `OciSecurityGroup` outputs |

## Related Presets

- **02-sharded-cluster** -- Use instead when the dataset exceeds single-node memory or write throughput needs horizontal scaling
