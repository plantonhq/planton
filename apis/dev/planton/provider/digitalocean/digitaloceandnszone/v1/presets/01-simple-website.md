# Simple Website Zone

This preset creates a DNS zone with a single A record pointing the root domain to your web server or load balancer IP. Minimal configuration for getting a domain live quickly. TTL is set to 1 hour for reasonable cache behavior.

## When to Use

- Simple website or single-page app with one backend IP
- Quick domain setup for demos or staging
- Root domain (apex) pointing directly to a Droplet or load balancer

## Key Configuration Choices

- **Root A record** (`name: "@"`, `type: A`) -- apex domain points to the specified IP.
- **TTL 3600** (`ttlSeconds: 3600`) -- 1-hour cache; balance between propagation speed and lookup latency.
- **Single value** (`values`) -- one IP; add more for round-robin if needed (same record type).
- **Domain name** (`domainName`) -- your registered domain; DigitalOcean will host DNS for it.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<your-domain.com>` | Your registered domain name | Domain registrar |
| `<web-server-ip>` | IPv4 address of Droplet, load balancer, or CDN | DigitalOcean dashboard or resource outputs |

## Related Presets

- **02-production-with-email** -- Use when you need MX and TXT records for email (SPF, etc.)
