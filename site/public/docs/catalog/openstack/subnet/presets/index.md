---
title: "Presets"
description: "Ready-to-deploy configuration presets for Subnet"
type: "preset-list"
componentSlug: "subnet"
componentTitle: "Subnet"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-standard-dhcp"
    rank: "01"
    title: "Standard DHCP Subnet"
    excerpt: "This preset creates an IPv4 subnet with DHCP enabled and Google public DNS servers. OpenStack automatically assigns the first usable IP as the gateway and allocates the remaining range via DHCP. This..."
  - slug: "02-isolated-no-gateway"
    rank: "02"
    title: "Isolated Subnet (No Gateway)"
    excerpt: "This preset creates an isolated subnet with no gateway and no DHCP. It is designed for backend networks where instances use statically assigned IPs and no routing to external networks is needed --..."
---

# Subnet Presets

Ready-to-deploy configuration presets for Subnet. Each preset is a complete manifest you can copy, customize, and deploy.
