---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer Member"
type: "preset-list"
componentSlug: "load-balancer-member"
componentTitle: "Load Balancer Member"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Pool Member"
    excerpt: "This preset adds a backend server to an Octavia pool. Each member has an IP address and port that the pool forwards traffic to based on its load-balancing algorithm. Create one member resource per..."
---

# Load Balancer Member Presets

Ready-to-deploy configuration presets for Load Balancer Member. Each preset is a complete manifest you can copy, customize, and deploy.
