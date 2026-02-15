---
title: "CNAME Record"
description: "This preset creates a standalone CNAME record aliasing a subdomain to another hostname. CNAME records are the standard way to point subdomains to load balancers, CDNs, or other services that expose a..."
type: "preset"
rank: "02"
presetSlug: "02-cname-record"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "civo"
icon: "package"
order: 2
---

# CNAME Record

This preset creates a standalone CNAME record aliasing a subdomain to another hostname. CNAME records are the standard way to point subdomains to load balancers, CDNs, or other services that expose a hostname rather than a static IP.

## When to Use

- Pointing a subdomain to a load balancer, CDN, or SaaS service hostname
- Creating aliases between subdomains (e.g., `app.example.com` -> `my-lb.civo.com`)
- Any scenario where the target is a hostname that may change its IP independently

## Key Configuration Choices

- **CNAME record type** (`type: CNAME`) -- creates a hostname-to-hostname alias; the target's IP is resolved at query time
- **Subdomain** (`name: app`) -- the hostname relative to the zone
- **1-hour TTL** (`ttl: 3600`) -- standard caching duration

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<dns-zone-id>` | ID of the target CivoDnsZone | `CivoDnsZone` status outputs |
| `app` | Subdomain name (relative to zone) | Your DNS naming plan |
| `<target-hostname.example.com>` | Target hostname (FQDN) | Load balancer, CDN, or service provider |

## Related Presets

- **01-a-record** -- Use instead when the target is a static IPv4 address rather than a hostname
