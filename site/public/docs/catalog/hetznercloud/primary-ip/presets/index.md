---
title: "Presets"
description: "Ready-to-deploy configuration presets for Primary IP"
type: "preset-list"
componentSlug: "primary-ip"
componentTitle: "Primary IP"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-standard-ipv4"
    rank: "01"
    title: "Standard IPv4 Primary IP"
    excerpt: "This preset allocates a persistent public IPv4 address in Hetzner Cloud's Falkenstein datacenter. The IP exists independently of any server -- it survives server deletion and can be reassigned,..."
  - slug: "02-mail-server-ipv4"
    rank: "02"
    title: "Mail Server IPv4 with Reverse DNS"
    excerpt: "This preset allocates a persistent public IPv4 address with a reverse DNS (rDNS) record and delete protection enabled. It is designed for mail servers and any service where clients verify the..."
---

# Primary IP Presets

Ready-to-deploy configuration presets for Primary IP. Each preset is a complete manifest you can copy, customize, and deploy.
