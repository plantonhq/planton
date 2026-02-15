---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer Pool"
type: "preset-list"
componentSlug: "load-balancer-pool"
componentTitle: "Load Balancer Pool"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-round-robin"
    rank: "01"
    title: "Round-Robin HTTP Pool"
    excerpt: "This preset creates a backend pool using the round-robin algorithm over HTTP. Traffic is distributed equally across all healthy members. This is the most common pool configuration for stateless web..."
  - slug: "02-sticky-session"
    rank: "02"
    title: "Sticky Session Pool (HTTP Cookie)"
    excerpt: "This preset creates a backend pool with round-robin distribution and HTTP cookie-based session persistence. Octavia inserts and tracks a cookie so that subsequent requests from the same client are..."
---

# Load Balancer Pool Presets

Ready-to-deploy configuration presets for Load Balancer Pool. Each preset is a complete manifest you can copy, customize, and deploy.
