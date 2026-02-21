# CNAME Alias

This preset creates a DNS CNAME record that aliases one domain name to another. CNAME records are the standard mechanism for pointing subdomains at canonical hostnames, external services, or CDN endpoints without hardcoding IP addresses. When the target's IP changes, the CNAME automatically follows without any record updates.

## When to Use

- Aliasing `www.example.com` to `app.example.com` so both resolve to the same endpoint
- Pointing a subdomain at an external SaaS service (e.g., `status.example.com` to `statuspage.io`)
- Routing traffic through a CDN or WAF by CNAMEing to the provider's edge hostname
- Any scenario where a domain should resolve to the same address as another domain

## Key Configuration Choices

- **CNAME record type** (`rtype: CNAME`) -- aliases one domain to another. The DNS resolver follows the chain to the target's A/AAAA records. CNAME records cannot coexist with other record types at the same domain name (RFC 1034).
- **Trailing dot on rdata** (`app.example.com.`) -- the trailing dot makes the hostname fully qualified, preventing the DNS resolver from appending the zone name. OCI normalizes rdata and may add the trailing dot automatically, but including it explicitly avoids ambiguity.
- **300-second TTL** (`ttl: 300`) -- 5-minute cache lifetime. Appropriate for most production workloads. Use lower values during DNS migrations.
- **Single CNAME target** -- CNAME record sets always contain exactly one item. Multiple CNAME records for the same domain are invalid per DNS standards.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<zone-name-or-ocid>` | Name or OCID of the target DNS zone (e.g., `example.com` or zone OCID) | OCI Console > DNS Management > Zones, or `OciDnsZone` status outputs (`zoneId`) |

## Related Presets

- **01-a-record** -- use instead when you want to map a domain directly to an IPv4 address rather than aliasing to another domain
