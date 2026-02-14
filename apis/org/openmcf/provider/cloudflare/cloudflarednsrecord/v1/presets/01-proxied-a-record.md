# Proxied A Record

Creates an A record with Cloudflare proxy (orange cloud) enabled. Traffic flows through Cloudflare's CDN and DDoS protection, hiding your origin IP. Use for web-facing hostnames where you want security and performance benefits.

## When to Use

- Web servers, APIs, or applications that benefit from Cloudflare proxy
- Hiding origin server IP from direct exposure
- DDoS protection and CDN caching for static/dynamic content

## Key Configuration Choices

- **Proxied** (`proxied: true`) -- Traffic goes through Cloudflare; use false for DNS-only (grey cloud).
- **TTL 1** (`ttl: 1`) -- Automatic TTL; recommended for proxied records.
- **zoneId** (`zoneId`) -- Cloudflare zone ID; use value wrapper or reference to CloudflareDnsZone.
- **type A** (`type: A`) -- IPv4; use AAAA for IPv6, CNAME for aliases.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID for the DNS zone | CloudflareDnsZone status.outputs.zone_id or Dashboard → Zone → Overview |
| `<origin-ip-address>` | IPv4 address of your origin server | Your web server or load balancer IP |
| `www` | Record name; use `@` for root, or subdomain | Desired hostname |

## Related Presets

- **02-mx-email** -- Use when configuring mail exchange records instead
