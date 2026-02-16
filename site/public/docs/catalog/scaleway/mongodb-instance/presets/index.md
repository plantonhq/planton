---
title: "Presets"
description: "Ready-to-deploy configuration presets for MongoDB Instance"
type: "preset-list"
componentSlug: "mongodb-instance"
componentTitle: "MongoDB Instance"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-dev-standalone"
    rank: "01"
    title: "Development Standalone MongoDB"
    excerpt: "This preset creates a single-node Scaleway MongoDB instance using the smallest available node type. It is the most affordable path to a working MongoDB database for development, testing, and..."
  - slug: "02-production-replica-set"
    rank: "02"
    title: "Production MongoDB Replica Set"
    excerpt: "This preset creates a 3-node Scaleway MongoDB replica set with Private Network connectivity and automated snapshot scheduling. The replica set provides automatic failover -- if the primary node..."
---

# MongoDB Instance Presets

Ready-to-deploy configuration presets for MongoDB Instance. Each preset is a complete manifest you can copy, customize, and deploy.
