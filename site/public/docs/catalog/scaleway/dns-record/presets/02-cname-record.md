---
title: "CNAME Record"
description: "This preset creates a DNS CNAME record that maps a hostname to another hostname. CNAMEs are used to create aliases -- for example, pointing `www.example.com` to the canonical hostname of an..."
type: "preset"
rank: "02"
presetSlug: "02-cname-record"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "scaleway"
icon: "package"
order: 2
---

# CNAME Record

This preset creates a DNS CNAME record that maps a hostname to another hostname. CNAMEs are used to create aliases -- for example, pointing `www.example.com` to the canonical hostname of an application, CDN, or external service.

## When to Use

- Creating a `www` alias that points to the root domain or a CDN endpoint
- Pointing a subdomain to a Kapsule cluster's wildcard DNS endpoint
- Aliasing service subdomains to external providers (e.g., `docs` to a hosted documentation platform)

## Key Configuration Choices

- **CNAME record type** (`type: CNAME`) -- creates an alias from one hostname to another
- **Subdomain name** (`name: www`) -- creates `www.<your-zone>` as the alias; CNAME records cannot be used at the zone apex (use ALIAS type instead)
- **Trailing dot on target** -- DNS convention; the target hostname should end with a dot to indicate a fully qualified domain name
- **1-hour TTL** (`ttl: 3600`) -- standard caching duration; suitable for most CNAME use cases
- **Keep empty zone** (`keepEmptyZone: true`) -- prevents accidental zone deletion when this is the last record being destroyed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-dns-zone-name>` | DNS zone name (e.g., `example.com`) | Scaleway console or `ScalewayDnsZone` status outputs |
| `<your-target-hostname.>` | Target hostname with trailing dot (e.g., `app.example.com.`) | The canonical hostname of your service or CDN |

## Related Presets

- **01-a-record** -- Use instead when pointing a hostname directly to an IPv4 address
