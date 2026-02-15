---
title: "CNAME Record"
description: "This preset creates a CNAME record that aliases a subdomain to another hostname. Common for `www` pointing to the root domain or a CDN, or for subdomains pointing to external services. The target..."
type: "preset"
rank: "02"
presetSlug: "02-cname-record"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "digitalocean"
icon: "package"
order: 2
---

# CNAME Record

This preset creates a CNAME record that aliases a subdomain to another hostname. Common for `www` pointing to the root domain or a CDN, or for subdomains pointing to external services. The target must be a hostname (FQDN), not an IP.

## When to Use

- `www.example.com` → `example.com` or a CDN/origin hostname
- Subdomain alias to another domain (e.g., `blog` → `blog.platform.com`)
- Pointing to third-party services (e.g., `mail` → `mail.provider.com`)
- Cannot use CNAME on root domain (`@`) on DigitalOcean—use ALIAS or A record instead

## Key Configuration Choices

- **Type CNAME** (`type: CNAME`) -- canonical name alias; target must be hostname.
- **Subdomain** (`name: www`) -- `www` is common; use any subdomain (e.g., `api`, `app`).
- **Target hostname** (`value`) -- FQDN of target (e.g., `example.com`, `lb-12345.nyc3.digitaloceanspaces.com`).
- **TTL 3600** (`ttlSeconds: 3600`) -- 1-hour cache.
- **Domain** (`domain`) -- DNS zone where the record is created.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<zone-domain>` | DNS zone domain (e.g., `example.com`) | `DigitalOceanDnsZone` status or domain name |
| `<target-hostname>` | Target hostname (FQDN) for the alias | Your origin, CDN, or third-party service |
| `www` | Subdomain name | Your desired subdomain (cannot be `@` for CNAME on DigitalOcean) |

## Related Presets

- **01-a-record** -- Use when target is an IP or for root domain (`@`)
