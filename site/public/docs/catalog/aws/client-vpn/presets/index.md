---
title: "Presets"
description: "Ready-to-deploy configuration presets for Client VPN"
type: "preset-list"
componentSlug: "client-vpn"
componentTitle: "Client VPN"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-certificate-split-tunnel"
    rank: "01"
    title: "Certificate-Based Split-Tunnel VPN"
    excerpt: "This preset creates an AWS Client VPN endpoint using mutual TLS certificate authentication with split-tunnel routing. Only traffic destined for the VPC flows through the VPN -- all other internet..."
  - slug: "02-certificate-full-tunnel"
    rank: "02"
    title: "Certificate-Based Full-Tunnel VPN"
    excerpt: "This preset creates an AWS Client VPN endpoint with full-tunnel routing, where all client traffic -- including internet traffic -- routes through the VPN. This provides complete network control and..."
---

# Client VPN Presets

Ready-to-deploy configuration presets for Client VPN. Each preset is a complete manifest you can copy, customize, and deploy.
