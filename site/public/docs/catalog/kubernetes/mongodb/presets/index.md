---
title: "Presets"
description: "Ready-to-deploy configuration presets for MongoDB"
type: "preset-list"
componentSlug: "mongodb"
componentTitle: "MongoDB"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-single-instance"
    rank: "01"
    title: "Single Instance MongoDB"
    excerpt: "This preset deploys a single-replica MongoDB instance with persistence enabled. Suitable for development, testing, or applications that do not require replica set features."
  - slug: "02-replica-set"
    rank: "02"
    title: "MongoDB Replica Set"
    excerpt: "This preset deploys a 3-node MongoDB replica set with persistence. Provides automatic failover and read scaling for production workloads."
---

# MongoDB Presets

Ready-to-deploy configuration presets for MongoDB. Each preset is a complete manifest you can copy, customize, and deploy.
