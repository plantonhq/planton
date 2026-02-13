---
title: "Cosmosdbaccount"
description: "Cosmosdbaccount deployment documentation"
icon: "package"
order: 100
componentName: "azurecosmosdbaccount"
---

# Azure Cosmos DB — Deployment Landscape and Design Rationale

## Overview of Azure Cosmos DB Service

Azure Cosmos DB is Microsoft's globally distributed, multi-model database service. It provides single-digit millisecond latency at the 99th percentile, automatic and instant scalability, and five well-defined consistency levels. Unlike traditional relational databases, Cosmos DB is designed for horizontal partitioning, uses Request Units (RU/s) as the unit of throughput, and supports multiple APIs through a single account resource.

Key architectural concepts:

- **Multi-model**: One account can expose SQL, MongoDB, Cassandra, Gremlin, or Table APIs
- **Globally distributed**: Data can replicate to any number of Azure regions with configurable failover
- **Tunable consistency**: Trade latency and throughput for stronger or weaker consistency guarantees
- **Partition-based scaling**: Data is distributed by partition key; throughput scales per partition

## API Comparison: Why We Support SQL and MongoDB

Azure Cosmos DB supports five APIs:

| API | Wire Protocol | Use Case | OpenMCF v1 |
|-----|----------------|----------|------------|
| **SQL (NoSQL)** | Cosmos DB native | Document workloads, SQL-like queries | ✅ Supported |
| **MongoDB** | MongoDB wire protocol | MongoDB migrations, driver compatibility | ✅ Supported |
| **Cassandra** | Apache Cassandra CQL | Wide-column workloads | ❌ Excluded |
| **Gremlin** | Apache TinkerPop | Graph workloads | ❌ Excluded |
| **Table** | Azure Table Storage | Key-value, legacy compatibility | ❌ Excluded |

**Rationale for SQL and MongoDB only:**

- **SQL/NoSQL (GlobalDocumentDB)** covers the majority of document and key-value workloads. It is the native API with the richest feature set (vector search, full-text search, change feed).
- **MongoDB** serves teams migrating from MongoDB or using MongoDB drivers. It is the second most common API and shares the same underlying storage model.
- **Cassandra, Gremlin, Table** represent niche workloads. Cassandra and Gremlin have different data models and provisioning patterns. Table API is legacy-oriented. Supporting these would significantly increase spec complexity for a small user base. They can be added in a future version if demand warrants.

## Consistency Levels Deep Dive

Cosmos DB offers five consistency levels with distinct latency and throughput tradeoffs:

| Level | Guarantee | Latency | Throughput | Typical Use |
|-------|-----------|---------|------------|-------------|
| **Strong** | Linearizable; reads return most recent committed write | Highest | Lowest | Financial ledgers, inventory |
| **BoundedStaleness** | Reads lag by ≤ K versions or ≤ T seconds | High | Low | Analytics with freshness bounds |
| **Session** | Read-your-writes within a session | Medium | Medium | Most web/mobile apps |
| **ConsistentPrefix** | No out-of-order reads | Medium | Medium | Event sourcing, feeds |
| **Eventual** | No ordering or freshness guarantees | Lowest | Highest | Social feeds, recommendations |

**Session** is the default and recommended for most applications because it provides read-your-writes within a client session without the cost of Strong Consistency. **Strong** is only available in single-region or single-write-region configurations. **BoundedStaleness** requires `max_interval_in_seconds` and `max_staleness_prefix`; for multi-region accounts, minimum values are 300 seconds and 100000 respectively.

## Throughput Models Comparison

| Model | Configuration | Billing | Pros | Cons |
|-------|---------------|---------|------|------|
| **Provisioned** | Fixed RU/s (400–1M+) | Predictable monthly | Predictable cost, guaranteed throughput | Over-provisioning wastes money |
| **Autoscale** | Max RU/s (1000–1M+), scales 10%–100% | Pay for actual usage within range | Handles variable traffic | Minimum 1000 RU/s floor |
| **Serverless** | `EnableServerless` capability | Pay per request | No minimum, ideal for spiky workloads | Higher per-request cost at scale |

Throughput can be provisioned at the database level (shared across containers) or at the container/collection level (dedicated). Container-level throughput overrides database-level when both are set. Serverless mode ignores all throughput fields.

## Partition Key Design Best Practices

The partition key is the single most critical design decision for Cosmos DB performance and cost:

**Choose a partition key that:**

- Has **high cardinality** (many distinct values) — e.g. `userId`, `tenantId`, `deviceId`
- **Distributes requests evenly** — avoid hot partitions
- Is **frequently used in query WHERE clauses** — enables efficient routing

**Avoid:**

- **Timestamp** — causes hot partitions as new data concentrates in one partition
- **Low-cardinality fields** — e.g. `status` with few values leads to uneven distribution
- **Frequently updated fields** — partition key cannot be changed after creation

For MongoDB API, the **shard key** is the equivalent concept. Same principles apply.

## Global Distribution and Failover Mechanics

- **geo_locations**: At least one required. The location with `failover_priority: 0` is the primary write region and should match the account `region` field.
- **automatic_failover_enabled**: When true, Azure promotes the next region in failover priority order if the write region fails. Recommended for any multi-region deployment.
- **multiple_write_locations_enabled**: Enables active-active writes. All regions accept writes; conflicts require a resolution policy. Requires `automatic_failover_enabled`. Increases cost and complexity.
- **zone_redundant**: Per-region setting. When true, replicas spread across availability zones within the region. Not all regions support zone redundancy.

