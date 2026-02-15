---
title: "Presets"
description: "Ready-to-deploy configuration presets for Route53 DNS Record"
type: "preset-list"
componentSlug: "route53-dns-record"
componentTitle: "Route53 DNS Record"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-alias-alb"
    rank: "01"
    title: "Alias Record to ALB"
    excerpt: "This preset creates a Route53 alias record pointing to an Application Load Balancer. Alias records are Route53's most powerful feature -- they work at the zone apex (e.g., `example.com`), incur no..."
  - slug: "02-a-record"
    rank: "02"
    title: "Simple A Record"
    excerpt: "This preset creates a standard A record that maps a domain name to one or more IPv4 addresses. This is the most basic DNS record type, used for pointing domains to servers, load balancers, or any..."
---

# Route53 DNS Record Presets

Ready-to-deploy configuration presets for Route53 DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
