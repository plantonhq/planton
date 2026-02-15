---
title: "Presets"
description: "Ready-to-deploy configuration presets for Global Address"
type: "preset-list"
componentSlug: "global-address"
componentTitle: "Global Address"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-external-static-ip"
    rank: "01"
    title: "External Static IP"
    excerpt: "This preset reserves a global external IPv4 address for use with HTTP(S) load balancers, Cloud CDN, or global forwarding rules. It is the simplest GcpGlobalAddress configuration — just a project, a..."
  - slug: "02-internal-vpc-peering-range"
    rank: "02"
    title: "Internal VPC Peering Range"
    excerpt: "This preset reserves a `/20` internal IP range for VPC peering with Google managed services. Cloud SQL, Memorystore (Redis), AlloyDB, and Filestore all use this mechanism to assign private IPs from..."
  - slug: "03-private-service-connect"
    rank: "03"
    title: "Private Service Connect Endpoint Address"
    excerpt: "This preset reserves a global internal IP address for a Private Service Connect (PSC) endpoint. PSC allows private, secure connectivity to Google APIs, Google managed services, or third-party..."
---

# Global Address Presets

Ready-to-deploy configuration presets for Global Address. Each preset is a complete manifest you can copy, customize, and deploy.
