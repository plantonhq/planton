---
title: "Free Plan Zone"
description: "Creates a Cloudflare DNS zone with no inline records, using the free plan. Ideal when you want to manage DNS records separately via CloudflareDnsRecord resources. Requires zone_name and account_id."
type: "preset"
rank: "01"
presetSlug: "01-free-plan"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "cloudflare"
icon: "package"
order: 1
---

# Free Plan Zone

Creates a Cloudflare DNS zone with no inline records, using the free plan. Ideal when you want to manage DNS records separately via CloudflareDnsRecord resources. Requires zone_name and account_id.

## When to Use

- Adding a new domain to Cloudflare for DNS management
- Bootstrap a zone before creating individual DNS records as separate resources
- Free tier with basic DNS, CDN, and DDoS protection when records are proxied

## Key Configuration Choices

- **Plan FREE** (`plan: FREE`) -- Default free tier; use PRO, BUSINESS, or ENTERPRISE for advanced features.
- **No inline records** (`records` omitted) -- Zone-only; add records via CloudflareDnsRecord resources.
- **zoneName** (`zoneName`) -- Fully qualified domain (e.g., example.com).
- **accountId** (`accountId`) -- Your Cloudflare account ID.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<your-domain.com>` | Fully qualified domain for the zone | Your registered domain |
| `<cloudflare-account-id>` | Cloudflare account ID | Cloudflare Dashboard → Overview → Account ID (right sidebar) |
