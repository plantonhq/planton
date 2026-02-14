# A Record

This preset creates a standalone A record pointing a subdomain to an IPv4 address. This is the most common DNS record type, used for mapping hostnames to servers. Use standalone DNS records when adding records to an existing zone managed separately.

## When to Use

- Adding a subdomain (e.g., `api.example.com`) to an existing DNS zone
- Pointing a hostname to a specific server IP
- Any scenario where the zone already exists and you need to add individual records

## Key Configuration Choices

- **A record type** (`type: A`) -- maps a hostname to an IPv4 address
- **Subdomain** (`name: api`) -- the hostname relative to the zone (e.g., `api` creates `api.example.com`)
- **1-hour TTL** (`ttl: 3600`) -- standard caching duration; lower for records that change frequently
- **Zone reference** (`zoneId`) -- ties this record to an existing CivoDnsZone

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<dns-zone-id>` | ID of the target CivoDnsZone | `CivoDnsZone` status outputs |
| `api` | Subdomain name (relative to zone) | Your DNS naming plan |
| `<server-ipv4-address>` | Target IPv4 address | `CivoComputeInstance` or `CivoIpAddress` status outputs |

## Related Presets

- **02-cname-record** -- Use instead when the target is another hostname rather than an IP address
