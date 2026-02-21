---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Zone"
type: "preset-list"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-public-primary"
    rank: "01"
    title: "Public Primary DNS Zone"
    excerpt: "This preset creates a publicly resolvable, authoritative DNS zone hosted on OCI's managed DNS service. The zone is configured as PRIMARY (OCI is the source of truth for all records) with GLOBAL scope..."
  - slug: "02-private-vcn"
    rank: "02"
    title: "Private VCN DNS Zone"
    excerpt: "This preset creates a private DNS zone resolvable only within VCNs attached to the specified DNS view. Private zones enable internal service discovery without exposing hostnames to the public..."
---

# DNS Zone Presets

Ready-to-deploy configuration presets for DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
