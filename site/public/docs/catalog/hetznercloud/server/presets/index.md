---
title: "Presets"
description: "Ready-to-deploy configuration presets for Server"
type: "preset-list"
componentSlug: "server"
componentTitle: "Server"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-quick-start"
    rank: "01"
    title: "Quick-Start Server"
    excerpt: "This preset creates the simplest possible Hetzner Cloud server: a shared-vCPU instance running Ubuntu with SSH access and auto-assigned public IPv4 and IPv6 addresses. It provisions a single..."
  - slug: "02-production-web"
    rank: "02"
    title: "Production Web Server"
    excerpt: "This preset creates a production-hardened Hetzner Cloud server for web workloads. It combines a firewall for inbound traffic control, a private network attachment for backend communication, automatic..."
  - slug: "03-private-backend"
    rank: "03"
    title: "Private Backend Server"
    excerpt: "This preset creates a Hetzner Cloud server with no public IP address, reachable only through a private network. Both public IPv4 and IPv6 are explicitly disabled via the `publicNet` block, making the..."
---

# Server Presets

Ready-to-deploy configuration presets for Server. Each preset is a complete manifest you can copy, customize, and deploy.
