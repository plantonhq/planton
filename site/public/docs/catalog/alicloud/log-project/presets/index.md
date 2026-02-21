---
title: "Presets"
description: "Ready-to-deploy configuration presets for Log Project"
type: "preset-list"
componentSlug: "log-project"
componentTitle: "Log Project"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-production-multi-store"
    rank: "01"
    title: "Production Multi-Store Log Project"
    excerpt: "This preset creates a production-ready Alibaba Cloud SLS project with three purpose-specific log stores: application logs (90-day retention), audit logs (365-day retention), and access logs (30-day..."
  - slug: "02-development"
    rank: "02"
    title: "Development Log Project"
    excerpt: "This preset creates a minimal SLS project with a single log store using 7-day retention and one shard. This is the lowest-cost configuration for development and testing environments where log volume..."
---

# Log Project Presets

Ready-to-deploy configuration presets for Log Project. Each preset is a complete manifest you can copy, customize, and deploy.
