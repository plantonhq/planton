# Basic Zone

Creates a Cloudflare DNS zone with no inline records. Ideal when you want to
manage DNS records separately via CloudflareDnsRecord resources. Requires only
`zoneName` and `accountId`; the zone defaults to a full (Cloudflare-hosted) zone.

## When to Use

- Adding a new domain to Cloudflare for DNS management
- Bootstrapping a zone before creating individual DNS records as separate resources
- A clean zone you will layer records, DNS settings, or DNSSEC onto later

## Key Configuration Choices

- **type** (omitted) -- Defaults to `full` (DNS hosted entirely with Cloudflare).
- **No inline records** (`records` omitted) -- Add records via CloudflareDnsRecord resources.
- **zoneName** (`zoneName`) -- Fully qualified domain (e.g., example.com).
- **accountId** (`accountId`) -- Your Cloudflare account ID.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<your-domain.com>` | Fully qualified domain for the zone | Your registered domain |
| `<cloudflare-account-id>` | Cloudflare account ID | Cloudflare Dashboard → Overview → Account ID (right sidebar) |

## Related Presets

- **02-dnssec-signed** -- A zone with DNSSEC enabled and the DS material exported
