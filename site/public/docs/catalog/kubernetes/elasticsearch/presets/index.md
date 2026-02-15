---
title: "Presets"
description: "Ready-to-deploy configuration presets for Elasticsearch"
type: "preset-list"
componentSlug: "elasticsearch"
componentTitle: "Elasticsearch"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-single-node"
    rank: "01"
    title: "Single Node Elasticsearch"
    excerpt: "This preset deploys a single-node Elasticsearch instance with persistence. No Kibana. Suitable for development, testing, or small log/search workloads."
  - slug: "02-with-kibana"
    rank: "02"
    title: "Elasticsearch with Kibana"
    excerpt: "This preset deploys a single-node Elasticsearch instance with Kibana enabled and exposed via ingress. Kibana provides a web UI for searching, visualizing, and dashboarding Elasticsearch data."
  - slug: "03-production-cluster"
    rank: "03"
    title: "Production Elasticsearch Cluster"
    excerpt: "This preset deploys a 3-node Elasticsearch cluster with Kibana for production workloads. The cluster provides data replication, shard distribution, and fault tolerance across nodes."
---

# Elasticsearch Presets

Ready-to-deploy configuration presets for Elasticsearch. Each preset is a complete manifest you can copy, customize, and deploy.
