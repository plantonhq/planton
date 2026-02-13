# AzureCosmosdbAccount

**Azure Cosmos DB Account** with bundled SQL/NoSQL databases and containers, or MongoDB databases and collections.

## Overview

`AzureCosmosdbAccount` provisions an Azure Cosmos DB account — a globally distributed, multi-model database service designed for low-latency, elastic scalability, and tunable consistency. Unlike relational databases, Cosmos DB uses Request Units (RU/s) for throughput, supports five consistency levels, and can span multiple regions with automatic failover. This component bundles the account with its databases and containers/collections because an account without at least one data container has no storage utility.

## Key Features

- **Dual API modes**: SQL/NoSQL (GlobalDocumentDB) and MongoDB via the `kind` field
- **Five consistency levels**: Strong, BoundedStaleness, Session, ConsistentPrefix, Eventual
- **Global distribution**: Multi-region with automatic failover and optional zone redundancy
- **Throughput models**: Provisioned RU/s, autoscale (10%–100% of max), or serverless (pay-per-request)
- **Network security**: VNet rules, IP firewall, and private endpoint support
- **Backup policies**: Periodic (configurable interval/retention) or Continuous (point-in-time restore)
- **Free tier**: 1000 RU/s and 25 GB per subscription (one account per subscription)

## API Modes

| kind | Use Case | Sub-Resources |
|------|----------|----------------|
| **GlobalDocumentDB** (default) | SQL-like queries over JSON documents | `sql_databases` → `containers` |
| **MongoDB** | MongoDB wire-protocol compatible | `mongo_databases` → `collections` |

Choose `GlobalDocumentDB` for document workloads using Cosmos DB SDK or SQL-like queries. Choose `MongoDB` for applications using MongoDB drivers or migrating from MongoDB.

## When to Use This Component

- **Document and key-value workloads** requiring single-digit millisecond latency at any scale
- **Globally distributed applications** needing reads and writes from multiple regions
- **MongoDB migrations** to Azure without application code changes
- **Serverless or variable traffic** where pay-per-request or autoscale fits better than fixed RU/s
- **Multi-tenant SaaS** with partition keys like `tenantId` or `userId`

## Consistency Levels

| Level | Description | Latency | Throughput | Use Case |
|-------|-------------|---------|------------|----------|
| **Strong** | Linearizable reads | Highest | Lowest | Financial transactions |
| **BoundedStaleness** | Reads lag by at most K versions or T seconds | High | Low | Analytics with freshness bounds |
| **Session** (default) | Read-your-writes within a session | Medium | Medium | Most applications |
| **ConsistentPrefix** | No out-of-order reads | Medium | Medium | Event sourcing |
| **Eventual** | No ordering guarantees | Lowest | Highest | Social feeds, recommendations |

## Throughput Models

| Model | Configuration | Billing | Best For |
|-------|---------------|---------|----------|
| **Provisioned** | `throughput` (400+ RU/s) | Fixed monthly cost | Steady, predictable traffic |
| **Autoscale** | `autoscale_max_throughput` (1000+ RU/s) | Scales 10%–100% of max | Variable traffic with ceilings |
| **Serverless** | `capabilities: ["EnableServerless"]` | Pay per request | Spiky, intermittent workloads |

Throughput can be set at the database level (shared) or container/collection level (dedicated). Container-level throughput overrides database-level when both are set.

## Key Configuration Options

| Field | Description | Default |
|-------|-------------|---------|
| `kind` | API mode: `GlobalDocumentDB` or `MongoDB` | `GlobalDocumentDB` |
| `consistency_policy` | Consistency level and BoundedStaleness params | Session |
| `geo_locations` | Regions with failover priority (at least one required) | — |
| `capabilities` | e.g. `EnableServerless`, `EnableMongo` | — |
| `automatic_failover_enabled` | Auto-promote next region on failure | `false` |
| `multiple_write_locations_enabled` | Active-active multi-region writes | `false` |
| `is_virtual_network_filter_enabled` | Restrict to VNet rules only | `false` |
| `backup` | Periodic or Continuous backup policy | Azure default (Periodic) |

## Related Resources

- **AzurePrivateEndpoint** — Private connectivity (subresource_names: `["Sql"]` or `["MongoDB"]`)
- **AzureSubnet** — For `virtual_network_rules` (subnet must have `Microsoft.AzureCosmosDB` service endpoint)
- **AzureResourceGroup** — Container for the account

## ForceNew Fields

Changing these fields destroys and recreates the account (data loss risk):

- `name` — Account endpoint hostname
- `kind` — API mode (GlobalDocumentDB vs MongoDB)
- `free_tier_enabled` — Free tier eligibility

## Further Reading

- [examples.md](./examples.md) — Copy-paste YAML examples
- [docs/README.md](./docs/README.md) — Deployment landscape and design rationale
