---
title: "Presets"
description: "Ready-to-deploy configuration presets for NLB Load Balancer"
type: "preset-list"
componentSlug: "nlb-load-balancer"
componentTitle: "NLB Load Balancer"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-internet-tcp"
    rank: "01"
    title: "Preset: Internet-Facing TCP"
    excerpt: "- You need a public-facing L4 load balancer for TCP traffic. - Simple setup: one server group, one listener, two availability zones. - Suitable for development, staging, or simple TCP services."
  - slug: "02-internal-tcp-drain"
    rank: "02"
    title: "Preset: Internal TCP with Connection Draining"
    excerpt: "- Internal (VPC-private) L4 load balancer for microservice traffic. - Graceful deployments where in-flight connections must complete before backends are removed. - Source-IP consistent hashing for..."
  - slug: "03-tcpssl-production"
    rank: "03"
    title: "Preset: TCPSSL Production"
    excerpt: "- Production internet-facing NLB with TLS termination at Layer 4. - Fixed public IPs via EIP binding for DNS A-records and firewall whitelisting. - Database or API traffic that needs encryption..."
---

# NLB Load Balancer Presets

Ready-to-deploy configuration presets for NLB Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
