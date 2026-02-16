---
title: "Presets"
description: "Ready-to-deploy configuration presets for Server Group"
type: "preset-list"
componentSlug: "server-group"
componentTitle: "Server Group"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-anti-affinity"
    rank: "01"
    title: "Anti-Affinity Server Group"
    excerpt: "This preset creates a server group with the anti-affinity policy. Instances placed in this group are scheduled on different physical hypervisors, maximizing fault tolerance. If a hypervisor fails,..."
  - slug: "02-affinity"
    rank: "02"
    title: "Affinity Server Group"
    excerpt: "This preset creates a server group with the affinity policy. Instances placed in this group are scheduled on the same physical hypervisor, minimizing network latency between them. Use this when..."
---

# Server Group Presets

Ready-to-deploy configuration presets for Server Group. Each preset is a complete manifest you can copy, customize, and deploy.
