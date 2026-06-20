---
title: "Presets"
description: "Ready-to-deploy configuration presets for TCP Route"
type: "preset-list"
componentSlug: "tcp-route"
componentTitle: "TCP Route"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-tcp-port-forward"
    rank: "01"
    title: "TCP Port Forwarding"
    excerpt: "The most common TCPRoute: forward all connections arriving on a Gateway's TCP listener to a backend Service. A TCP route has no matching -- the listener's port selects the traffic, and the route..."
  - slug: "02-tcp-weighted-backends"
    rank: "02"
    title: "TCP Weighted Backends"
    excerpt: "A TCPRoute rule that splits raw TCP connections across two backends by weight -- the building block for a canary or blue/green rollout of a non-HTTP service. Connection rejections (for invalid..."
---

# TCP Route Presets

Ready-to-deploy configuration presets for TCP Route. Each preset is a complete manifest you can copy, customize, and deploy.
