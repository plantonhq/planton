# Public R2 Bucket with Custom Domain

Creates a public R2 bucket served via a custom domain (e.g., media.example.com). Combines public access with a branded CDN URL. Requires a Cloudflare DNS zone for the domain.

## When to Use

- Static assets (images, fonts, videos) for a web app
- Public media or file hosting with a custom domain
- CDN-backed object storage with your domain

## Key Configuration Choices

- **publicAccess: true** (`publicAccess: true`) -- Enables public read access.
- **customDomain** (`customDomain`) -- Requires zoneId (value wrapper) and domain. Must be in the zone.
- **zoneId** (`customDomain.zoneId`) -- Cloudflare zone ID; use value wrapper or reference to CloudflareDnsZone.

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
