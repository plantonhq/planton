---
title: "Simple Website DNS Zone"
description: "This preset creates a DNS zone with an apex A record pointing to a server IP and a `www` CNAME aliasing to the apex. This is the most common DNS configuration for a website: visitors reach the site..."
type: "preset"
rank: "01"
presetSlug: "01-simple-website"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "civo"
icon: "package"
order: 1
---

# Simple Website DNS Zone

This preset creates a DNS zone with an apex A record pointing to a server IP and a `www` CNAME aliasing to the apex. This is the most common DNS configuration for a website: visitors reach the site via both `example.com` and `www.example.com`.

## When to Use

- Any website or web application that needs DNS on Civo
- Standard domain setup where both apex and `www` should resolve to the same server
- Simple hosting scenarios without email or complex routing

## Key Configuration Choices

- **Inline records** -- zone and records created together as a single unit; Civo's DNS model supports this as the primary pattern
- **Apex A record** (`name: "@"`) -- points the root domain to a server IP
- **www CNAME** (`name: www`) -- aliases `www.example.com` to the apex, so changes only need to be made in one place
- **1-hour TTL** (`ttlSeconds: 3600`) -- recommended default; balance between DNS propagation speed and resolver caching

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `example.com` | Your domain name (FQDN) | Your domain registrar |
| `<server-ipv4-address>` | Public IP of your web server | `CivoComputeInstance` or `CivoIpAddress` status outputs |

## Related Presets

- **02-with-email** -- Use instead when the domain also needs MX and SPF records for email delivery
