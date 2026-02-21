---
title: "Presets"
description: "Ready-to-deploy configuration presets for Public IP"
type: "preset-list"
componentSlug: "public-ip"
componentTitle: "Public IP"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-reserved-unassigned"
    rank: "01"
    title: "Reserved Unassigned Public IP"
    excerpt: "This preset allocates a reserved public IP without assigning it to any resource. The IP persists independently of any compute instance or load balancer, making it suitable for pre-provisioning stable..."
  - slug: "02-reserved-assigned"
    rank: "02"
    title: "Reserved Assigned Public IP"
    excerpt: "This preset allocates a reserved public IP and immediately assigns it to an existing private IP on a VNIC. The IP persists across instance reboots and can be reassigned to a different private IP..."
---

# Public IP Presets

Ready-to-deploy configuration presets for Public IP. Each preset is a complete manifest you can copy, customize, and deploy.
