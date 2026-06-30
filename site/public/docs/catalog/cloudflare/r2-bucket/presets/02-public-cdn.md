---
title: "Public R2 Bucket with Custom Domain"
description: "Creates a public R2 bucket served via a custom domain (e.g., media.example.com). Combines public access with a branded CDN URL. Requires a Cloudflare DNS zone for the domain."
type: "preset"
rank: "02"
presetSlug: "02-public-cdn"
componentSlug: "r2-bucket"
componentTitle: "R2 Bucket"
provider: "cloudflare"
icon: "package"
order: 2
---

# Public R2 Bucket with Custom Domain

Creates a public R2 bucket served via a custom domain (e.g., media.example.com). Combines public access with a branded CDN URL. Requires a Cloudflare DNS zone for the domain.

## When to Use

- Static assets (images, fonts, videos) for a web app
- Public media or file hosting with a custom domain
- CDN-backed object storage with your domain

## Key Configuration Choices

- **publicAccess: true** (`publicAccess: true`) -- Enables the managed r2.dev public URL.
- **customDomains** (`customDomains`) -- A list; each entry requires zoneId (value wrapper) and domain in that zone. A bucket may serve multiple custom domains.
- **minTls** (`customDomains[].minTls`) -- Minimum TLS version ("1.0"-"1.3"); "1.2" is a sensible secure default.
- **cors** (`cors.rules`) -- Allows browsers to fetch objects cross-origin; tighten `origins` from `"*"` to your app's origin in production.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<bucket-name>` | Unique bucket name | Choose DNS-safe name (e.g., media-cdn) |
| `<cloudflare-account-id>` | Cloudflare account ID | Dashboard → Overview → Account ID |
| `<cloudflare-zone-id>` | Zone ID for your domain | CloudflareDnsZone status.outputs.zone_id |
| `<cdn-subdomain>` | Subdomain for the bucket | e.g., media, cdn, static |
| `<your-domain.com>` | Your domain | Domain in the Cloudflare zone |

## Related Presets

- **01-private** -- Use when bucket should not be publicly accessible
