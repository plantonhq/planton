---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer"
type: "preset-list"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-https-letsencrypt"
    rank: "01"
    title: "HTTPS Load Balancer with Let's Encrypt"
    excerpt: "This preset creates a Scaleway Load Balancer with automatic TLS certificate provisioning via Let's Encrypt, HTTP health checks, and two frontends (HTTPS on 443 and HTTP on 80). The LB is attached to..."
  - slug: "02-http-simple"
    rank: "02"
    title: "Simple HTTP Load Balancer"
    excerpt: "This preset creates a minimal Scaleway Load Balancer with a single HTTP frontend and TCP health checks. No TLS certificates or Private Network attachment are configured, making this the simplest..."
---

# Load Balancer Presets

Ready-to-deploy configuration presets for Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
