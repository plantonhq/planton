# A Record

This preset creates a standard A record that points a hostname to an IPv4 address. Use for the root domain or any subdomain that should resolve to a Droplet, load balancer, or other IP. TTL is set to 1 hour for a balance between cache efficiency and propagation speed.

## When to Use

- Root domain (`@`) or subdomain pointing to a web server
- Directing traffic to a Droplet IP or load balancer IP
- Any hostname that maps to a single IPv4 address

## Key Configuration Choices

- **Type A** (`type: A`) -- IPv4 address record; use AAAA for IPv6.
- **Root domain** (`name: "@"`) -- apex record; use `www`, `api`, etc. for subdomains.
- **TTL 3600** (`ttlSeconds: 3600`) -- 1-hour cache; lower for frequent changes, higher for stability.
- **Domain reference** (`domain`) -- the DNS zone (domain) where this record lives; use `DigitalOceanDnsZone` reference or literal domain.
- **Value** (`value`) -- IPv4 address or reference to `DigitalOceanDroplet`/`DigitalOceanLoadBalancer` IP output.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<zone-domain>` | DNS zone domain (e.g., `example.com`) | `DigitalOceanDnsZone` status or domain name |
| `<target-ip>` | IPv4 address for the record | Droplet/LB IP from DigitalOcean or resource outputs |
| `@` | Record name; use `@` for root, or subdomain like `www` | Your desired hostname |

## Related Presets

- **02-cname-record** -- Use when pointing to another hostname (alias) instead of IP
