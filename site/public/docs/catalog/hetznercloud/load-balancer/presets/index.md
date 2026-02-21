---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer"
type: "preset-list"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-https-web-app"
    rank: "01"
    title: "HTTPS Web Application"
    excerpt: "This preset creates a public-facing HTTPS load balancer that terminates TLS at the edge and forwards plain HTTP to backend servers. It automatically redirects all HTTP traffic to HTTPS, discovers..."
  - slug: "02-private-internal"
    rank: "02"
    title: "Private Internal Load Balancer"
    excerpt: "This preset creates a load balancer attached to a Hetzner Cloud private network with its public interface disabled. It distributes HTTP traffic across explicit server targets using their private IPs,..."
  - slug: "03-tcp-pass-through"
    rank: "03"
    title: "TCP Pass-Through"
    excerpt: "This preset creates a layer-4 TCP load balancer that forwards raw TCP connections to backend servers without any application-layer inspection. The load balancer does not parse HTTP headers, manage..."
---

# Load Balancer Presets

Ready-to-deploy configuration presets for Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
