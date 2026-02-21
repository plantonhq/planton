---
title: "Presets"
description: "Ready-to-deploy configuration presets for Bastion"
type: "preset-list"
componentSlug: "bastion"
componentTitle: "Bastion"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-standard-ssh-gateway"
    rank: "01"
    title: "Standard SSH Gateway"
    excerpt: "This preset creates an OCI Bastion with a client CIDR allow list and 3-hour maximum session TTL. The bastion provides secure, time-limited SSH access to compute instances and other resources in a..."
  - slug: "02-dns-proxy-enabled"
    rank: "02"
    title: "DNS Proxy Enabled"
    excerpt: "This preset creates an OCI Bastion with DNS proxy and SOCKS5 support enabled, allowing sessions to target resources using fully qualified domain names (FQDNs) instead of IP addresses. This is..."
---

# Bastion Presets

Ready-to-deploy configuration presets for Bastion. Each preset is a complete manifest you can copy, customize, and deploy.
