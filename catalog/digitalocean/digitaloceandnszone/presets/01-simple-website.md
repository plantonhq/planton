# Simple Website Zone

This preset creates a DNS zone with the two records a basic website needs: the apex pointing at the web server's IPv4 address and `www` following the apex. The zone starts serving on DigitalOcean's name servers immediately; the domain resolves publicly once the registrar delegates to `ns1`/`ns2`/`ns3.digitalocean.com` (the zone's `nameservers` output).

## When to Use

- A single-server website or landing page
- The starting point for any new domain hosted on DigitalOcean DNS

## Key Configuration Choices

- **Apex A record** (`name: "@"`) — the bare domain resolves to the server. The value can also reference another resource's output (e.g. a Droplet's `ipv4_address`) for chart composition.
- **`www` as CNAME to the apex** — one address to maintain; `www` follows automatically. The trailing dot on the target matches how DigitalOcean stores it.
- **One-hour TTLs** — a sensible production default; omit `ttlSeconds` to take DigitalOcean's 1800-second default.

## Placeholders to Replace

- `metadata.name` — your zone resource's name.
- `domainName` (`example.com` is a documentation example) — your domain.
- The A record's `values` (`203.0.113.10` is a documentation example) — your server's public IPv4 address.

## Related Presets

- **02-production-with-email** — adds mail routing (MX), SPF, and certificate authority pinning (CAA).
