---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer"
type: "preset-list"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-https-ssl-termination"
    rank: "01"
    title: "HTTPS Load Balancer with SSL Termination"
    excerpt: "This preset creates a load balancer that terminates TLS on port 443 and forwards traffic to backend Droplets over HTTP on port 80. Uses tag-based targeting so any Droplet with the `web` tag in the..."
  - slug: "02-http-basic"
    rank: "02"
    title: "Simple HTTP Load Balancer"
    excerpt: "This preset creates a basic HTTP load balancer that forwards port 80 traffic to backend Droplets on port 8080. Uses explicit droplet IDs for targeting. Health checks ensure only healthy backends..."
---

# Load Balancer Presets

Ready-to-deploy configuration presets for Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
