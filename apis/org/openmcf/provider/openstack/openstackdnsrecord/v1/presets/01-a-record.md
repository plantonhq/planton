# A Record

This preset creates a standalone DNS A record in Designate, mapping a hostname to an IPv4 address. Use standalone records (instead of inline records in `OpenStackDnsZone`) when individual records need to be independently managed or visible as separate DAG nodes in InfraCharts.

## When to Use

- Pointing a hostname to an instance's floating IP or load balancer VIP
- Any DNS record that should be managed independently from the zone lifecycle
- InfraCharts where the DNS record depends on another resource's output (e.g., a floating IP address)

## Key Configuration Choices

- **A record** (`type: A`) -- maps hostname to IPv4 address
- **5-minute TTL** (`ttl: 300`) -- moderate caching; lower for frequently changing records, higher for stable records
- **Trailing dot on recordName** -- DNS convention for FQDNs (e.g., `app.example.com.`)
- **ForceNew** -- `zoneId`, `recordName`, `type`, and `region` are immutable

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<zone-id>` | ID of the DNS zone this record belongs to | OpenStack console or `OpenStackDnsZone` status outputs |
| `<hostname.your-domain.com.>` | Fully qualified record name with trailing dot | Your DNS naming convention |
| `<ip-address>` | IPv4 address to point to (e.g., `203.0.113.42`) | Instance, floating IP, or LB status outputs |

## Related Presets

- **02-cname-record** -- Use instead when aliasing one hostname to another
