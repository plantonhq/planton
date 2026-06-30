# A Record

This preset creates a DNS A record that maps a hostname to an IPv4 address. This is the most common DNS record type, used to point domain names at servers, load balancers, or other IP-based endpoints.

## When to Use

- Pointing a subdomain to a Scaleway Instance, Load Balancer, or Public Gateway IP
- Creating the primary DNS entry for a web application or API
- Any scenario requiring hostname-to-IP resolution

## Key Configuration Choices

- **A record type** (`type: A`) -- resolves a hostname to an IPv4 address
- **Subdomain name** (`name: app`) -- creates `app.<your-zone>` as the fully qualified name; use empty string for the zone apex
- **1-hour TTL** (`ttl: 3600`) -- standard caching duration; reduce to 300 during migrations, increase to 86400 for static records
- **Keep empty zone** (`keepEmptyZone: true`) -- prevents accidental zone deletion when this is the last record being destroyed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-dns-zone-name>` | DNS zone name (e.g., `example.com`) | Scaleway console or `ScalewayDnsZone` status outputs |
| `<your-server-ip>` | IPv4 address to point the record at (e.g., `51.159.26.100`) | Scaleway console, `ScalewayLoadBalancer` or `ScalewayInstance` status outputs |

## Related Presets

- **02-cname-record** -- Use instead when pointing a hostname to another hostname (e.g., `www` to a CDN or external service)
