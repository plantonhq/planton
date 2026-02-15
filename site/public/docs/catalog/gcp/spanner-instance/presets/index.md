---
title: "Presets"
description: "Ready-to-deploy configuration presets for Spanner Instance"
type: "preset-list"
componentSlug: "spanner-instance"
componentTitle: "Spanner Instance"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-free-instance"
    rank: "01"
    title: "Free Instance"
    excerpt: "This preset provisions a zero-cost Cloud Spanner instance using the FREE_INSTANCE type. It is ideal for development, prototyping, CI/CD testing, and learning Spanner without incurring any charges."
  - slug: "02-regional-production"
    rank: "02"
    title: "Regional Production"
    excerpt: "This preset provisions a production-ready Cloud Spanner instance with a single node, ENTERPRISE edition, and automatic backup scheduling. It is suitable for production workloads with predictable..."
  - slug: "03-autoscaling-production"
    rank: "03"
    title: "Autoscaling Production"
    excerpt: "This preset provisions a production Cloud Spanner instance with autoscaling enabled. Spanner automatically adjusts compute capacity between 1 and 3 nodes based on CPU and storage utilization targets...."
---

# Spanner Instance Presets

Ready-to-deploy configuration presets for Spanner Instance. Each preset is a complete manifest you can copy, customize, and deploy.
