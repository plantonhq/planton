---
title: "Presets"
description: "Ready-to-deploy configuration presets for Neptune Cluster"
type: "preset-list"
componentSlug: "neptune-cluster"
componentTitle: "Neptune Cluster"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-graph-database"
    rank: "01"
    title: "Graph Database (Standard Provisioned)"
    excerpt: "This preset creates a standard provisioned Neptune cluster with a single instance (db.r6g.large). Ideal for development, testing, or moderate graph workloads that need Gremlin or SPARQL support..."
  - slug: "02-high-availability"
    rank: "02"
    title: "High Availability Neptune Cluster"
    excerpt: "This preset creates a highly available Neptune cluster with 2 instances (1 primary writer + 1 read replica) across Availability Zones. IAM database authentication is enabled, deletion protection is..."
  - slug: "03-serverless-v2"
    rank: "03"
    title: "Neptune Serverless v2"
    excerpt: "This preset creates a Neptune cluster with Serverless v2 auto-scaling, where compute capacity automatically adjusts between 1.0 and 16.0 Neptune Capacity Units (NCUs) based on workload demand. Ideal..."
---

# Neptune Cluster Presets

Ready-to-deploy configuration presets for Neptune Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