## Network Security: VNet Rules, IP Firewall, Private Endpoints

**VNet rules** (`virtual_network_rules`):

- Require `is_virtual_network_filter_enabled: true`
- Each rule references a subnet; traffic from that subnet is allowed
- The subnet must have the `Microsoft.AzureCosmosDB` service endpoint enabled
- Use `AzureSubnet` output via `value_from` for infra-chart composition

**IP firewall** (`ip_range_filter`):

- CIDR ranges or individual IPv4 addresses
- Applied in addition to VNet rules
- `0.0.0.0` allows all Azure datacenter IPs (common for dev/test)
- For Azure Portal access, include: `104.42.195.92`, `40.76.54.131`, `52.176.6.30`, `52.169.50.45`, `52.187.184.26`

**Private endpoints**:

- Use `AzurePrivateEndpoint` with `private_connection_resource_id` referencing `account_id`
- Subresource names: `["Sql"]` for SQL API, `["MongoDB"]` for MongoDB API
- When `public_network_access_enabled: false`, only private endpoint traffic is allowed

## Backup Strategies: Periodic vs Continuous

| Type | Description | Cost | Use Case |
|------|-------------|------|----------|
| **Periodic** | Backups at configurable intervals (60–1440 min), retention 8–720 hours | Lower | Most workloads |
| **Continuous** | Point-in-time restore to any moment | Higher | Compliance, critical data |

**Important**: Once set to Continuous, the backup type **cannot be changed back** to Periodic. Choose carefully.

**Continuous tiers**: `Continuous7Days` (7-day restore window) or `Continuous30Days` (30-day window, higher cost).

## 80/20 Scoping Rationale

### What We Included for v1

| Feature | Rationale |
|---------|-----------|
| SQL and MongoDB APIs | Covers 80%+ of document/NoSQL workloads |
| Five consistency levels | Core differentiator; Session default is right for most apps |
| Provisioned, autoscale, serverless | All three throughput models in common use |
| Global distribution + failover | Essential for production multi-region |
| VNet rules + IP firewall | Standard network security |
| Periodic + Continuous backup | Covers DR and compliance needs |
| Database/container bundling | Account without data has no utility (DD03) |
| Partition key / shard key | Required for containers/collections |
| TTL (default_ttl) | Common for session/event data |
| MongoDB indexes | Basic index support for Mongo collections |

### Excluded Features (Deferred to v2)

| Feature | Rationale |
|---------|-----------|
| Custom indexing policy (SQL) | Azure defaults work for most; advanced tuning is niche |
| Conflict resolution policy | Only needed for multi-write; complex to expose |
| Analytical store | Separate feature for analytics workloads |
| Stored procedures, triggers, UDFs | Application-level; not infrastructure |
| RBAC (Azure AD auth) | Can enable post-deployment; key-based auth is default |
| Managed identity | Advanced auth pattern |
| Cassandra, Gremlin, Table APIs | Niche; separate resource kinds if needed |
| Free tier toggle in examples | One per subscription; edge case |

### Deliberately Hardcoded

| Setting | Value | Rationale |
|---------|-------|------------|
| EnableMongo capability | Auto-added for MongoDB kind | Required for MongoDB API; IaC adds if missing |

## Comparison with Alternatives

| Service | Strengths | Weaknesses |
|---------|-----------|------------|
| **Azure Cosmos DB** | Global distribution, five consistency levels, multi-API | Higher cost at scale; RU model learning curve |
| **AWS DynamoDB** | Simpler pricing, strong AWS integration | Single-region by default; fewer consistency options |
| **MongoDB Atlas** | Native MongoDB, rich tooling | Different deployment model; not Azure-native |
| **Firebase/Firestore** | Real-time sync, client SDKs | Limited query model; vendor lock-in |

Cosmos DB is the right choice when you need Azure-native global distribution, tunable consistency, and either SQL-like document queries or MongoDB compatibility.

## Cost Optimization Strategies

1. **Start with Session consistency** — Strong and BoundedStaleness cost more RU/s per read.
2. **Use serverless for dev/test** — No minimum RU/s; pay only for actual requests.
3. **Use autoscale for variable production** — Avoid over-provisioning; scales down to 10% of max.
4. **Choose partition keys wisely** — Hot partitions increase cost; high cardinality distributes load.
5. **Share throughput at database level** — When containers have similar traffic, database-level throughput can be cheaper than per-container.
6. **Consider free tier** — 1000 RU/s + 25 GB per subscription for one account (dev/test).
7. **Continuous backup** — Only enable when compliance or PITR is required; Periodic is cheaper.

## References to Azure Documentation

- [Azure Cosmos DB Overview](https://learn.microsoft.com/en-us/azure/cosmos-db/introduction)
- [Consistency Levels](https://learn.microsoft.com/en-us/azure/cosmos-db/consistency-levels)
- [Partitioning and Horizontal Scaling](https://learn.microsoft.com/en-us/azure/cosmos-db/partitioning-overview)
- [Request Units in Azure Cosmos DB](https://learn.microsoft.com/en-us/azure/cosmos-db/request-units)
- [Serverless](https://learn.microsoft.com/en-us/azure/cosmos-db/serverless)
- [Configure Virtual Network Service Endpoints](https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-vnet-service-endpoint)
- [Backup and Restore](https://learn.microsoft.com/en-us/azure/cosmos-db/online-backup-restore)
- [Private Endpoints for Cosmos DB](https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-private-endpoints)
