---
title: "Presets"
description: "Ready-to-deploy configuration presets for Elastic File System"
type: "preset-list"
componentSlug: "elastic-file-system"
componentTitle: "Elastic File System"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-general-purpose-regional"
    rank: "01"
    title: "General Purpose Regional EFS"
    excerpt: "Regional, encrypted, bursting throughput, backup enabled, no access points. Simplest production-safe starting point."
  - slug: "02-one-zone-dev"
    rank: "02"
    title: "One Zone Dev EFS"
    excerpt: "One Zone storage (us-east-1a), encrypted, bursting throughput. Lower cost for dev/test. Single subnet."
  - slug: "03-production-elastic-with-access-points"
    rank: "03"
    title: "Production Elastic EFS with Access Points"
    excerpt: "Regional, encrypted, elastic throughput, lifecycle policies (AFTER_30_DAYS IA, AFTER_1_ACCESS primary), backup, 2 access points (app-data with uid/gid 1000, logs with uid/gid 1001)."
---

# Elastic File System Presets

Ready-to-deploy configuration presets for Elastic File System. Each preset is a complete manifest you can copy, customize, and deploy.
