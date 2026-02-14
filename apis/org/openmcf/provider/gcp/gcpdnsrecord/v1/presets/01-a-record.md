# A Record

This preset creates a standard DNS A record pointing a domain name to an IPv4 address. This is the most common DNS record type, used for mapping hostnames to IP addresses.

## When to Use

- Pointing a domain or subdomain to a server's IPv4 address
- Load balancer IP address mapping
- Any hostname-to-IP resolution

## Key Configuration Choices

- **A record type** -- maps hostname to IPv4 address
- **5-minute TTL** (`ttlSeconds: 300`) -- default from spec; balance between responsiveness and DNS cache efficiency
- **Trailing dot on name** -- required by Cloud DNS to indicate FQDN (e.g., `www.example.com.`)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<dns-zone-name>` | Name of the Cloud DNS managed zone | `GcpDnsZone` status outputs |
| `<subdomain.example.com.>` | FQDN with trailing dot (e.g., `api.example.com.`) | Your DNS naming scheme |
| `<ipv4-address>` | Target IPv4 address (e.g., `34.120.0.1`) | Load balancer or VM external IP |

## Related Presets

- **02-cname-record** -- Use for aliases to other hostnames instead of IP addresses
