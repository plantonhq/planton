---
title: "Presets"
description: "Ready-to-deploy configuration presets for Neo4j"
type: "preset-list"
componentSlug: "neo4j"
componentTitle: "Neo4j"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-single-instance"
    rank: "01"
    title: "Single Instance Neo4j"
    excerpt: "This preset deploys a single-replica Neo4j graph database with persistence. Neo4j is memory-intensive, so this preset allocates higher memory than typical database defaults."
---

# Neo4j Presets

Ready-to-deploy configuration presets for Neo4j. Each preset is a complete manifest you can copy, customize, and deploy.
