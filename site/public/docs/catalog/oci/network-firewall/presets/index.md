---
title: "Presets"
description: "Ready-to-deploy configuration presets for Network Firewall"
type: "preset-list"
componentSlug: "network-firewall"
componentTitle: "Network Firewall"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-web-perimeter"
    rank: "01"
    title: "Web Perimeter Firewall"
    excerpt: "This preset creates an OCI Network Firewall with a policy designed for a standard web-facing perimeter. Inbound HTTP and HTTPS traffic from the internet is allowed to reach internal web servers, all..."
  - slug: "02-ids-with-url-filtering"
    rank: "02"
    title: "IDS-Enabled Firewall with URL Filtering"
    excerpt: "This preset creates an OCI Network Firewall with a policy that combines L4 traffic control, L7 URL-based filtering, and intrusion detection. Malicious URLs are rejected before reaching application..."
---

# Network Firewall Presets

Ready-to-deploy configuration presets for Network Firewall. Each preset is a complete manifest you can copy, customize, and deploy.
