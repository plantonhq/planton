---
title: "Presets"
description: "Ready-to-deploy configuration presets for Network Port"
type: "preset-list"
componentSlug: "network-port"
componentTitle: "Network Port"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-standard-fixed-ip"
    rank: "01"
    title: "Standard Port with Fixed IP"
    excerpt: "This preset creates a port on a network with an IP auto-assigned from a specific subnet. The port gets the project's default security group applied automatically. This is the most common port..."
  - slug: "02-no-security-groups"
    rank: "02"
    title: "Port with No Security Groups"
    excerpt: "This preset creates a port with all security groups removed, including the default security group that OpenStack normally applies. Traffic flows unrestricted through this port. Use this for load..."
---

# Network Port Presets

Ready-to-deploy configuration presets for Network Port. Each preset is a complete manifest you can copy, customize, and deploy.
