# A Record

This preset creates a DNS A record mapping a subdomain to an IPv4 address within an Azure DNS Zone. A records are the most fundamental DNS record type, used to point domain names to the IP addresses of servers, load balancers, and other internet-facing resources.

## When to Use

- Pointing a subdomain (e.g., `api.example.com`) to a specific IPv4 address
- Creating root domain (`@`) or wildcard (`*`) A records
- Mapping domain names to Azure Public IP addresses, external servers, or CDN endpoints

## Key Configuration Choices

- **Record type** (`type: A`) -- Maps a name to an IPv4 address
- **TTL** (`ttlSeconds: 300`) -- 5-minute cache; balances DNS propagation speed with resolver load. Use 60s for records that change frequently, 3600s for stable records
- **Single value** -- Add additional IPs to the `values` list for round-robin DNS

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Resource group containing the DNS zone | Azure portal or `AzureResourceGroup` status outputs |
| `<your-domain.com>` | The DNS zone name | Azure portal or `AzureDnsZone` status outputs |
| `<subdomain>` | Record name: `@` for apex, `www` for subdomain, `*` for wildcard | Your DNS design |
| `<ipv4-address>` | Target IPv4 address (e.g., `203.0.113.10`) | Azure portal or `AzurePublicIp` / `AzureLoadBalancer` status outputs |

## Related Presets

- **02-cname-record** -- Use instead when pointing to another domain name (e.g., CDN endpoint, Traffic Manager)
