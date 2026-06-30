---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer Monitor"
type: "preset-list"
componentSlug: "load-balancer-monitor"
componentTitle: "Load Balancer Monitor"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-https-web"
    rank: "01"
    title: "Preset: HTTPS web health check"
    excerpt: "An HTTPS monitor that probes `/healthz` on each origin and expects a 2xx response, suitable for a web/API pool behind a Cloudflare Load Balancer."
  - slug: "02-tcp-port"
    rank: "02"
    title: "Preset: TCP port health check"
    excerpt: "A TCP monitor that checks whether a port accepts connections — suitable for non-HTTP origins (databases, message brokers, custom TCP services) behind a Cloudflare Spectrum / TCP load balancer."
---

# Load Balancer Monitor Presets

Ready-to-deploy configuration presets for Load Balancer Monitor. Each preset is a complete manifest you can copy, customize, and deploy.
