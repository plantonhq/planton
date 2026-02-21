---
title: "Presets"
description: "Ready-to-deploy configuration presets for Hetzner Cloud Firewall"
type: "preset-list"
componentSlug: "hetzner-cloud-firewall"
componentTitle: "Hetzner Cloud Firewall"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-web-server"
    rank: "01"
    title: "Web Server Firewall"
    excerpt: "This preset creates a firewall for public-facing web servers, allowing inbound SSH, HTTP, HTTPS, and ICMP from all IPv4 and IPv6 addresses. It covers the most common Hetzner Cloud server deployment:..."
  - slug: "02-ssh-only"
    rank: "02"
    title: "SSH-Only Firewall"
    excerpt: "This preset creates a minimal-surface firewall that allows only SSH and ICMP inbound. All other inbound traffic is dropped by Hetzner Cloud's deny-by-default policy. This is the tightest useful..."
  - slug: "03-private-network"
    rank: "03"
    title: "Private Network Firewall"
    excerpt: "This preset creates a firewall that restricts all inbound traffic to a single private CIDR range, blocking all access from the public internet. Only SSH and ICMP are permitted, and only from hosts..."
---

# Hetzner Cloud Firewall Presets

Ready-to-deploy configuration presets for Hetzner Cloud Firewall. Each preset is a complete manifest you can copy, customize, and deploy.
