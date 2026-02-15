---
title: "Presets"
description: "Ready-to-deploy configuration presets for DocumentDB"
type: "preset-list"
componentSlug: "documentdb"
componentTitle: "DocumentDB"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-production-ha"
    rank: "01"
    title: "Production HA DocumentDB Cluster"
    excerpt: "This preset creates a highly available DocumentDB cluster with 3 instances (1 primary + 2 replicas) across Availability Zones. Storage is encrypted, backups are retained for 7 days, and deletion..."
  - slug: "02-development"
    rank: "02"
    title: "Development DocumentDB Cluster"
    excerpt: "This preset creates a single-instance DocumentDB cluster for development and testing. It uses a smaller instance class (db.t3.medium) and skips the final snapshot on deletion to simplify teardown...."
---

# DocumentDB Presets

Ready-to-deploy configuration presets for DocumentDB. Each preset is a complete manifest you can copy, customize, and deploy.
