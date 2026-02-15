---
title: "Presets"
description: "Ready-to-deploy configuration presets for Floating IP Associate"
type: "preset-list"
componentSlug: "floating-ip-associate"
componentTitle: "Floating IP Associate"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Floating IP Association"
    excerpt: "This preset binds an existing floating IP to a port. It is the \"join\" resource that connects a pre-allocated floating IP (from `OpenStackFloatingIp`) to a port (from `OpenStackNetworkPort`),..."
---

# Floating IP Associate Presets

Ready-to-deploy configuration presets for Floating IP Associate. Each preset is a complete manifest you can copy, customize, and deploy.
