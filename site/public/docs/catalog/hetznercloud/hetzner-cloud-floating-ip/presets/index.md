---
title: "Presets"
description: "Ready-to-deploy configuration presets for Hetzner Cloud Floating IP"
type: "preset-list"
componentSlug: "hetzner-cloud-floating-ip"
componentTitle: "Hetzner Cloud Floating IP"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-reserved-ipv4"
    rank: "01"
    title: "Reserved IPv4 Floating IP"
    excerpt: "This preset allocates an unassigned IPv4 Floating IP in Hetzner Cloud. The IP is reserved but not attached to any server, making it available for later assignment when your failover infrastructure..."
  - slug: "02-failover-ipv4"
    rank: "02"
    title: "Failover IPv4 Floating IP"
    excerpt: "This preset allocates an IPv4 Floating IP and assigns it to a server, with delete protection enabled. It represents the standard production failover configuration where a stable public endpoint must..."
  - slug: "03-mail-failover-ipv4"
    rank: "03"
    title: "Mail Failover IPv4 with Reverse DNS"
    excerpt: "This preset allocates an IPv4 Floating IP with a reverse DNS (rDNS) record, assigns it to a server, and enables delete protection. It is designed for mail servers that need both failover capability..."
---

# Hetzner Cloud Floating IP Presets

Ready-to-deploy configuration presets for Hetzner Cloud Floating IP. Each preset is a complete manifest you can copy, customize, and deploy.
