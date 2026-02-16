---
title: "Presets"
description: "Ready-to-deploy configuration presets for Floating IP"
type: "preset-list"
componentSlug: "floating-ip"
componentTitle: "Floating IP"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-allocation-only"
    rank: "01"
    title: "Floating IP Allocation Only"
    excerpt: "This preset allocates a floating IP from an external network without associating it with a port. The floating IP is reserved but not yet bound to any instance. Use `OpenStackFloatingIpAssociate` as a..."
  - slug: "02-with-port-association"
    rank: "02"
    title: "Floating IP with Port Association"
    excerpt: "This preset allocates a floating IP and immediately associates it with a port, providing external connectivity to whatever is attached to that port (typically an instance). This is a single-resource..."
---

# Floating IP Presets

Ready-to-deploy configuration presets for Floating IP. Each preset is a complete manifest you can copy, customize, and deploy.
