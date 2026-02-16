---
title: "Presets"
description: "Ready-to-deploy configuration presets for Bigtable Instance"
type: "preset-list"
componentSlug: "bigtable-instance"
componentTitle: "Bigtable Instance"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-dev-single-node"
    rank: "01"
    title: "Dev Single Node"
    excerpt: "Minimal development instance with a single auto-allocated cluster. Deletion protection disabled for easy teardown."
  - slug: "02-ha-production"
    rank: "02"
    title: "HA Production"
    excerpt: "Multi-cluster production instance with autoscaling and automatic replication across two zones."
  - slug: "03-enterprise-encrypted"
    rank: "03"
    title: "Enterprise Encrypted"
    excerpt: "Multi-cluster production instance with CMEK encryption, aggressive autoscaling, and storage utilization targets."
---

# Bigtable Instance Presets

Ready-to-deploy configuration presets for Bigtable Instance. Each preset is a complete manifest you can copy, customize, and deploy.
