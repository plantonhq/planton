---
title: "Single Instance Solr with ZooKeeper"
description: "This preset deploys a single-node SolrCloud instance with a single ZooKeeper node for coordination. Suitable for development or small search workloads."
type: "preset"
rank: "01"
presetSlug: "01-single-instance"
componentSlug: "solr"
componentTitle: "Solr"
provider: "kubernetes"
icon: "package"
order: 1
---

# Single Instance Solr with ZooKeeper

This preset deploys a single-node SolrCloud instance with a single ZooKeeper node for coordination. Suitable for development or small search workloads.

## When to Use

- Development or testing of Solr search functionality
- Small search indexes with moderate query throughput
- Evaluating SolrCloud before scaling to a production cluster

## Key Configuration Choices

- **Single Solr replica** with 10Gi disk -- stores search indexes; increase based on corpus size
- **Single ZooKeeper** with 1Gi disk -- coordinates SolrCloud; a single instance is sufficient for non-HA setups
- **Higher Solr memory** (`4Gi` limit) -- Solr uses JVM heap heavily; adjust `-Xmx` in Solr config accordingly

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
