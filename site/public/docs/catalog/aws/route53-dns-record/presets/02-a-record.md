---
title: "Simple A Record"
description: "This preset creates a standard A record that maps a domain name to one or more IPv4 addresses. This is the most basic DNS record type, used for pointing domains to servers, load balancers, or any..."
type: "preset"
rank: "02"
presetSlug: "02-a-record"
componentSlug: "route53-dns-record"
componentTitle: "Route53 DNS Record"
provider: "aws"
icon: "package"
order: 2
---

# Simple A Record

This preset creates a standard A record that maps a domain name to one or more IPv4 addresses. This is the most basic DNS record type, used for pointing domains to servers, load balancers, or any resource with a static IP address. Uses a 5-minute TTL that balances caching efficiency with change propagation speed.

## When to Use

- Pointing a subdomain to a server with a known, static IPv4 address
- External (non-AWS) resources where alias records are not applicable
- Simple DNS mappings where advanced routing policies are not needed

## Key Configuration Choices

- **A record type** (`type: A`) -- Maps domain name to IPv4 address(es)
- **5-minute TTL** (`ttl: 300`) -- Standard DNS caching duration; reduce to 60 seconds during planned changes for faster propagation
- **Simple routing** -- No routing policy specified; uses standard DNS behavior (returns all values, client chooses)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<route53-hosted-zone-id>` | ID of your Route53 hosted zone | AWS Route53 console or `AwsRoute53Zone` status outputs |
| `<your-subdomain.your-domain.com>` | Fully qualified domain name (e.g., `api.example.com`) | Your domain naming convention |
| `<ipv4-address>` | Target IPv4 address (e.g., `203.0.113.50`); add more entries for round-robin | Your server or infrastructure provider |

## Related Presets

- **01-alias-alb** -- Use instead when pointing to an AWS ALB (free queries, automatic IP tracking, works at zone apex)
