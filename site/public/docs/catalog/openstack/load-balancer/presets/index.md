---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer"
type: "preset-list"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Load Balancer"
    excerpt: "This preset creates an Octavia load balancer with a VIP on the specified subnet. The load balancer itself is just the VIP endpoint -- attach listeners, pools, members, and monitors to complete the..."
---

# Load Balancer Presets

Ready-to-deploy configuration presets for Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
