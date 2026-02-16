---
title: "Production Elasticsearch Cluster"
description: "This preset deploys a 3-node Elasticsearch cluster with Kibana for production workloads. The cluster provides data replication, shard distribution, and fault tolerance across nodes."
type: "preset"
rank: "03"
presetSlug: "03-production-cluster"
componentSlug: "elasticsearch"
componentTitle: "Elasticsearch"
provider: "kubernetes"
icon: "package"
order: 3
---

# Production Elasticsearch Cluster

This preset deploys a 3-node Elasticsearch cluster with Kibana for production workloads. The cluster provides data replication, shard distribution, and fault tolerance across nodes.

## When to Use

- Production log aggregation, full-text search, or analytics
- Workloads requiring data replication and shard distribution for resilience
- High-throughput indexing or query-heavy environments

## Key Configuration Choices

- **3 Elasticsearch nodes** with 50Gi disk each -- enables primary and replica shards across nodes; tolerates 1 node failure
- **High memory** (`1Gi` request, `8Gi` limit) -- Elasticsearch relies on JVM heap and OS filesystem cache; more memory = better query performance
- **Kibana with ingress** -- web UI for dashboards, search, and index management
- **Persistence enabled** -- index data survives node restarts

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
| `<your-kibana.example.com>` | Hostname for Kibana ingress access | Your DNS provider |

## Related Presets

- **01-single-node** -- Minimal single-node Elasticsearch for development
- **02-with-kibana** -- Single-node with Kibana for staging
