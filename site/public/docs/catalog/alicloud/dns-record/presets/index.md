---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Record"
type: "preset-list"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-a-record"
    rank: "01"
    title: "A Record"
    excerpt: "This preset creates a DNS A record that maps a subdomain to an IPv4 address. A records are the most common DNS record type, used to point domain names at web servers, application servers, or any host..."
  - slug: "02-cname-record"
    rank: "02"
    title: "CNAME Record"
    excerpt: "This preset creates a DNS CNAME record that aliases a subdomain to another domain name. CNAME records are commonly used for CDN integration, service abstraction, and multi-environment routing."
  - slug: "03-mx-record"
    rank: "03"
    title: "MX Record"
    excerpt: "This preset creates a DNS MX (Mail Exchange) record that routes email for a domain to a mail server. MX records use a priority value to determine the order in which mail servers are tried."
---

# DNS Record Presets

Ready-to-deploy configuration presets for DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
