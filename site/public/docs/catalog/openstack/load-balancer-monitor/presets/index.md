---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer Monitor"
type: "preset-list"
componentSlug: "load-balancer-monitor"
componentTitle: "Load Balancer Monitor"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-http-health-check"
    rank: "01"
    title: "HTTP Health Check Monitor"
    excerpt: "This preset creates an HTTP health monitor that checks backend members by sending a GET request to `/healthz` every 10 seconds. A member is considered healthy after 3 consecutive successful responses..."
  - slug: "02-tcp-health-check"
    rank: "02"
    title: "TCP Health Check Monitor"
    excerpt: "This preset creates a TCP health monitor that checks backend members by attempting a TCP connection every 10 seconds. If the connection succeeds, the member is healthy. This is the standard monitor..."
---

# Load Balancer Monitor Presets

Ready-to-deploy configuration presets for Load Balancer Monitor. Each preset is a complete manifest you can copy, customize, and deploy.
