---
title: "CNAME Record"
description: "This preset creates a DNS CNAME record that aliases one hostname to another. CNAME records are used when you want a subdomain to resolve to the same address as another domain without hardcoding an IP..."
type: "preset"
rank: "02"
presetSlug: "02-cname-record"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "gcp"
icon: "package"
order: 2
---

# CNAME Record

This preset creates a DNS CNAME record that aliases one hostname to another. CNAME records are used when you want a subdomain to resolve to the same address as another domain without hardcoding an IP address.

## When to Use

- Aliasing `www.example.com` to `example.com` or to a CDN/load balancer hostname
- Pointing subdomains to SaaS provider endpoints (e.g., `status.example.com` to `statuspage.io`)
- Any domain alias where the target IP may change and you want automatic following

## Key Configuration Choices

- **CNAME record type** -- aliases one hostname to another; resolvers follow the chain
- **Trailing dot on both name and value** -- required by Cloud DNS for FQDN
- **5-minute TTL** (`ttlSeconds: 300`) -- default from spec

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<dns-zone-name>` | Name of the Cloud DNS managed zone | `GcpDnsZone` status outputs |
| `<subdomain.example.com.>` | Source FQDN with trailing dot (e.g., `www.example.com.`) | Your DNS naming scheme |
| `<target.example.com.>` | Target FQDN with trailing dot (e.g., `cdn.example.com.`) | Your target service hostname |

## Related Presets

- **01-a-record** -- Use when pointing directly to an IP address instead of another hostname
