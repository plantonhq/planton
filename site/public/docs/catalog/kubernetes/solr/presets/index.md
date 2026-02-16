---
title: "Presets"
description: "Ready-to-deploy configuration presets for Solr"
type: "preset-list"
componentSlug: "solr"
componentTitle: "Solr"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-single-instance"
    rank: "01"
    title: "Single Instance Solr with ZooKeeper"
    excerpt: "This preset deploys a single-node SolrCloud instance with a single ZooKeeper node for coordination. Suitable for development or small search workloads."
---

# Solr Presets

Ready-to-deploy configuration presets for Solr. Each preset is a complete manifest you can copy, customize, and deploy.
